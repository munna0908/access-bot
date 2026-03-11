package filesystem

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/providers"
)

const (
	defaultPinataGateway = "https://gateway.pinata.cloud/ipfs"
	pinataAPIURL         = "https://api.pinata.cloud"
)

// PinataClient implements FilesystemProvider using Pinata's IPFS gateway.
type PinataClient struct {
	httpClient *http.Client
	gatewayURL string
	gatewayKey string // For dedicated gateway authentication
	jwt        string // For private file access
	isPrivate  bool   // Whether to use private file access
}

// Ensure PinataClient implements FilesystemProvider.
var _ providers.FilesystemProvider = (*PinataClient)(nil)

// PinataConfig holds configuration for the Pinata client.
type PinataConfig struct {
	GatewayURL string        // Custom gateway URL (e.g., https://your-gateway.mypinata.cloud/ipfs)
	GatewayKey string        // Optional: Gateway key for dedicated gateways
	JWT        string        // Optional: Pinata JWT for private file access
	IsPrivate  bool          // Whether files are private (requires JWT)
	Timeout    time.Duration // HTTP client timeout
}

// NewPinataClient creates a new Pinata filesystem client.
func NewPinataClient(cfg PinataConfig) *PinataClient {
	gatewayURL := cfg.GatewayURL
	if gatewayURL == "" {
		gatewayURL = defaultPinataGateway
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &PinataClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		gatewayURL: gatewayURL,
		gatewayKey: cfg.GatewayKey,
		jwt:        cfg.JWT,
		isPrivate:  cfg.IsPrivate,
	}
}

// accessLinkRequest is the request body for creating private access links.
type accessLinkRequest struct {
	URL     string `json:"url"`
	Expires int    `json:"expires"`
	Date    int64  `json:"date"`
	Method  string `json:"method"`
}

// accessLinkResponse is the response from the access link API.
type accessLinkResponse struct {
	Data string `json:"data"` // The signed URL
}

// FetchByCID retrieves file content from Pinata by CID.
func (p *PinataClient) FetchByCID(ctx context.Context, cid string) (*models.FileContent, error) {
	if cid == "" {
		return nil, fmt.Errorf("CID cannot be empty")
	}

	var fileURL string

	if p.isPrivate && p.jwt != "" {
		// For private files, create a temporary access link first
		signedURL, err := p.createAccessLink(ctx, cid)
		if err != nil {
			return nil, fmt.Errorf("failed to create access link: %w", err)
		}
		fileURL = signedURL
	} else {
		// For public files, use gateway directly
		fileURL = fmt.Sprintf("%s/%s", p.gatewayURL, cid)
	}

	return p.fetchFromURL(ctx, cid, fileURL)
}

// createAccessLink creates a temporary signed URL for private file access.
func (p *PinataClient) createAccessLink(ctx context.Context, cid string) (string, error) {
	// For private files, use /files/ path instead of /ipfs/
	// Extract base gateway URL (remove /ipfs suffix if present)
	baseGateway := p.gatewayURL
	if len(baseGateway) > 5 && baseGateway[len(baseGateway)-5:] == "/ipfs" {
		baseGateway = baseGateway[:len(baseGateway)-5]
	}
	gatewayFileURL := fmt.Sprintf("%s/files/%s", baseGateway, cid)

	reqBody := accessLinkRequest{
		URL:     gatewayFileURL,
		Expires: 30, // 30 seconds expiry
		Date:    time.Now().Unix(),
		Method:  "GET",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("%s/v3/files/sign", pinataAPIURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.jwt)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Pinata API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var accessResp accessLinkResponse
	if err := json.Unmarshal(body, &accessResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return accessResp.Data, nil
}

// fetchFromURL fetches content from a URL.
func (p *PinataClient) fetchFromURL(ctx context.Context, cid, url string) (*models.FileContent, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add gateway key header if provided and not using signed URL
	if p.gatewayKey != "" && !p.isPrivate {
		req.Header.Set("x-pinata-gateway-token", p.gatewayKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from gateway: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // CID not found
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gateway error (%d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	return &models.FileContent{
		CID:         cid,
		ContentType: contentType,
		Content:     string(body),
	}, nil
}
