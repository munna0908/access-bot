package providers

import (
	"context"

	"github.com/access-bot/internal/models"
)

// CategoryIndexProvider defines the interface for participant category CID lookup.
type CategoryIndexProvider interface {
	// GetCategoryIndex retrieves the category-to-CID mapping for a participant.
	// Returns nil if the participant is not found.
	GetCategoryIndex(ctx context.Context, participantID string) (*models.CategoryIndex, error)
}
