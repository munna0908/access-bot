package providers

import (
	"context"

	"github.com/access-bot/internal/models"
)

// FilesystemProvider defines the interface for fetching content by CID.
type FilesystemProvider interface {
	// FetchByCID retrieves file content by its CID.
	// Returns nil if the CID is not found.
	FetchByCID(ctx context.Context, cid string) (*models.FileContent, error)
}
