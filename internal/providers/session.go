package providers

import (
	"context"

	"github.com/access-bot/internal/models"
)

// SessionProvider defines the interface for session retrieval.
type SessionProvider interface {
	// GetSession retrieves a session by its ID.
	// Returns nil if the session is not found.
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
}
