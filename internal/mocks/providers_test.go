package mocks

import (
	"context"
	"errors"
	"testing"
)

func TestMockSessionProvider(t *testing.T) {
	provider := NewMockSessionProvider()
	ctx := context.Background()

	// Test getting existing session
	session, err := provider.GetSession(ctx, "sess_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}
	if session.SessionID != "sess_123" {
		t.Errorf("expected session_id 'sess_123', got %q", session.SessionID)
	}
	if session.ParticipantID != "user_001" {
		t.Errorf("expected participant_id 'user_001', got %q", session.ParticipantID)
	}

	// Test getting non-existent session
	session, err = provider.GetSession(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session != nil {
		t.Error("expected nil session for nonexistent ID")
	}

	// Test expired session
	session, err = provider.GetSession(ctx, "sess_expired")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("expected expired session, got nil")
	}

	// Test revoked session
	session, err = provider.GetSession(ctx, "sess_revoked")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil || !session.Revoked {
		t.Error("expected revoked session")
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
	// Test session provider mutations
	sessionProvider := NewMockSessionProvider()
	ctx := context.Background()

	session, _ := sessionProvider.GetSession(ctx, "sess_123")
	// Mutate the returned session
	session.Revoked = true

	// Original should be unchanged
	original, _ := sessionProvider.GetSession(ctx, "sess_123")
	if original.Revoked {
		t.Error("mutation affected original session")
	}

	// Test category index provider mutations
	categoryProvider := NewMockCategoryIndexProvider()
	index, _ := categoryProvider.GetCategoryIndex(ctx, "user_001")
	index.Categories["NEW"] = "new_cid"

	originalIndex, _ := categoryProvider.GetCategoryIndex(ctx, "user_001")
	if _, ok := originalIndex.Categories["NEW"]; ok {
		t.Error("mutation affected original index")
	}
}
