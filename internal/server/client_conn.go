package server

import (
	"encoding/json"
	"sync"
	"time"

	"CoLinkPlan/internal/protocol"
	"CoLinkPlan/pkg/logger"

	"github.com/gorilla/websocket"
)

// ClientConn wraps a connected Gateway Client
type ClientConn struct {
	ID              string
	Conn            *websocket.Conn
	ConnMutex       sync.Mutex
	Hub             *Hub
	MaxParallel     int
	ActiveTasks     int
	SupportedModels map[string]bool

	// Penalized until this time
	PenaltyUntil time.Time

	// Pending streams mapped by RequestID
	PendingStreams map[string]chan protocol.WSPayload
	PendingMutex   sync.RWMutex

	closeCh chan struct{}
}

func NewClientConn(hub *Hub, conn *websocket.Conn, id string) *ClientConn {
	return &ClientConn{
		ID:              id,
		Conn:            conn,
		Hub:             hub,
		SupportedModels: make(map[string]bool),
		PendingStreams:  make(map[string]chan protocol.WSPayload),
		closeCh:         make(chan struct{}),
	}
}

func (c *ClientConn) SendMessage(payload protocol.WSPayload) error {
	c.ConnMutex.Lock()
	defer c.ConnMutex.Unlock()
	return c.Conn.WriteJSON(payload)
}

func (c *ClientConn) ReadLoop() {
	defer func() {
		c.Hub.unregister <- c
		c.ConnMutex.Lock()
		c.Conn.Close()
		c.ConnMutex.Unlock()
		close(c.closeCh)
	}()

	c.Conn.SetReadDeadline(time.Now().Add(15 * time.Second * 2)) // Expect pong
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(15 * time.Second * 2))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			logger.Log.Error("Read error from client", "client_id", c.ID, "err", err)
			break
		}

		var payload protocol.WSPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			logger.Log.Error("Invalid WS payload", "client_id", c.ID, "err", err)
			continue
		}

		switch payload.Type {
		case protocol.MsgTypeRegister:
			dataBytes, _ := json.Marshal(payload.Data)
			var reg protocol.RegisterData
			json.Unmarshal(dataBytes, &reg)

			c.Hub.mu.Lock()
			c.MaxParallel = reg.MaxParallel
			for _, m := range reg.Models {
				c.SupportedModels[m] = true
			}
			logger.Log.Info("Client registered", "client_id", c.ID, "max_parallel", c.MaxParallel, "models", reg.Models)
			c.Hub.mu.Unlock()

		case protocol.MsgTypeStream, protocol.MsgTypeFinish, protocol.MsgTypeError:
			dataBytes, _ := json.Marshal(payload.Data)

			var reqID string

			// Just quickly peak for RequestID using a generic map to route
			var generic map[string]interface{}
			json.Unmarshal(dataBytes, &generic)

			if idVal, ok := generic["request_id"].(string); ok {
				reqID = idVal
			}

			if reqID != "" {
				c.PendingMutex.RLock()
				streamCh, ok := c.PendingStreams[reqID]
				c.PendingMutex.RUnlock()

				if ok {
					streamCh <- payload

					// Release the parallel slot when the request is done (Finish or Error)
					if payload.Type == protocol.MsgTypeFinish || payload.Type == protocol.MsgTypeError {
						c.Hub.CompleteTask(c, reqID)
					}
				} else {
					logger.Log.Warn("Received message for unknown stream", "request_id", reqID, "client_id", c.ID)
				}
			}
		}
	}
}
