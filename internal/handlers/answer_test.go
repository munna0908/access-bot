package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/access-bot/internal/logging"
	"github.com/access-bot/internal/mocks"
	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/services/accesslayer"
)

func setupTestHandler() (*AnswerHandler, *mocks.MockLLMProvider) {
	sessionProvider := mocks.NewMockSessionProvider()
	categoryIndexProvider := mocks.NewMockCategoryIndexProvider()
	filesystemProvider := mocks.NewMockFilesystemProvider()
	llmProvider := mocks.NewMockLLMProvider()
	logger := logging.NewLogger("error")

	service := accesslayer.NewService(
		sessionProvider,
		categoryIndexProvider,
		filesystemProvider,
		llmProvider,
		logger,
	)

	handler := NewAnswerHandler(service, logger)
	return handler, llmProvider
}

func TestAnswerHandler_HappyPath(t *testing.T) {
	handler, llmProvider := setupTestHandler()
	llmProvider.SetDefaultResponse("Here's a great dinner recommendation.")

	reqBody := models.AnswerRequest{
		RequestID:          "req_test_001",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_123",
		Question:           "What should I order for dinner?",
		RequiredCategories: []string{"FOOD", "HEALTH"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/answer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp models.AnswerResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.Answer == "" {
		t.Error("expected non-empty answer")
	}
}

func TestAnswerHandler_MethodNotAllowed(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/v1/answer", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestAnswerHandler_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/v1/answer", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestAnswerHandler_ValidationError(t *testing.T) {
	handler, _ := setupTestHandler()

	// Missing required fields
	reqBody := models.AnswerRequest{
		RequestID: "req_test_001",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/answer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var resp models.AnswerResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "invalid_request" {
		t.Errorf("expected status 'invalid_request', got %q", resp.Status)
	}
}

func TestAnswerHandler_PermissionDenied(t *testing.T) {
	handler, _ := setupTestHandler()

	// Request category not authorized in session
	reqBody := models.AnswerRequest{
		RequestID:          "req_test_001",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "sess_123",
		Question:           "What's my address?",
		RequiredCategories: []string{"ADDRESS"}, // Not authorized in sess_123
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/answer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}

	var resp models.AnswerResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "permission_denied" {
		t.Errorf("expected status 'permission_denied', got %q", resp.Status)
	}
	if resp.Answer != "Sorry, I can't answer that." {
		t.Errorf("expected safe failure message, got %q", resp.Answer)
	}
}

func TestAnswerHandler_SessionNotFound(t *testing.T) {
	handler, _ := setupTestHandler()

	// Session not found now returns permission_denied
	reqBody := models.AnswerRequest{
		RequestID:          "req_test_001",
		ParticipantID:      "user_001",
		AgentID:            "openclaw_whatsapp_bot",
		SessionID:          "nonexistent_session",
		Question:           "Order dinner",
		RequiredCategories: []string{"FOOD"},
		Channel:            "whatsapp",
		Timestamp:          time.Now().Unix(),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/answer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}

	var resp models.AnswerResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "permission_denied" {
		t.Errorf("expected status 'permission_denied', got %q", resp.Status)
	}
}

func TestHealthHandler(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %q", resp["status"])
	}
	if resp["service"] != "access-layer" {
		t.Errorf("expected service 'access-layer', got %q", resp["service"])
	}
}

func TestRouter(t *testing.T) {
	sessionProvider := mocks.NewMockSessionProvider()
	categoryIndexProvider := mocks.NewMockCategoryIndexProvider()
	filesystemProvider := mocks.NewMockFilesystemProvider()
	llmProvider := mocks.NewMockLLMProvider()
	logger := logging.NewLogger("error")

	service := accesslayer.NewService(
		sessionProvider,
		categoryIndexProvider,
		filesystemProvider,
		llmProvider,
		logger,
	)

	router := NewRouter(service, logger)

	// Test root path
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	router.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d for root path, got %d", http.StatusOK, rr.Code)
	}

	// Test health endpoint
	req = httptest.NewRequest(http.MethodGet, "/health", nil)
	rr = httptest.NewRecorder()
	router.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d for health endpoint, got %d", http.StatusOK, rr.Code)
	}

	// Test 404
	req = httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rr = httptest.NewRecorder()
	router.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d for nonexistent path, got %d", http.StatusNotFound, rr.Code)
	}
}
