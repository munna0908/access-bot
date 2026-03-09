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

	if strings.Contains(promptLower, "dinner") || strings.Contains(promptLower, "food") || strings.Contains(promptLower, "order") {
		if strings.Contains(userPrompt, "FOOD") && strings.Contains(userPrompt, "HEALTH") {
			return "A vegetarian South Indian dinner would suit your preferences tonight. Avoid dairy-based dishes given your lactose intolerance.", nil
		}
		if strings.Contains(userPrompt, "FOOD") {
			return "Based on your preferences, I'd recommend a vegetarian South Indian restaurant. Masala Dosa would be a great choice.", nil
		}
	}

	if strings.Contains(promptLower, "address") || strings.Contains(promptLower, "delivery") {
		if strings.Contains(userPrompt, "ADDRESS") {
			return "Your delivery address is 123 Main Street, Apartment 4B, San Francisco, CA 94102. Gate code is 1234.", nil
		}
	}

	if strings.Contains(promptLower, "health") || strings.Contains(promptLower, "medical") {
		if strings.Contains(userPrompt, "HEALTH") {
			return "Based on your health profile, you should avoid high sodium foods and high glycemic index items due to your Type 2 Diabetes and mild hypertension.", nil
		}
	}

	if strings.Contains(promptLower, "payment") || strings.Contains(promptLower, "pay") || strings.Contains(promptLower, "budget") {
		if strings.Contains(userPrompt, "PAYMENT") {
			return "Your preferred payment method is Apple Pay with a typical meal budget of $15-25.", nil
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
