package mocks

import (
	"context"
	"errors"
	"sync"

	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/providers"
)

// Session validation error reasons (matching participant-intelligence-service).
var (
	ErrSessionNotFound            = errors.New("not_found")
	ErrSessionExpired             = errors.New("expired")
	ErrSessionRevoked             = errors.New("revoked")
	ErrSessionExhausted           = errors.New("exhausted")
	ErrSessionAgentMismatch       = errors.New("agent_mismatch")
	ErrSessionParticipantMismatch = errors.New("participant_mismatch")
	ErrSessionMissingScope        = errors.New("missing_scope")
)

// MockSessionProvider is an in-memory mock implementation of SessionProvider.
type MockSessionProvider struct {
	mu       sync.RWMutex
	sessions map[string]*models.Session
}

// Ensure MockSessionProvider implements SessionProvider.
var _ providers.SessionProvider = (*MockSessionProvider)(nil)

// NewMockSessionProvider creates a new mock session provider with sample data.
func NewMockSessionProvider() *MockSessionProvider {
	provider := &MockSessionProvider{
		sessions: make(map[string]*models.Session),
	}

	// Add sample sessions
	provider.sessions["sess_123"] = &models.Session{
		SessionID:     "sess_123",
		ParticipantID: "user_001",
		AgentID:       "openclaw_whatsapp_bot",
		ApprovedScopes: []string{
			"preferences.food.read",
			"health.read",
		},
		ExpiresAt:     1900000000, // Far future timestamp (year 2030)
		RemainingUses: 5,
		Revoked:       false,
	}

	provider.sessions["sess_456"] = &models.Session{
		SessionID:     "sess_456",
		ParticipantID: "user_002",
		AgentID:       "openclaw_telegram_bot",
		ApprovedScopes: []string{
			"preferences.food.read",
			"health.read",
			"profile.address.read",
			"schedule.read",
		},
		ExpiresAt:     2000000000, // Far future timestamp (year 2033)
		RemainingUses: 10,
		Revoked:       false,
	}
	provider.sessions["sess_1234"] = &models.Session{
		SessionID:     "sess_456",
		ParticipantID: "user_004",
		AgentID:       "openclaw_telegram_bot",
		ApprovedScopes: []string{
			"preferences.food.read",
			"profile.address.read",
			"schedule.read",
		},
		ExpiresAt:     2000000000, // Far future timestamp (year 2033)
		RemainingUses: 10,
		Revoked:       false,
	}

	provider.sessions["sess_expired"] = &models.Session{
		SessionID:     "sess_expired",
		ParticipantID: "user_001",
		AgentID:       "openclaw_whatsapp_bot",
		ApprovedScopes: []string{
			"preferences.food.read",
		},
		ExpiresAt:     1609459200, // Past timestamp (2021-01-01)
		RemainingUses: 5,
		Revoked:       false,
	}

	provider.sessions["sess_revoked"] = &models.Session{
		SessionID:     "sess_revoked",
		ParticipantID: "user_001",
		AgentID:       "openclaw_whatsapp_bot",
		ApprovedScopes: []string{
			"preferences.food.read",
		},
		ExpiresAt:     2000000000, // Far future timestamp (year 2033)
		RemainingUses: 5,
		Revoked:       true,
	}

	provider.sessions["sess_exhausted"] = &models.Session{
		SessionID:     "sess_exhausted",
		ParticipantID: "user_001",
		AgentID:       "openclaw_whatsapp_bot",
		ApprovedScopes: []string{
			"preferences.food.read",
		},
		ExpiresAt:     2000000000, // Far future timestamp (year 2033)
		RemainingUses: 0,
		Revoked:       false,
	}

	return provider
}

// ValidateSession validates a session against the stored sessions.
func (m *MockSessionProvider) ValidateSession(ctx context.Context, req *models.ValidateSessionRequest) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[req.SessionID]
	if !ok {
		return ErrSessionNotFound
	}

	// Check participant_id matches
	if session.ParticipantID != req.ParticipantID {
		return ErrSessionParticipantMismatch
	}

	// Check agent_id matches
	if session.AgentID != req.AgentID {
		return ErrSessionAgentMismatch
	}

	// Check not expired
	if session.ExpiresAt <= req.CurrentTime {
		return ErrSessionExpired
	}

	// Check not revoked
	if session.Revoked {
		return ErrSessionRevoked
	}

	// Check usage limit
	if session.RemainingUses <= 0 {
		return ErrSessionExhausted
	}

	// Check all required scopes are approved
	scopeSet := make(map[string]bool)
	for _, scope := range session.ApprovedScopes {
		scopeSet[scope] = true
	}
	for _, requiredScope := range req.RequiredScopes {
		if !scopeSet[requiredScope] {
			return ErrSessionMissingScope
		}
	}

	return nil
}

// AddSession adds a session to the mock store (for testing).
func (m *MockSessionProvider) AddSession(session *models.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.SessionID] = session
}

// RemoveSession removes a session from the mock store (for testing).
func (m *MockSessionProvider) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}
