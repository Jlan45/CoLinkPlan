package adapter

import (
	"context"
)

// ProviderAdapter defines the interface for different AI providers
type ProviderAdapter interface {
	// Name returns the provider name (e.g., openai, claude)
	Name() string
	// Call initiates a streaming request to the upstream API.
	// It pushes decoded chunks to the streamCh and errors to errCh.
	// It closes streamCh when the response is fully read.
	Call(ctx context.Context, requestID string, model string, reqBody []byte, streamCh chan<- interface{}, errCh chan<- error)
}
