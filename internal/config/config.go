package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Server             ServerConfig
	LLMProvider        string // "claude", "gemini", or "mock"
	FilesystemProvider string // "pinata" or "mock"
	Claude             ClaudeConfig
	Gemini             GeminiConfig
	Pinata             PinataConfig
	Timeouts           TimeoutConfig
	LogLevel           string
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host string
	Port string
}

// ClaudeConfig holds Anthropic API configuration.
type ClaudeConfig struct {
	APIKey string
	Model  string
}

// GeminiConfig holds Google Gemini API configuration.
type GeminiConfig struct {
	APIKey string
	Model  string
}

// PinataConfig holds Pinata IPFS configuration.
type PinataConfig struct {
	GatewayURL string // Custom gateway URL (leave empty for public gateway)
	GatewayKey string // Optional: Gateway key for dedicated gateways
	JWT        string // Optional: Pinata JWT for private file access
	IsPrivate  bool   // Whether files are private (requires JWT)
}

// TimeoutConfig holds various timeout configurations.
type TimeoutConfig struct {
	HTTPClient  time.Duration
	ServerRead  time.Duration
	ServerWrite time.Duration
}

// CategoryScopeMapping maps participant data categories to required OAuth scopes.
var CategoryScopeMapping = map[string]string{
	"HEALTH":  "health.read",
	"FOOD":    "preferences.food.read",
	"ADDRESS": "profile.address.read",
	"PAYMENT": "finance.payment.read",
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("HOST", "localhost"),
			Port: getEnv("PORT", "8080"),
		},
		LLMProvider:        getEnv("LLM_PROVIDER", "auto"),        // "claude", "gemini", "mock", or "auto"
		FilesystemProvider: getEnv("FILESYSTEM_PROVIDER", "auto"), // "pinata", "mock", or "auto"
		Claude: ClaudeConfig{
			APIKey: getEnv("ANTHROPIC_API_KEY", ""),
			Model:  getEnv("CLAUDE_MODEL", "claude-sonnet-4-20250514"),
		},
		Gemini: GeminiConfig{
			APIKey: getEnv("GEMINI_API_KEY", ""),
			Model:  getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		},
		Pinata: PinataConfig{
			GatewayURL: getEnv("PINATA_GATEWAY_URL", ""),
			GatewayKey: getEnv("PINATA_GATEWAY_KEY", ""),
			JWT:        getEnv("PINATA_JWT", ""),
			IsPrivate:  getEnv("PINATA_PRIVATE", "false") == "true",
		},
		Timeouts: TimeoutConfig{
			HTTPClient:  getDurationEnv("HTTP_CLIENT_TIMEOUT", 30),
			ServerRead:  getDurationEnv("SERVER_READ_TIMEOUT", 10),
			ServerWrite: getDurationEnv("SERVER_WRITE_TIMEOUT", 30),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultSeconds int) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return time.Duration(defaultSeconds) * time.Second
}

// GetRequiredScope returns the OAuth scope required for a given category.
func GetRequiredScope(category string) (string, bool) {
	scope, ok := CategoryScopeMapping[category]
	return scope, ok
}

// GetAllValidCategories returns all supported category names.
func GetAllValidCategories() []string {
	categories := make([]string, 0, len(CategoryScopeMapping))
	for cat := range CategoryScopeMapping {
		categories = append(categories, cat)
	}
	return categories
}
