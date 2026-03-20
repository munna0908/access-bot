package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/access-bot/internal/models"
)

func TestMockSessionProvider_ValidateSession(t *testing.T) {
	provider := NewMockSessionProvider()
	ctx := context.Background()

	// Test validating existing session with correct params
	req := &models.ValidateSessionRequest{
		ParticipantID:  "user_001",
		AgentID:        "openclaw_whatsapp_bot",
		SessionID:      "sess_123",
		RequiredScopes: []string{"preferences.food.read"},
		CurrentTime:    1700000000, // Before expiry (1900000000)
	}
	err := provider.ValidateSession(ctx, req)
	if err != nil {
		t.Fatalf("expected valid session, got error: %v", err)
	}

	// Test non-existent session
	req = &models.ValidateSessionRequest{
		ParticipantID:  "user_001",
		AgentID:        "openclaw_whatsapp_bot",
		SessionID:      "nonexistent",
		RequiredScopes: []string{},
		CurrentTime:    1700000000,
	}
	err = provider.ValidateSession(ctx, req)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}

	// Test expired session
	req = &models.ValidateSessionRequest{
		ParticipantID:  "user_001",
		AgentID:        "openclaw_whatsapp_bot",
		SessionID:      "sess_expired",
		RequiredScopes: []string{},
		CurrentTime:    1700000000,
	}
	err = provider.ValidateSession(ctx, req)
	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}

	// Test revoked session
	req = &models.ValidateSessionRequest{
		ParticipantID:  "user_001",
		AgentID:        "openclaw_whatsapp_bot",
		SessionID:      "sess_revoked",
		RequiredScopes: []string{},
		CurrentTime:    1700000000,
	}
	err = provider.ValidateSession(ctx, req)
	if !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("expected ErrSessionRevoked, got %v", err)
	}

	// Test exhausted session
	req = &models.ValidateSessionRequest{
		ParticipantID:  "user_001",
		AgentID:        "openclaw_whatsapp_bot",
		SessionID:      "sess_exhausted",
		RequiredScopes: []string{},
		CurrentTime:    1700000000,
	}
	err = provider.ValidateSession(ctx, req)
	if !errors.Is(err, ErrSessionExhausted) {
		t.Errorf("expected ErrSessionExhausted, got %v", err)
	}

	// Test participant mismatch
	req = &models.ValidateSessionRequest{
		ParticipantID:  "wrong_user",
		AgentID:        "openclaw_whatsapp_bot",
		SessionID:      "sess_123",
		RequiredScopes: []string{},
		CurrentTime:    1700000000,
	}
	err = provider.ValidateSession(ctx, req)
	if !errors.Is(err, ErrSessionParticipantMismatch) {
		t.Errorf("expected ErrSessionParticipantMismatch, got %v", err)
	}

	// Test agent mismatch
	req = &models.ValidateSessionRequest{
		ParticipantID:  "user_001",
		AgentID:        "wrong_agent",
		SessionID:      "sess_123",
		RequiredScopes: []string{},
		CurrentTime:    1700000000,
	}
	err = provider.ValidateSession(ctx, req)
	if !errors.Is(err, ErrSessionAgentMismatch) {
		t.Errorf("expected ErrSessionAgentMismatch, got %v", err)
	}

	// Test missing scope
	req = &models.ValidateSessionRequest{
		ParticipantID:  "user_001",
		AgentID:        "openclaw_whatsapp_bot",
		SessionID:      "sess_123",
		RequiredScopes: []string{"profile.address.read"}, // Not in sess_123
		CurrentTime:    1700000000,
	}
	err = provider.ValidateSession(ctx, req)
	if !errors.Is(err, ErrSessionMissingScope) {
		t.Errorf("expected ErrSessionMissingScope, got %v", err)
	}
}

func TestMockCategoryIndexProvider(t *testing.T) {
	provider := NewMockCategoryIndexProvider()
	ctx := context.Background()

	// Test getting existing index
	index, err := provider.GetCategoryIndex(ctx, "user_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}
	if index.ParticipantID != "user_001" {
		t.Errorf("expected participant_id 'user_001', got %q", index.ParticipantID)
	}
	if len(index.Categories) != 4 {
		t.Errorf("expected 4 categories, got %d", len(index.Categories))
	}

	// Test getting non-existent index
	index, err = provider.GetCategoryIndex(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if index != nil {
		t.Error("expected nil index for nonexistent participant")
	}
}

func TestMockFilesystemProvider(t *testing.T) {
	provider := NewMockFilesystemProvider()
	ctx := context.Background()

	// Test getting existing file
	file, err := provider.FetchByCID(ctx, "bafy_food_cid_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file == nil {
		t.Fatal("expected file, got nil")
	}
	if file.CID != "bafy_food_cid_001" {
		t.Errorf("expected CID 'bafy_food_cid_001', got %q", file.CID)
	}
	if file.ContentType != "text/markdown" {
		t.Errorf("expected content_type 'text/markdown', got %q", file.ContentType)
	}
	if file.Content == "" {
		t.Error("expected non-empty content")
	}

	// Test getting non-existent file
	file, err = provider.FetchByCID(ctx, "nonexistent_cid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file != nil {
		t.Error("expected nil file for nonexistent CID")
	}
}

func TestMockLLMProvider(t *testing.T) {
	provider := NewMockLLMProvider()
	ctx := context.Background()

	// Test default response (contextual)
	response, err := provider.Complete(ctx, "system prompt", "User Question:\nOrder some dinner\n\n## FOOD\n# Food Preferences")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response == "" {
		t.Error("expected non-empty response")
	}

	// Test setting default response
	provider.SetDefaultResponse("Custom response")
	response, err = provider.Complete(ctx, "system", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "Custom response" {
		t.Errorf("expected 'Custom response', got %q", response)
	}

	// Test failure mode
	testErr := errors.New("API error")
	provider.SetShouldFail(true, testErr)
	_, err = provider.Complete(ctx, "system", "user")
	if err == nil {
		t.Error("expected error")
	}
	if err != testErr {
		t.Errorf("expected %v, got %v", testErr, err)
	}

	// Test model name
	if provider.GetModelName() != "claude-mock" {
		t.Errorf("expected 'claude-mock', got %q", provider.GetModelName())
	}

	provider.SetModelName("custom-model")
	if provider.GetModelName() != "custom-model" {
		t.Errorf("expected 'custom-model', got %q", provider.GetModelName())
	}
}

func TestMockProviderMutations(t *testing.T) {
	ctx := context.Background()

	// Test category index provider mutations
	categoryProvider := NewMockCategoryIndexProvider()
	index, _ := categoryProvider.GetCategoryIndex(ctx, "user_001")
	index.Categories["NEW"] = "new_cid"

	originalIndex, _ := categoryProvider.GetCategoryIndex(ctx, "user_001")
	if _, ok := originalIndex.Categories["NEW"]; ok {
		t.Error("mutation affected original index")
	}
}
