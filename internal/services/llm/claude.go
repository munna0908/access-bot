package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/access-bot/internal/providers"
)

const (
	anthropicAPIURL     = "https://api.anthropic.com/v1/messages"
	anthropicAPIVersion = "2023-06-01"
	maxTokens           = 1024
)

// ClaudeClient implements the LLMProvider interface using the Anthropic API.
type ClaudeClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
}

// Ensure ClaudeClient implements LLMProvider.
var _ providers.LLMProvider = (*ClaudeClient)(nil)

// NewClaudeClient creates a new Claude API client.
func NewClaudeClient(apiKey, model string, timeout time.Duration) *ClaudeClient {
	return &ClaudeClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		apiKey: apiKey,
		model:  model,
	}
}

// anthropicRequest represents the request body for the Anthropic API.
type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse represents the response from the Anthropic API.
type anthropicResponse struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Role         string          `json:"role"`
	Content      []contentBlock  `json:"content"`
	Model        string          `json:"model"`
	StopReason   string          `json:"stop_reason"`
	StopSequence *string         `json:"stop_sequence"`
	Usage        anthropicUsage  `json:"usage"`
	Error        *anthropicError `json:"error,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Complete sends a prompt to Claude and returns the response.
func (c *ClaudeClient) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("anthropic API key not configured")
	}

	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages: []message{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp anthropicResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
			return "", fmt.Errorf("API error (%d): %s - %s", resp.StatusCode, errResp.Error.Type, errResp.Error.Message)
		}
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", nil
	}

	// Extract text from content blocks
	var result string
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			result += block.Text
		}
	}

	return result, nil
}

// GetModelName returns the model name.
func (c *ClaudeClient) GetModelName() string {
	return c.model
}
