package mocks

import (
	"context"

	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/providers"
)

// CIDCategoryIndexProvider returns a fixed set of Pinata CIDs for every
// participant. Used when CATEGORY_CID_* env vars are set, so the access-bot
// can fetch real files from Pinata without requiring on-chain registration.
type CIDCategoryIndexProvider struct {
	cids map[string]string
}

var _ providers.CategoryIndexProvider = (*CIDCategoryIndexProvider)(nil)

// NewCIDCategoryIndexProvider creates a provider that returns the given CIDs
// for all participants.
func NewCIDCategoryIndexProvider(food, health, address, schedule string) *CIDCategoryIndexProvider {
	cids := map[string]string{}
	if food != "" {
		cids["FOOD"] = food
	}
	if health != "" {
		cids["HEALTH"] = health
	}
	if address != "" {
		cids["ADDRESS"] = address
	}
	if schedule != "" {
		cids["SCHEDULE"] = schedule
	}
	return &CIDCategoryIndexProvider{cids: cids}
}

func (p *CIDCategoryIndexProvider) GetCategoryIndex(_ context.Context, participantID string) (*models.CategoryIndex, error) {
	return &models.CategoryIndex{
		ParticipantID: participantID,
		Categories:    p.cids,
	}, nil
}
