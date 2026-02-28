package protocol

// ChatCompletionRequest represents a standard OpenAI API request body
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents a standard OpenAI API response body (non-stream)
type ChatCompletionResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   UsageStat `json:"usage"`
}

type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"` // Used in non-stream
	Delta        *Message `json:"delta,omitempty"`   // Used in stream
	FinishReason string   `json:"finish_reason,omitempty"`
}

type UsageStat struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionStreamResponse represents a chunk of steam response
type ChatCompletionStreamResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}
