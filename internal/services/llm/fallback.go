package llm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/access-bot/internal/providers"
)

// FallbackLLMProvider tries a primary LLM and falls back to a secondary on any error.
type FallbackLLMProvider struct {
	primary   providers.LLMProvider
	secondary providers.LLMProvider
}

// Ensure FallbackLLMProvider implements LLMProvider.
var _ providers.LLMProvider = (*FallbackLLMProvider)(nil)

// NewFallbackLLMProvider creates a provider that tries primary first, then secondary.
func NewFallbackLLMProvider(primary, secondary providers.LLMProvider) *FallbackLLMProvider {
	return &FallbackLLMProvider{primary: primary, secondary: secondary}
}

// Complete calls the primary provider; on any error falls back to the secondary.
func (f *FallbackLLMProvider) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	result, err := f.primary.Complete(ctx, systemPrompt, userPrompt)
	if err == nil {
		return result, nil
	}

	slog.Warn("primary LLM failed, falling back",
		"primary", f.primary.GetModelName(),
		"fallback", f.secondary.GetModelName(),
		"error", err.Error(),
	)

	result, err2 := f.secondary.Complete(ctx, systemPrompt, userPrompt)
	if err2 != nil {
		return "", fmt.Errorf("primary (%s): %w; fallback (%s): %v",
			f.primary.GetModelName(), err, f.secondary.GetModelName(), err2)
	}
	return result, nil
}

// GetModelName returns both model names to indicate the fallback chain.
func (f *FallbackLLMProvider) GetModelName() string {
	return fmt.Sprintf("%s→%s", f.primary.GetModelName(), f.secondary.GetModelName())
}
