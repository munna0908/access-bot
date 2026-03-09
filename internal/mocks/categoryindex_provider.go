package mocks

import (
	"context"
	"sync"

	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/providers"
)

// MockCategoryIndexProvider is an in-memory mock implementation of CategoryIndexProvider.
type MockCategoryIndexProvider struct {
	mu      sync.RWMutex
	indexes map[string]*models.CategoryIndex
}

// Ensure MockCategoryIndexProvider implements CategoryIndexProvider.
var _ providers.CategoryIndexProvider = (*MockCategoryIndexProvider)(nil)

// NewMockCategoryIndexProvider creates a new mock category index provider with sample data.
func NewMockCategoryIndexProvider() *MockCategoryIndexProvider {
	provider := &MockCategoryIndexProvider{
		indexes: make(map[string]*models.CategoryIndex),
	}

	// Add sample participant category indexes
	provider.indexes["user_001"] = &models.CategoryIndex{
		ParticipantID: "user_001",
		Categories: map[string]string{
			"HEALTH":  "bafy_health_cid_001",
			"FOOD":    "bafy_food_cid_001",
			"ADDRESS": "bafy_address_cid_001",
			"PAYMENT": "bafy_payment_cid_001",
		},
	}

	provider.indexes["user_002"] = &models.CategoryIndex{
		ParticipantID: "user_002",
		Categories: map[string]string{
			"HEALTH":  "bafy_health_cid_002",
			"FOOD":    "bafy_food_cid_002",
			"ADDRESS": "bafy_address_cid_002",
			"PAYMENT": "bafy_payment_cid_002",
		},
	}

	provider.indexes["user_004"] = &models.CategoryIndex{
		ParticipantID: "user_002",
		Categories: map[string]string{
			"HEALTH":  "bafy_health_cid_002",
			"FOOD":    "bafy_food_cid_002",
			"ADDRESS": "bafy_address_cid_002",
			"PAYMENT": "bafy_payment_cid_002",
		},
	}

	// User with partial categories
	provider.indexes["user_partial"] = &models.CategoryIndex{
		ParticipantID: "user_partial",
		Categories: map[string]string{
			"FOOD": "bafy_food_cid_partial",
		},
	}

	return provider
}

// GetCategoryIndex retrieves the category-to-CID mapping for a participant.
func (m *MockCategoryIndexProvider) GetCategoryIndex(ctx context.Context, participantID string) (*models.CategoryIndex, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	index, ok := m.indexes[participantID]
	if !ok {
		return nil, nil
	}

	// Return a copy to prevent mutations
	indexCopy := &models.CategoryIndex{
		ParticipantID: index.ParticipantID,
		Categories:    make(map[string]string),
	}
	for k, v := range index.Categories {
		indexCopy.Categories[k] = v
	}

	return indexCopy, nil
}

// AddCategoryIndex adds a category index to the mock store (for testing).
func (m *MockCategoryIndexProvider) AddCategoryIndex(index *models.CategoryIndex) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.indexes[index.ParticipantID] = index
}

// RemoveCategoryIndex removes a category index from the mock store (for testing).
func (m *MockCategoryIndexProvider) RemoveCategoryIndex(participantID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.indexes, participantID)
}
