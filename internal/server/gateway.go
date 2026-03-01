package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"CoLinkPlan/internal/db"
	"CoLinkPlan/internal/limiter"
	"CoLinkPlan/internal/protocol"
	"CoLinkPlan/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Gateway struct {
	Hub     *Hub
	DB      *db.DB
	Limiter *limiter.RateLimiter
}

func NewGateway(hub *Hub, database *db.DB, rl *limiter.RateLimiter) *Gateway {
	return &Gateway{
		Hub:     hub,
		DB:      database,
		Limiter: rl,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (g *Gateway) WsHandler(c *gin.Context) {
	token := c.GetHeader("Client-Token")
	if token == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("Failed to upgrade to websocket", "err", err)
		return
	}

	client := NewClientConn(g.Hub, conn, token+"_"+uuid.New().String()[:8])
	g.Hub.register <- client

	go client.ReadLoop()
}

// ModelsHandler returns all model names currently available across connected nodes.
// GET /v1/models
// GET /v1/models/:model
func (g *Gateway) ModelsHandler(c *gin.Context) {
	now := time.Now().Unix()
	modelNames := g.Hub.ListModels()

	type ModelObject struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	// If a specific model is requested
	if id := c.Param("model"); id != "" {
		for _, m := range modelNames {
			if m == id {
				c.JSON(http.StatusOK, ModelObject{
					ID:      m,
					Object:  "model",
					Created: now,
					OwnedBy: "co-link",
				})
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"message": fmt.Sprintf("The model '%s' does not exist", id),
			"type":    "invalid_request_error",
			"code":    "model_not_found",
		}})
		return
	}

	data := make([]ModelObject, 0, len(modelNames))
	for _, m := range modelNames {
		data = append(data, ModelObject{
			ID:      m,
			Object:  "model",
			Created: now,
			OwnedBy: "co-link",
		})
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": data})
}

// authAndRateCheck validates the API key and enforces rate limits.
// Returns (keyRecord, true) on success, or writes an error JSON and returns (nil, false).
func (g *Gateway) authAndRateCheck(c *gin.Context) (*db.APIKeyRecord, bool) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
		return nil, false
	}
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")

	keyRecord, err := g.DB.GetAPIKey(c.Request.Context(), apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
		return nil, false
	}

	allowed, err := g.Limiter.Allow(c.Request.Context(), apiKey, keyRecord.RPM)
	if err != nil {
		logger.Log.Error("Rate limiter error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return nil, false
	}
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		return nil, false
	}
	return keyRecord, true
}

func (g *Gateway) ChatCompletionsHandler(c *gin.Context) {
	keyRecord, ok := g.authAndRateCheck(c)
	if !ok {
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

	var req protocol.ChatCompletionRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON mapping"})
		return
	}

	// Model allowed check
	modelAllowed := false
	for _, m := range keyRecord.AllowedModelList() {
		if m == "*" || m == req.Model {
			modelAllowed = true
			break
		}
	}
	if !modelAllowed {
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Model %s not allowed for this API Key", req.Model)})
		return
	}

	reqID := "req-" + uuid.New().String()

	var payload interface{}
	json.Unmarshal(bodyBytes, &payload)

	// We no longer force stream=true so the client adapter knows if it should proxy a stream or not

	// Dispatch and stream from hub
	streamCh, dispatchErr := g.dispatchWithRetry(c, reqID, req.Model, payload)
	if dispatchErr != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": dispatchErr.Error()})
		return
	}

	if req.Stream {
		g.handleStreamResponse(c, streamCh)
	} else {
		g.handleNonStreamResponse(c, req.Model, streamCh)
	}
}

// dispatchWithRetry attempts to route the call up to maxRetries times,
// returning the stream channel or an error.
func (g *Gateway) dispatchWithRetry(c *gin.Context, reqID, model string, payload interface{}) (chan protocol.WSPayload, error) {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		streamCh, err := g.Hub.RouteCall(c.Request.Context(), reqID, model, payload)
		if err != nil {
			logger.Log.Warn("Dispatch failed", "err", err, "attempt", i+1)
			continue
		}

		// Peek at first message to detect early errors
		firstMsg, ok := <-streamCh
		if !ok {
			continue
		}
		if firstMsg.Type == protocol.MsgTypeError {
			logger.Log.Warn("Client returned error on first message, retrying", "attempt", i+1)
			continue
		}

		// Rebuild a channel that includes the already-consumed firstMsg
		merged := make(chan protocol.WSPayload, 64)
		go func() {
			merged <- firstMsg
			for msg := range streamCh {
				merged <- msg
			}
			close(merged)
		}()
		return merged, nil
	}
	return nil, fmt.Errorf("no available clients after %d retries", maxRetries)
}

// handleStreamResponse pipes the hub stream directly to the HTTP client as SSE.
func (g *Gateway) handleStreamResponse(c *gin.Context, streamCh chan protocol.WSPayload) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case msg, ok := <-streamCh:
			if !ok {
				return
			}
			switch msg.Type {
			case protocol.MsgTypeFinish:
				c.Writer.Write([]byte("data: [DONE]\n\n"))
				c.Writer.Flush()
				return
			case protocol.MsgTypeError:
				writeSSEChunk(c.Writer, msg)
				c.Writer.Flush()
				return
			default:
				writeSSEChunk(c.Writer, msg)
				c.Writer.Flush()
			}
		}
	}
}

// handleNonStreamResponse collects the single non-stream response object from upstream
func (g *Gateway) handleNonStreamResponse(c *gin.Context, model string, streamCh chan protocol.WSPayload) {
	for {
		select {
		case <-c.Request.Context().Done():
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Client disconnected"})
			return
		case msg, ok := <-streamCh:
			if !ok {
				// Channel closed early
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Stream closed prematurely"})
				return
			}
			switch msg.Type {
			case protocol.MsgTypeFinish:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Stream finished before returning data"})
				return
			case protocol.MsgTypeError:
				dataBytes, _ := json.Marshal(msg.Data)
				var errData protocol.ErrorData
				json.Unmarshal(dataBytes, &errData)
				c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{
					"message": errData.Message,
					"type":    "upstream_error",
					"code":    errData.Code,
				}})
				return
			case protocol.MsgTypeStream:
				// For non-streaming requests, the very first chunk contains the entire JSON response from upstream.
				dataBytes, _ := json.Marshal(msg.Data)
				var sd protocol.StreamData
				if err := json.Unmarshal(dataBytes, &sd); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse provider response"})
					return
				}
				c.JSON(http.StatusOK, sd.Chunk)
				return // we are fully done after receiving the one response object
			}
		}
	}
}

func writeSSEChunk(w http.ResponseWriter, msg protocol.WSPayload) error {
	if msg.Type == protocol.MsgTypeError {
		dataBytes, _ := json.Marshal(msg.Data)
		w.Write([]byte(fmt.Sprintf("data: {\"error\": %s}\n\n", string(dataBytes))))
		return nil
	}

	dataBytes, _ := json.Marshal(msg.Data)
	var streamData protocol.StreamData
	json.Unmarshal(dataBytes, &streamData)

	chunkBytes, _ := json.Marshal(streamData.Chunk)
	_, err := w.Write([]byte(fmt.Sprintf("data: %s\n\n", string(chunkBytes))))
	return err
}
