package mocks

import (
	"context"
	"strings"

	"github.com/access-bot/internal/providers"
)

// MockLLMProvider is an in-memory mock implementation of LLMProvider.
type MockLLMProvider struct {
	modelName          string
	defaultResponse    string
	hasDefaultResponse bool
	shouldFail         bool
	failError          error
}

// Ensure MockLLMProvider implements LLMProvider.
var _ providers.LLMProvider = (*MockLLMProvider)(nil)

// NewMockLLMProvider creates a new mock LLM provider.
func NewMockLLMProvider() *MockLLMProvider {
	return &MockLLMProvider{
		modelName:       "claude-mock",
		defaultResponse: "",
		shouldFail:      false,
	}
}

// Complete returns a mock response based on the prompt content.
func (m *MockLLMProvider) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if m.shouldFail {
		return "", m.failError
	}

	if m.hasDefaultResponse {
		return m.defaultResponse, nil
	}

	// Generate contextual mock responses based on prompt content
	promptLower := strings.ToLower(userPrompt)

	// Food ordering → numbered list (matches what the telegram bot parses)
	if strings.Contains(promptLower, "list exactly 3") || strings.Contains(promptLower, "numbered list") ||
		(strings.Contains(promptLower, "order food") || strings.Contains(promptLower, "suggest") && strings.Contains(promptLower, "food")) {
		return "1. Masala Dosa\n2. Pad Thai (no egg, no peanuts)\n3. Vegetable Biryani", nil
	}

	if strings.Contains(promptLower, "dinner") || strings.Contains(promptLower, "food") || strings.Contains(promptLower, "order") {
		if strings.Contains(userPrompt, "FOOD") {
			return "1. Masala Dosa\n2. Pad Thai (no egg, no peanuts)\n3. Vegetable Biryani", nil
		}
	}

	// Delivery address query
	if strings.Contains(promptLower, "delivery address") || strings.Contains(promptLower, "my address") {
		return "123 Palm Grove Apartments, Flat 4B, Tower 2, Koramangala, Bangalore – 560034. Delivery between 10 AM – 8 PM. Gate code: 1234.", nil
	}

	if strings.Contains(promptLower, "address") || strings.Contains(promptLower, "delivery") {
		if strings.Contains(userPrompt, "ADDRESS") {
			return "123 Palm Grove Apartments, Flat 4B, Tower 2, Koramangala, Bangalore – 560034. Gate code: 1234.", nil
		}
	}

	if strings.Contains(promptLower, "health") || strings.Contains(promptLower, "medical") {
		if strings.Contains(userPrompt, "HEALTH") {
			return "Based on your health profile, avoid high sodium and high glycemic index foods due to Type 2 Diabetes and mild hypertension. Also avoid peanuts and dairy.", nil
		}
	}

	if strings.Contains(promptLower, "schedule") || strings.Contains(promptLower, "meeting") || strings.Contains(promptLower, "calendar") {
		if strings.Contains(userPrompt, "SCHEDULE") {
			return "You have a standup at 10 AM and a product review at 3 PM today.", nil
		}
	}

	// Default response when context is insufficient
	return "Sorry, I can't answer that.", nil
}

// GetModelName returns the mock model name.
func (m *MockLLMProvider) GetModelName() string {
	return m.modelName
}

// SetDefaultResponse sets a fixed response for all calls (for testing).
func (m *MockLLMProvider) SetDefaultResponse(response string) {
	m.defaultResponse = response
	m.hasDefaultResponse = true
}

// ClearDefaultResponse clears any set default response (for testing).
func (m *MockLLMProvider) ClearDefaultResponse() {
	m.defaultResponse = ""
	m.hasDefaultResponse = false
}

// SetShouldFail configures the mock to return an error (for testing).
func (m *MockLLMProvider) SetShouldFail(shouldFail bool, err error) {
	m.shouldFail = shouldFail
	m.failError = err
}

// SetModelName sets the model name returned by GetModelName (for testing).
func (m *MockLLMProvider) SetModelName(name string) {
	m.modelName = name
}
