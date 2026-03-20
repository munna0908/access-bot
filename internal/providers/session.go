package providers

import (
	"context"

	"github.com/access-bot/internal/models"
)

// SessionProvider defines the interface for session validation.
type SessionProvider interface {
	// ValidateSession validates a session against the participant intelligence service.
	// Returns nil if valid, or an error with the failure reason.
	ValidateSession(ctx context.Context, req *models.ValidateSessionRequest) error
}
