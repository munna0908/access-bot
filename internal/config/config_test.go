package config

import (
	"os"
	"testing"
)

func TestGetRequiredScope(t *testing.T) {
	tests := []struct {
		category  string
		wantScope string
		wantOk    bool
	}{
		{"HEALTH", "health.read", true},
		{"FOOD", "preferences.food.read", true},
		{"ADDRESS", "profile.address.read", true},
		{"PAYMENT", "finance.payment.read", true},
		{"INVALID", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			scope, ok := GetRequiredScope(tt.category)
			if ok != tt.wantOk {
				t.Errorf("GetRequiredScope(%q) ok = %v, want %v", tt.category, ok, tt.wantOk)
			}
			if scope != tt.wantScope {
				t.Errorf("GetRequiredScope(%q) = %q, want %q", tt.category, scope, tt.wantScope)
			}
		})
	}
}

func TestGetAllValidCategories(t *testing.T) {
	categories := GetAllValidCategories()
	if len(categories) != 4 {
		t.Errorf("expected 4 categories, got %d", len(categories))
	}

	expected := map[string]bool{
		"HEALTH":  true,
		"FOOD":    true,
		"ADDRESS": true,
		"PAYMENT": true,
	}

	for _, cat := range categories {
		if !expected[cat] {
			t.Errorf("unexpected category: %s", cat)
		}
	}
}

func TestLoad(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("PORT", "9090")
	os.Setenv("HOST", "0.0.0.0")
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	os.Setenv("CLAUDE_MODEL", "claude-test")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("CLAUDE_MODEL")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg := Load()

	if cfg.Server.Port != "9090" {
		t.Errorf("expected Port=9090, got %s", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected Host=0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Claude.APIKey != "test-key" {
		t.Errorf("expected APIKey=test-key, got %s", cfg.Claude.APIKey)
	}
	if cfg.Claude.Model != "claude-test" {
		t.Errorf("expected Model=claude-test, got %s", cfg.Claude.Model)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel=debug, got %s", cfg.LogLevel)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Clear relevant environment variables
	os.Unsetenv("PORT")
	os.Unsetenv("HOST")
	os.Unsetenv("LOG_LEVEL")

	cfg := Load()

	if cfg.Server.Port != "8080" {
		t.Errorf("expected default Port=8080, got %s", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("expected default Host=localhost, got %s", cfg.Server.Host)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default LogLevel=info, got %s", cfg.LogLevel)
	}
}
