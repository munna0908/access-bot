package validation

import (
	"strings"
	"testing"

	"github.com/access-bot/internal/models"
)

func TestValidateAnswerRequest(t *testing.T) {
	tests := []struct {
		name      string
		req       *models.AnswerRequest
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid request",
			req: &models.AnswerRequest{
				RequestID:          "req_123",
				ParticipantID:      "user_001",
				AgentID:            "openclaw_whatsapp_bot",
				SessionID:          "sess_123",
				Question:           "What should I order for dinner?",
				RequiredCategories: []string{"FOOD", "HEALTH"},
				Channel:            "whatsapp",
				Timestamp:          1773162200,
			},
			wantError: false,
		},
		{
			name: "missing request_id",
			req: &models.AnswerRequest{
				ParticipantID:      "user_001",
				AgentID:            "openclaw_whatsapp_bot",
				SessionID:          "sess_123",
				Question:           "What should I order for dinner?",
				RequiredCategories: []string{"FOOD"},
				Channel:            "whatsapp",
				Timestamp:          1773162200,
			},
			wantError: true,
			errorMsg:  "request_id is required",
		},
		{
			name: "missing participant_id",
			req: &models.AnswerRequest{
				RequestID:          "req_123",
				AgentID:            "openclaw_whatsapp_bot",
				SessionID:          "sess_123",
				Question:           "What should I order for dinner?",
				RequiredCategories: []string{"FOOD"},
				Channel:            "whatsapp",
				Timestamp:          1773162200,
			},
			wantError: true,
			errorMsg:  "participant_id is required",
		},
		{
			name: "missing agent_id",
			req: &models.AnswerRequest{
				RequestID:          "req_123",
				ParticipantID:      "user_001",
				SessionID:          "sess_123",
				Question:           "What should I order for dinner?",
				RequiredCategories: []string{"FOOD"},
				Channel:            "whatsapp",
				Timestamp:          1773162200,
			},
			wantError: true,
			errorMsg:  "agent_id is required",
		},
		{
			name: "missing session_id",
			req: &models.AnswerRequest{
				RequestID:          "req_123",
				ParticipantID:      "user_001",
				AgentID:            "openclaw_whatsapp_bot",
				Question:           "What should I order for dinner?",
				RequiredCategories: []string{"FOOD"},
				Channel:            "whatsapp",
				Timestamp:          1773162200,
			},
			wantError: true,
			errorMsg:  "session_id is required",
		},
		{
			name: "missing question",
			req: &models.AnswerRequest{
				RequestID:          "req_123",
				ParticipantID:      "user_001",
				AgentID:            "openclaw_whatsapp_bot",
				SessionID:          "sess_123",
				RequiredCategories: []string{"FOOD"},
				Channel:            "whatsapp",
				Timestamp:          1773162200,
			},
			wantError: true,
			errorMsg:  "question is required",
		},
		{
			name: "empty required_categories",
			req: &models.AnswerRequest{
				RequestID:          "req_123",
				ParticipantID:      "user_001",
				AgentID:            "openclaw_whatsapp_bot",
				SessionID:          "sess_123",
				Question:           "What should I order for dinner?",
				RequiredCategories: []string{},
				Channel:            "whatsapp",
				Timestamp:          1773162200,
			},
			wantError: true,
			errorMsg:  "required_categories must contain at least one category",
		},
		{
			name: "invalid category",
			req: &models.AnswerRequest{
				RequestID:          "req_123",
				ParticipantID:      "user_001",
				AgentID:            "openclaw_whatsapp_bot",
				SessionID:          "sess_123",
				Question:           "What should I order for dinner?",
				RequiredCategories: []string{"INVALID_CATEGORY"},
				Channel:            "whatsapp",
				Timestamp:          1773162200,
			},
			wantError: true,
			errorMsg:  "invalid category",
		},
		{
			name: "missing channel",
			req: &models.AnswerRequest{
				RequestID:          "req_123",
				ParticipantID:      "user_001",
				AgentID:            "openclaw_whatsapp_bot",
				SessionID:          "sess_123",
				Question:           "What should I order for dinner?",
				RequiredCategories: []string{"FOOD"},
				Timestamp:          1773162200,
			},
			wantError: true,
			errorMsg:  "channel is required",
		},
		{
			name: "invalid timestamp",
			req: &models.AnswerRequest{
				RequestID:          "req_123",
				ParticipantID:      "user_001",
				AgentID:            "openclaw_whatsapp_bot",
				SessionID:          "sess_123",
				Question:           "What should I order for dinner?",
				RequiredCategories: []string{"FOOD"},
				Channel:            "whatsapp",
				Timestamp:          0,
			},
			wantError: true,
			errorMsg:  "timestamp must be a positive integer",
		},
		{
			name: "multiple validation errors",
			req: &models.AnswerRequest{
				RequestID:          "",
				ParticipantID:      "",
				AgentID:            "",
				SessionID:          "",
				Question:           "",
				RequiredCategories: []string{},
				Channel:            "",
				Timestamp:          0,
			},
			wantError: true,
			errorMsg:  "request_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAnswerRequest(tt.req)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateCategories(t *testing.T) {
	tests := []struct {
		name       string
		categories []string
		wantValid  []string
		wantError  bool
	}{
		{
			name:       "all valid categories",
			categories: []string{"HEALTH", "FOOD", "ADDRESS", "PAYMENT"},
			wantValid:  []string{"HEALTH", "FOOD", "ADDRESS", "PAYMENT"},
			wantError:  false,
		},
		{
			name:       "single valid category",
			categories: []string{"FOOD"},
			wantValid:  []string{"FOOD"},
			wantError:  false,
		},
		{
			name:       "invalid category",
			categories: []string{"INVALID"},
			wantError:  true,
		},
		{
			name:       "mixed valid and invalid",
			categories: []string{"FOOD", "INVALID"},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ValidateCategories(tt.categories)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if len(valid) != len(tt.wantValid) {
					t.Errorf("expected %d valid categories, got %d", len(tt.wantValid), len(valid))
				}
			}
		})
	}
}
