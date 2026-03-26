package mocks

import (
	"context"
	"strings"

	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/providers"
)

// ─── DemoSessionProvider ─────────────────────────────────────────────────────
// Accepts every session without validation. Used when DEMO_MODE=true.

type DemoSessionProvider struct{}

var _ providers.SessionProvider = (*DemoSessionProvider)(nil)

func NewDemoSessionProvider() *DemoSessionProvider { return &DemoSessionProvider{} }

func (d *DemoSessionProvider) ValidateSession(_ context.Context, _ *models.ValidateSessionRequest) error {
	return nil // demo: all sessions are valid
}

// ─── DemoCategoryIndexProvider ───────────────────────────────────────────────
// Returns common demo CIDs for every participant. Used when DEMO_MODE=true.

type DemoCategoryIndexProvider struct{}

var _ providers.CategoryIndexProvider = (*DemoCategoryIndexProvider)(nil)

func NewDemoCategoryIndexProvider() *DemoCategoryIndexProvider {
	return &DemoCategoryIndexProvider{}
}

func (d *DemoCategoryIndexProvider) GetCategoryIndex(_ context.Context, participantID string) (*models.CategoryIndex, error) {
	return &models.CategoryIndex{
		ParticipantID: participantID,
		Categories: map[string]string{
			"FOOD":    "demo_food_cid",
			"HEALTH":  "demo_health_cid",
			"ADDRESS": "demo_address_cid",
			"PAYMENT": "demo_payment_cid",
		},
	}, nil
}

// ─── DemoFilesystemProvider ───────────────────────────────────────────────────
// Returns the content of sample_data/*.md files inline.
// Used when DEMO_MODE=true so no IPFS/Pinata is required.

type DemoFilesystemProvider struct {
	files map[string]*models.FileContent
}

var _ providers.FilesystemProvider = (*DemoFilesystemProvider)(nil)

func NewDemoFilesystemProvider() *DemoFilesystemProvider {
	return &DemoFilesystemProvider{
		files: map[string]*models.FileContent{
			"demo_food_cid":    {CID: "demo_food_cid", ContentType: "text/markdown", Content: demoFoodContent},
			"demo_health_cid":  {CID: "demo_health_cid", ContentType: "text/markdown", Content: demoHealthContent},
			"demo_address_cid": {CID: "demo_address_cid", ContentType: "text/markdown", Content: demoAddressContent},
			"demo_payment_cid": {CID: "demo_payment_cid", ContentType: "text/markdown", Content: demoPaymentContent},
		},
	}
}

func (d *DemoFilesystemProvider) FetchByCID(_ context.Context, cid string) (*models.FileContent, error) {
	f, ok := d.files[cid]
	if !ok {
		return nil, nil
	}
	copy := *f
	return &copy, nil
}

// ─── DemoLLMProvider ─────────────────────────────────────────────────────────
// Returns realistic numbered food lists and address text for demo prompts.
// Wraps MockLLMProvider so all existing mock behaviour is preserved; only
// food-ordering and address queries get a more realistic response.

type DemoLLMProvider struct {
	inner *MockLLMProvider
}

var _ providers.LLMProvider = (*DemoLLMProvider)(nil)

func NewDemoLLMProvider() *DemoLLMProvider {
	return &DemoLLMProvider{inner: NewMockLLMProvider()}
}

func (d *DemoLLMProvider) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	lower := strings.ToLower(userPrompt)

	// Food ordering → always return a clean numbered list
	if strings.Contains(lower, "list exactly 3") || strings.Contains(lower, "numbered list") ||
		(strings.Contains(lower, "order food") || strings.Contains(lower, "order some food")) {
		return "1. Masala Dosa\n2. Pad Thai (no egg, no peanuts)\n3. Vegetable Biryani", nil
	}

	// Address query
	if strings.Contains(lower, "delivery address") || strings.Contains(lower, "my address") {
		return "123 Palm Grove Apartments, Flat 4B, Tower 2, Koramangala, Bangalore – 560034.\nDelivery between 10 AM – 8 PM. Gate code: 1234.", nil
	}

	return d.inner.Complete(ctx, systemPrompt, userPrompt)
}

func (d *DemoLLMProvider) GetModelName() string { return "claude-demo" }

// ─── Inline sample_data content ───────────────────────────────────────────────

const demoFoodContent = `# Food Preferences

## Dietary Type
- Vegetarian (Lacto-ovo, but avoiding dairy due to intolerance)

## Cuisine Preferences
1. South Indian (favorite)
2. North Indian
3. Italian
4. Thai

## Favorite Dishes
- Masala Dosa
- Vegetable Biryani
- Margherita Pizza (dairy-free cheese)
- Pad Thai (no egg)

## Disliked Foods
- Bitter gourd
- Okra
- Mushrooms
- Eggplant

## Spice Tolerance
- Medium to High

## Meal Preferences
| Meal | Preference |
|------|------------|
| Breakfast | Light, South Indian (idli, dosa) |
| Lunch | Full meal, rice-based |
| Dinner | Light, early (before 8 PM) |

## Favorite Restaurants
- Saravana Bhavan (South Indian)
- Mainland China (Indo-Chinese)
- Pizza Hut (for dairy-free options)

## Food Ordering Notes
- Prefers home delivery over dine-in
- Usually orders for 1-2 people
- Typical budget: Rs 300-500 per meal`

const demoHealthContent = `# Health Profile

## Medical Conditions
- Type 2 Diabetes (diagnosed 2019)
- Mild Hypertension
- Lactose Intolerance

## Allergies
- Peanuts (severe)
- Shellfish (moderate)
- Penicillin

## Medications
| Medication | Dosage | Frequency |
|------------|--------|-----------|
| Metformin | 500mg | Twice daily |
| Lisinopril | 10mg | Once daily |

## Dietary Restrictions
- Low sodium diet (< 2000mg/day)
- Low glycemic index foods preferred
- Avoid dairy products

## Recent Vitals
- Blood Pressure: 128/82 mmHg
- Blood Sugar (fasting): 126 mg/dL
- Weight: 78 kg
- Height: 175 cm

## Emergency Contact
- Name: Priya Sharma
- Relation: Spouse
- Phone: +91 98765 43210`

const demoAddressContent = `# Address Information

## Primary Address (Home)
- **Address Line 1:** 123 Palm Grove Apartments
- **Address Line 2:** Flat 4B, Tower 2
- **Landmark:** Near City Mall
- **Area:** Koramangala
- **City:** Bangalore
- **State:** Karnataka
- **PIN Code:** 560034
- **Country:** India

## Delivery Instructions
- Gate code: 1234
- Security will call before allowing entry
- Prefer delivery between 10 AM - 8 PM
- Leave at door if not available (for trusted deliveries)

## Contact for Delivery
- Phone: +91 99887 76655
- Alternate: +91 98765 43210

## Work Address
- **Company:** TechCorp Solutions
- **Address:** 456 Tech Park, Whitefield
- **City:** Bangalore
- **PIN Code:** 560066
- **Delivery Hours:** 10 AM - 6 PM (weekdays only)`

const demoPaymentContent = `# Payment Information

## Preferred Payment Methods
1. Apple Pay (default)
2. Google Pay
3. Credit Card

## UPI IDs
- rahul@okicici (primary)
- rahul.sharma@paytm

## Budget Preferences
| Category | Typical Budget |
|----------|----------------|
| Meals | Rs 300-500 |
| Groceries | Rs 2000-3000/week |

## Subscription Services
- Swiggy One (active)

## Spending Limits
- Daily limit: Rs 5000
- Single transaction alert: Above Rs 2000`
