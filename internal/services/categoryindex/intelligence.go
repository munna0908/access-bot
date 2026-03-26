package categoryindex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/providers"
)

// IntelligenceClient implements CategoryIndexProvider by reading on-chain
// category→CID mappings from the participant-intelligence-service.
type IntelligenceClient struct {
	httpClient *http.Client
	baseURL    string
}

// IntelligenceConfig holds configuration for the intelligence client.
type IntelligenceConfig struct {
	BaseURL string
	Timeout time.Duration
}

// Ensure IntelligenceClient implements CategoryIndexProvider.
var _ providers.CategoryIndexProvider = (*IntelligenceClient)(nil)

// NewIntelligenceClient creates a new CategoryIndexProvider backed by the
// participant-intelligence-service.
func NewIntelligenceClient(cfg IntelligenceConfig) *IntelligenceClient {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &IntelligenceClient{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    cfg.BaseURL,
	}
}

// intelligenceObjectResponse mirrors the JSON returned by
// GET /v1/intelligence/:participantId in the intelligence service.
type intelligenceObjectResponse struct {
	ParticipantID string                     `json:"participantId"`
	CategoryRefs  map[string]categoryRefJSON `json:"categoryRefs"`
}

// categoryRefJSON mirrors the CategoryRef shape in the intelligence service.
type categoryRefJSON struct {
	Ref           string `json:"ref"`
	SchemaVersion string `json:"schemaVersion"`
	UpdatedAt     int64  `json:"updatedAt"`
}

// GetCategoryIndex fetches the CID mapping for all categories for a participant
// from the intelligence service and returns it as a CategoryIndex.
func (c *IntelligenceClient) GetCategoryIndex(ctx context.Context, participantID string) (*models.CategoryIndex, error) {
	url := fmt.Sprintf("%s/v1/intelligence/%s", c.baseURL, participantID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to intelligence service failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		// Participant has no on-chain data yet — return an empty index rather
		// than a hard error so the access layer can respond with a clear reason.
		return &models.CategoryIndex{
			ParticipantID: participantID,
			Categories:    map[string]string{},
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("intelligence service returned %d: %s", resp.StatusCode, string(body))
	}

	var obj intelligenceObjectResponse
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, fmt.Errorf("failed to parse intelligence response: %w", err)
	}

	// Convert categoryRefs → flat category→CID map, skipping empty refs.
	categories := make(map[string]string, len(obj.CategoryRefs))
	for cat, ref := range obj.CategoryRefs {
		if ref.Ref != "" {
			categories[cat] = ref.Ref
		}
	}

	return &models.CategoryIndex{
		ParticipantID: participantID,
		Categories:    categories,
	}, nil
}
