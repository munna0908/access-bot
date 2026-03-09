package providers

import (
	"context"
)

// LLMProvider defines the interface for LLM interactions.
type LLMProvider interface {
	// Complete sends a prompt to the LLM and returns the response.
	Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error)

	// GetModelName returns the name of the LLM model being used.
	GetModelName() string
}
