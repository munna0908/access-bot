package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/providers"
)

// IntelligenceClient implements SessionProvider by calling the participant-intelligence-service.
type IntelligenceClient struct {
	baseURL    string
	httpClient *http.Client
}

// Ensure IntelligenceClient implements SessionProvider.
var _ providers.SessionProvider = (*IntelligenceClient)(nil)

// IntelligenceConfig holds configuration for the intelligence service client.
type IntelligenceConfig struct {
	BaseURL string
	Timeout time.Duration
}

// NewIntelligenceClient creates a new intelligence service client.
func NewIntelligenceClient(cfg IntelligenceConfig) *IntelligenceClient {
	return &IntelligenceClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// validateSessionRequest is the request body for the validation endpoint.
type validateSessionRequest struct {
	ParticipantID      string   `json:"participantId"`
	AgentID            string   `json:"agentId"`
	SessionID          string   `json:"sessionId"`
	RequiredCategories []string `json:"requiredCategories"`
	RequiredScopes     []string `json:"requiredScopes"`
	CurrentTime        int64    `json:"currentTime"`
}

// validateSessionResponse is the response from the validation endpoint.
type validateSessionResponse struct {
	Valid  bool    `json:"valid"`
	Reason *string `json:"reason"`
}

// ValidateSession validates a session by calling the intelligence service.
func (c *IntelligenceClient) ValidateSession(ctx context.Context, req *models.ValidateSessionRequest) error {
	url := fmt.Sprintf("%s/v1/sessions/validate", c.baseURL)

	body := validateSessionRequest{
		ParticipantID:      req.ParticipantID,
		AgentID:            req.AgentID,
		SessionID:          req.SessionID,
		RequiredCategories: req.RequiredCategories,
		RequiredScopes:     req.RequiredScopes,
		CurrentTime:        req.CurrentTime,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call intelligence service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("intelligence service returned status %d", resp.StatusCode)
	}

	var result validateSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Valid {
		reason := "unknown"
		if result.Reason != nil {
			reason = *result.Reason
		}
		return errors.New(reason)
	}

	return nil
}
