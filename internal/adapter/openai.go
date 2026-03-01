package adapter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"CoLinkPlan/pkg/logger"
)

type OpenAIAdapter struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func NewOpenAIAdapter(apiKey, baseURL string) *OpenAIAdapter {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIAdapter{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (a *OpenAIAdapter) Name() string {
	return "openai"
}

func (a *OpenAIAdapter) Call(ctx context.Context, requestID string, model string, reqBody []byte, streamCh chan<- interface{}, errCh chan<- error) {
	defer close(streamCh)
	defer close(errCh)

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(reqBody, &payloadMap); err != nil {
		errCh <- fmt.Errorf("failed to unmarshal request: %w", err)
		return
	}

	payloadMap["model"] = model
	isStream := false
	if s, ok := payloadMap["stream"].(bool); ok && s {
		isStream = true
	}

	modifiedBody, err := json.Marshal(payloadMap)
	if err != nil {
		errCh <- fmt.Errorf("failed to marshal openai request: %w", err)
		return
	}

	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(a.BaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(modifiedBody))
	if err != nil {
		errCh <- fmt.Errorf("failed to create http request: %w", err)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	logger.Log.Info("Sending request to OpenAI", "request_id", requestID, "url", url, "model", model, "stream", isStream)

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		errCh <- fmt.Errorf("http request failed: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errCh <- fmt.Errorf("openai api returned status %d", resp.StatusCode)
		return
	}

	if !isStream {
		var res map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			errCh <- fmt.Errorf("error reading non-stream response: %w", err)
			return
		}
		select {
		case streamCh <- res:
		case <-ctx.Done():
		}
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			// We just read it as raw interface{} or map to push down the stream.
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				logger.Log.Warn("Failed to unmarshal openai chunk", "request_id", requestID, "err", err, "data", data)
				continue
			}

			select {
			case streamCh <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		errCh <- fmt.Errorf("error reading stream: %w", err)
	}
}
