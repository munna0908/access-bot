package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/access-bot/internal/providers"
)

const (
	geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"
)

// GeminiClient implements the LLMProvider interface using the Google Gemini API.
type GeminiClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
}

// Ensure GeminiClient implements LLMProvider.
var _ providers.LLMProvider = (*GeminiClient)(nil)

// NewGeminiClient creates a new Gemini API client.
func NewGeminiClient(apiKey, model string, timeout time.Duration) *GeminiClient {
	return &GeminiClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		apiKey: apiKey,
		model:  model,
	}
}

// geminiRequest represents the request body for the Gemini API.
type geminiRequest struct {
	Contents       []geminiContent       `json:"contents"`
	SystemInstruct *geminiSystemInstruct `json:"systemInstruction,omitempty"`
	GenConfig      *geminiGenConfig      `json:"generationConfig,omitempty"`
}

type geminiSystemInstruct struct {
	Parts []geminiPart `json:"parts"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

// geminiResponse represents the response from the Gemini API.
type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *geminiError      `json:"error,omitempty"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Complete sends a prompt to Gemini and returns the response.
func (g *GeminiClient) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if g.apiKey == "" {
		return "", fmt.Errorf("gemini API key not configured")
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: userPrompt},
				},
			},
		},
		GenConfig: &geminiGenConfig{
			MaxOutputTokens: 4096,
			Temperature:     0.1,
		},
	}

	// Add system instruction if provided
	if systemPrompt != "" {
		reqBody.SystemInstruct = &geminiSystemInstruct{
			Parts: []geminiPart{
				{Text: systemPrompt},
			},
		}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf(geminiAPIURL, g.model) + "?key=" + g.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp geminiResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
			return "", fmt.Errorf("API error (%d): %s - %s", resp.StatusCode, errResp.Error.Status, errResp.Error.Message)
		}
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var apiResp geminiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return "", nil
	}

	// Extract text from response
	var result string
	for _, part := range apiResp.Candidates[0].Content.Parts {
		result += part.Text
	}

	return result, nil
}

// GetModelName returns the model name.
func (g *GeminiClient) GetModelName() string {
	return g.model
}
