package mocks

import (
	"context"
	"sync"

	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/providers"
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
			"finance.payment.read",
		},
		ExpiresAt:     1773165600,
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
			"finance.payment.read",
		},
		ExpiresAt:     1773165600,
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
		ExpiresAt:     1773165600,
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
		ExpiresAt:     1773165600,
		RemainingUses: 0,
		Revoked:       false,
	}

	return provider
}

// GetSession retrieves a session by its ID.
func (m *MockSessionProvider) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, nil
	}

	// Return a copy to prevent mutations
	sessionCopy := *session
	scopesCopy := make([]string, len(session.ApprovedScopes))
	copy(scopesCopy, session.ApprovedScopes)
	sessionCopy.ApprovedScopes = scopesCopy

	return &sessionCopy, nil
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
