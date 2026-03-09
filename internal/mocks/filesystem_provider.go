package mocks

import (
	"context"
	"sync"

	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/providers"
)

// MockFilesystemProvider is an in-memory mock implementation of FilesystemProvider.
type MockFilesystemProvider struct {
	mu    sync.RWMutex
	files map[string]*models.FileContent
}

// Ensure MockFilesystemProvider implements FilesystemProvider.
var _ providers.FilesystemProvider = (*MockFilesystemProvider)(nil)

// NewMockFilesystemProvider creates a new mock filesystem provider with sample data.
func NewMockFilesystemProvider() *MockFilesystemProvider {
	provider := &MockFilesystemProvider{
		files: make(map[string]*models.FileContent),
	}

	// Add sample file contents for user_001
	provider.files["bafy_health_cid_001"] = &models.FileContent{
		CID:         "bafy_health_cid_001",
		ContentType: "text/markdown",
		Content: `# Health Profile

## Medical Conditions
- Type 2 Diabetes (managed)
- Mild hypertension

## Allergies
- Lactose intolerant
- No known drug allergies

## Dietary Restrictions (Medical)
- Low sodium diet recommended
- Avoid high glycemic index foods

## Current Medications
- Metformin 500mg twice daily

## Last Updated
2024-01-15`,
	}

	provider.files["bafy_food_cid_001"] = &models.FileContent{
		CID:         "bafy_food_cid_001",
		ContentType: "text/markdown",
		Content: `# Food Preferences

## Favorite Cuisines
- South Indian
- Thai
- Mediterranean

## Diet Type
- Vegetarian

## Foods to Avoid
- Mushroom
- Dairy products (lactose intolerant)
- Very spicy food

## Meal Preferences
- Prefers home delivery over dining out
- Usually eats dinner around 7:30 PM
- Enjoys trying new vegetarian restaurants

## Favorite Dishes
- Masala Dosa
- Pad Thai (no egg)
- Falafel wrap`,
	}

	provider.files["bafy_address_cid_001"] = &models.FileContent{
		CID:         "bafy_address_cid_001",
		ContentType: "text/markdown",
		Content: `# Address Information

## Primary Address
123 Main Street
Apartment 4B
San Francisco, CA 94102
USA

## Delivery Instructions
- Ring doorbell twice
- Leave at door if no answer
- Gate code: 1234

## Preferred Delivery Times
- Weekdays: 6 PM - 9 PM
- Weekends: 12 PM - 9 PM`,
	}

	provider.files["bafy_payment_cid_001"] = &models.FileContent{
		CID:         "bafy_payment_cid_001",
		ContentType: "text/markdown",
		Content: `# Payment Preferences

## Preferred Payment Method
- Digital wallet (Apple Pay)

## Budget Guidelines
- Typical meal budget: $15-25
- Special occasion budget: up to $50

## Tipping Preferences
- Default tip: 20%
- Excellent service: 25%`,
	}

	// Add sample file contents for user_002
	provider.files["bafy_health_cid_002"] = &models.FileContent{
		CID:         "bafy_health_cid_002",
		ContentType: "text/markdown",
		Content: `# Health Profile

## Medical Conditions
- None reported

## Allergies
- Peanuts (severe)
- Shellfish

## Dietary Restrictions (Medical)
- Must avoid all peanut products
- Must avoid shellfish

## Last Updated
2024-02-01`,
	}

	provider.files["bafy_food_cid_002"] = &models.FileContent{
		CID:         "bafy_food_cid_002",
		ContentType: "text/markdown",
		Content: `# Food Preferences

## Favorite Cuisines
- Japanese
- Italian
- Mexican

## Diet Type
- Omnivore

## Foods to Avoid
- Peanuts (allergy)
- Shellfish (allergy)

## Favorite Dishes
- Ramen
- Margherita Pizza
- Tacos al Pastor`,
	}

	provider.files["bafy_address_cid_002"] = &models.FileContent{
		CID:         "bafy_address_cid_002",
		ContentType: "text/markdown",
		Content: `# Address Information

## Primary Address
456 Oak Avenue
Suite 200
New York, NY 10001
USA

## Delivery Instructions
- Call on arrival
- Front desk will accept packages`,
	}

	provider.files["bafy_payment_cid_002"] = &models.FileContent{
		CID:         "bafy_payment_cid_002",
		ContentType: "text/markdown",
		Content: `# Payment Preferences

## Preferred Payment Method
- Credit card

## Budget Guidelines
- Typical meal budget: $20-35`,
	}

	// Partial user data
	provider.files["bafy_food_cid_partial"] = &models.FileContent{
		CID:         "bafy_food_cid_partial",
		ContentType: "text/markdown",
		Content: `# Food Preferences

## Diet Type
- Vegan

## Favorite Cuisines
- Indian
- Ethiopian`,
	}

	return provider
}

// FetchByCID retrieves file content by its CID.
func (m *MockFilesystemProvider) FetchByCID(ctx context.Context, cid string) (*models.FileContent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	file, ok := m.files[cid]
	if !ok {
		return nil, nil
	}

	// Return a copy to prevent mutations
	fileCopy := *file
	return &fileCopy, nil
}

// AddFile adds a file to the mock store (for testing).
func (m *MockFilesystemProvider) AddFile(file *models.FileContent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[file.CID] = file
}

// RemoveFile removes a file from the mock store (for testing).
func (m *MockFilesystemProvider) RemoveFile(cid string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.files, cid)
}
