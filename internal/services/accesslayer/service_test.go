package accesslayer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/access-bot/internal/domain"
	"github.com/access-bot/internal/logging"
	"github.com/access-bot/internal/mocks"
	"github.com/access-bot/internal/models"
)

func setupTestService() (*Service, *mocks.MockSessionProvider, *mocks.MockCategoryIndexProvider, *mocks.MockFilesystemProvider, *mocks.MockLLMProvider) {
	sessionProvider := mocks.NewMockSessionProvider()
	categoryIndexProvider := mocks.NewMockCategoryIndexProvider()
	filesystemProvider := mocks.NewMockFilesystemProvider()
	llmProvider := mocks.NewMockLLMProvider()
	logger := logging.NewLogger("error") // Suppress logs during tests

	service := NewService(
		sessionProvider,
		categoryIndexProvider,
		filesystemProvider,
		llmProvider,
		logger,
	)

	return service, sessionProvider, categoryIndexProvider, filesystemProvider, llmProvider
}

func TestProcessAnswerRequest_HappyPath(t *testing.T) {
	service, _, _, _, llmProvider := setupTestService()

	llmProvider.SetDefaultResponse("A vegetarian South Indian dinner would suit your preferences tonight.")

	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_123",
		Question:           "Order some dinner",
		RequiredCategories: []string{"FOOD", "HEALTH"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusOK) {
		t.Errorf("expected status %s, got %s (reason: %v)", domain.StatusOK, resp.Status, resp.Reason)
	}
	if resp.Answer == "" {
		t.Error("expected non-empty answer")
	}
	if len(resp.UsedCategories) != 2 {
		t.Errorf("expected 2 used categories, got %d", len(resp.UsedCategories))
	}
}

func TestProcessAnswerRequest_InvalidRequest(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	req := &models.AnswerRequest{
		// Missing required fields
		RequestID: "req_123",
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusInvalidRequest) {
		t.Errorf("expected status %s, got %s", domain.StatusInvalidRequest, resp.Status)
	}
	if resp.Answer != "Sorry, I can't answer that." {
		t.Errorf("expected safe failure message, got %q", resp.Answer)
	}
}

func TestProcessAnswerRequest_SessionNotFound(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "nonexistent_session",
		Question:           "Order dinner",
		RequiredCategories: []string{"FOOD"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusCannotAnswer) {
		t.Errorf("expected status %s, got %s", domain.StatusCannotAnswer, resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != string(domain.ReasonSessionNotFound) {
		t.Errorf("expected reason %s, got %v", domain.ReasonSessionNotFound, resp.Reason)
	}
}

func TestProcessAnswerRequest_SessionExpired(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_expired",
		Question:           "Order dinner",
		RequiredCategories: []string{"FOOD"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusPermissionDen) {
		t.Errorf("expected status %s, got %s", domain.StatusPermissionDen, resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != string(domain.ReasonSessionExpired) {
		t.Errorf("expected reason %s, got %v", domain.ReasonSessionExpired, resp.Reason)
	}
}

func TestProcessAnswerRequest_SessionRevoked(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_revoked",
		Question:           "Order dinner",
		RequiredCategories: []string{"FOOD"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusPermissionDen) {
		t.Errorf("expected status %s, got %s", domain.StatusPermissionDen, resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != string(domain.ReasonSessionRevoked) {
		t.Errorf("expected reason %s, got %v", domain.ReasonSessionRevoked, resp.Reason)
	}
}

func TestProcessAnswerRequest_UsageLimitExceeded(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_exhausted",
		Question:           "Order dinner",
		RequiredCategories: []string{"FOOD"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusPermissionDen) {
		t.Errorf("expected status %s, got %s", domain.StatusPermissionDen, resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != string(domain.ReasonUsageLimitExceeded) {
		t.Errorf("expected reason %s, got %v", domain.ReasonUsageLimitExceeded, resp.Reason)
	}
}

func TestProcessAnswerRequest_ParticipantMismatch(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "wrong_user",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_123",
		Question:           "Order dinner",
		RequiredCategories: []string{"FOOD"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusPermissionDen) {
		t.Errorf("expected status %s, got %s", domain.StatusPermissionDen, resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != string(domain.ReasonSessionParticipantMismatch) {
		t.Errorf("expected reason %s, got %v", domain.ReasonSessionParticipantMismatch, resp.Reason)
	}
}

func TestProcessAnswerRequest_AgentMismatch(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "user_001",
		AgentID:            "wrong_agent",
		SessionID:          "sess_123",
		Question:           "Order dinner",
		RequiredCategories: []string{"FOOD"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusPermissionDen) {
		t.Errorf("expected status %s, got %s", domain.StatusPermissionDen, resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != string(domain.ReasonSessionAgentMismatch) {
		t.Errorf("expected reason %s, got %v", domain.ReasonSessionAgentMismatch, resp.Reason)
	}
}

func TestProcessAnswerRequest_InsufficientScopes(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	// Request ADDRESS category which requires profile.address.read scope
	// but sess_123 only has preferences.food.read and health.read
	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_123",
		Question:           "What is my address?",
		RequiredCategories: []string{"ADDRESS"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusPermissionDen) {
		t.Errorf("expected status %s, got %s", domain.StatusPermissionDen, resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != string(domain.ReasonInsufficientAuthorizedScope) {
		t.Errorf("expected reason %s, got %v", domain.ReasonInsufficientAuthorizedScope, resp.Reason)
	}
}

func TestProcessAnswerRequest_CategoryIndexNotFound(t *testing.T) {
	service, sessionProvider, _, _, _ := setupTestService()

	// Add a session for a user that doesn't have a category index
	sessionProvider.AddSession(&models.Session{
		SessionID:      "sess_noindex",
		ParticipantID:  "user_noindex",
		AgentID:        "openclaw_whatsapp_bot",
		ApprovedScopes: []string{"preferences.food.read"},
		ExpiresAt:      time.Now().Add(time.Hour).Unix(),
		RemainingUses:  5,
		Revoked:        false,
	})

	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "user_noindex",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_noindex",
		Question:           "Order dinner",
		RequiredCategories: []string{"FOOD"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusCannotAnswer) {
		t.Errorf("expected status %s, got %s", domain.StatusCannotAnswer, resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != string(domain.ReasonCategoryNotFound) {
		t.Errorf("expected reason %s, got %v", domain.ReasonCategoryNotFound, resp.Reason)
	}
}

func TestProcessAnswerRequest_LLMFailure(t *testing.T) {
	service, _, _, _, llmProvider := setupTestService()

	llmProvider.SetShouldFail(true, errors.New("API error"))

	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_123",
		Question:           "Order dinner",
		RequiredCategories: []string{"FOOD"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusCannotAnswer) {
		t.Errorf("expected status %s, got %s", domain.StatusCannotAnswer, resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != string(domain.ReasonLLMCallFailed) {
		t.Errorf("expected reason %s, got %v", domain.ReasonLLMCallFailed, resp.Reason)
	}
}

func TestProcessAnswerRequest_EmptyLLMResponse(t *testing.T) {
	service, _, _, _, llmProvider := setupTestService()

	llmProvider.SetDefaultResponse("")

	req := &models.AnswerRequest{
		RequestID:          "req_123",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_123",
		Question:           "Order dinner",
		RequiredCategories: []string{"FOOD"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	resp := service.ProcessAnswerRequest(context.Background(), req)

	if resp.Status != string(domain.StatusCannotAnswer) {
		t.Errorf("expected status %s, got %s", domain.StatusCannotAnswer, resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != string(domain.ReasonEmptyResponse) {
		t.Errorf("expected reason %s, got %v", domain.ReasonEmptyResponse, resp.Reason)
	}
}

func TestCheckScopeCoverage(t *testing.T) {
	service, _, _, _, _ := setupTestService()

	tests := []struct {
		name           string
		categories     []string
		approvedScopes []string
		wantMissing    int
	}{
		{
			name:           "all scopes present",
			categories:     []string{"FOOD", "HEALTH"},
			approvedScopes: []string{"preferences.food.read", "health.read"},
			wantMissing:    0,
		},
		{
			name:           "missing one scope",
			categories:     []string{"FOOD", "ADDRESS"},
			approvedScopes: []string{"preferences.food.read"},
			wantMissing:    1,
		},
		{
			name:           "all scopes missing",
			categories:     []string{"FOOD", "HEALTH"},
			approvedScopes: []string{},
			wantMissing:    2,
		},
		{
			name:           "extra scopes present",
			categories:     []string{"FOOD"},
			approvedScopes: []string{"preferences.food.read", "health.read", "profile.address.read"},
			wantMissing:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing := service.checkScopeCoverage(tt.categories, tt.approvedScopes)
			if len(missing) != tt.wantMissing {
				t.Errorf("expected %d missing scopes, got %d: %v", tt.wantMissing, len(missing), missing)
			}
		})
	}
}
