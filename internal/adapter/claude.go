package adapter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"CoLinkPlan/internal/protocol"
	"CoLinkPlan/pkg/logger"
)

type ClaudeAdapter struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func NewClaudeAdapter(apiKey, baseURL string) *ClaudeAdapter {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	return &ClaudeAdapter{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (a *ClaudeAdapter) Name() string {
	return "claude"
}

// claudeRequest represents the basic structure of a Messages API request
type claudeRequest struct {
	Model       string          `json:"model"`
	System      string          `json:"system,omitempty"`
	Messages    []claudeMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Stream      bool            `json:"stream,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (a *ClaudeAdapter) Call(ctx context.Context, requestID string, model string, reqBody []byte, streamCh chan<- interface{}, errCh chan<- error) {
	defer close(streamCh)
	defer close(errCh)

	var req protocol.ChatCompletionRequest
	if err := json.Unmarshal(reqBody, &req); err != nil {
		errCh <- fmt.Errorf("failed to parse request: %w", err)
		return
	}

	// Convert OpenAI request to Claude request
	cReq := claudeRequest{
		Model:       model,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
		Temperature: req.Temperature,
	}

	if cReq.MaxTokens == 0 {
		cReq.MaxTokens = 4096 // Claude requires max_tokens
	}

	for _, m := range req.Messages {
		contentStr := ""
		if s, ok := m.Content.(string); ok {
			contentStr = s
		} else {
			// If it's an array of objects (like images), we convert to string or stringify it for now
			b, _ := json.Marshal(m.Content)
			contentStr = string(b)
		}

		if m.Role == "system" {
			// Claude expects system prompt at root level
			cReq.System = contentStr
		} else {
			cReq.Messages = append(cReq.Messages, claudeMessage{
				Role:    m.Role,
				Content: contentStr,
			})
		}
	}

	reqBody, err := json.Marshal(cReq)
	if err != nil {
		errCh <- fmt.Errorf("failed to marshal claude request: %w", err)
		return
	}

	url := fmt.Sprintf("%s/messages", strings.TrimSuffix(a.BaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		errCh <- fmt.Errorf("failed to create http request: %w", err)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	logger.Log.Info("Sending request to Claude", "request_id", requestID, "url", url, "model", model)

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		errCh <- fmt.Errorf("http request failed: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errCh <- fmt.Errorf("claude api returned status %d", resp.StatusCode)
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	var eventType string
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Handle various Claude streaming events to map to OpenAI format
			if eventType == "content_block_delta" {
				var delta map[string]interface{}
				if err := json.Unmarshal([]byte(data), &delta); err != nil {
					continue
				}

				if d, ok := delta["delta"].(map[string]interface{}); ok {
					if dType, ok := d["type"].(string); ok && dType == "text_delta" {
						if text, ok := d["text"].(string); ok {
							// Map text to OpenAI Choice schema
							openAIOBJ := map[string]interface{}{
								"id":      requestID,
								"object":  "chat.completion.chunk",
								"created": time.Now().Unix(),
								"model":   model,
								"choices": []map[string]interface{}{
									{
										"index": 0,
										"delta": map[string]interface{}{
											"content": text,
										},
									},
								},
							}

							select {
							case streamCh <- openAIOBJ:
							case <-ctx.Done():
								return
							}
						}
					}
				}
			} else if eventType == "message_stop" {
				return // End of stream
			} else if eventType == "error" {
				logger.Log.Error("Claude stream returned error format", "data", data)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		errCh <- fmt.Errorf("error reading stream: %w", err)
	}
}
