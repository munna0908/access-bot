package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/access-bot/internal/config"
	"github.com/access-bot/internal/handlers"
	"github.com/access-bot/internal/logging"
	"github.com/access-bot/internal/mocks"
	"github.com/access-bot/internal/providers"
	"github.com/access-bot/internal/services/accesslayer"
	"github.com/access-bot/internal/services/categoryindex"
	"github.com/access-bot/internal/services/filesystem"
	"github.com/access-bot/internal/services/llm"
	"github.com/access-bot/internal/services/session"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Create logger
	logger := logging.NewLogger(cfg.LogLevel)

	ctx := context.Background()

	// Print banner
	printBanner()

	// Create providers — demo mode serves common sample data for any participant
	var sessionProvider providers.SessionProvider
	var categoryIndexProvider providers.CategoryIndexProvider
	var filesystemProvider providers.FilesystemProvider
	var llmProvider providers.LLMProvider

	if cfg.DemoMode {
		logger.Info(ctx, "demo_mode_enabled", "note", "accepting any session/participant")
		sessionProvider = mocks.NewDemoSessionProvider()
		categoryIndexProvider = createCategoryIndexProvider(ctx, cfg, logger)
		filesystemProvider = createFilesystemProvider(ctx, cfg, logger)
		llmProvider = createLLMProvider(ctx, cfg, logger)
	} else {
		sessionProvider = createSessionProvider(ctx, cfg, logger)
		categoryIndexProvider = createCategoryIndexProvider(ctx, cfg, logger)
		filesystemProvider = createFilesystemProvider(ctx, cfg, logger)
		llmProvider = createLLMProvider(ctx, cfg, logger)
	}

	// Create service
	service := accesslayer.NewService(
		sessionProvider,
		categoryIndexProvider,
		filesystemProvider,
		llmProvider,
		logger,
	)

	// Create router
	router := handlers.NewRouter(service, logger)

	// Create server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router.Handler(),
		ReadTimeout:  cfg.Timeouts.ServerRead,
		WriteTimeout: cfg.Timeouts.ServerWrite,
	}

	// Start server in goroutine
	go func() {
		logger.Info(ctx, "server_starting", "address", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "server_error", "error", err.Error())
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "server_shutting_down")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "server_shutdown_error", "error", err.Error())
		os.Exit(1)
	}

	logger.Info(ctx, "server_stopped")
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                   ACCESS LAYER SERVICE                    ║
║                         v1.0.0                            ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func createLLMProvider(ctx context.Context, cfg *config.Config, logger *logging.Logger) providers.LLMProvider {
	provider := cfg.LLMProvider

	// Auto-detect provider based on available API keys
	if provider == "auto" {
		if cfg.Claude.APIKey != "" {
			provider = "claude"
		} else if cfg.Gemini.APIKey != "" {
			provider = "gemini"
		} else {
			provider = "mock"
		}
	}

	switch provider {
	case "claude":
		logger.Info(ctx, "using_llm_provider", "provider", "claude", "model", cfg.Claude.Model)
		return llm.NewClaudeClient(cfg.Claude.APIKey, cfg.Claude.Model, cfg.Timeouts.HTTPClient)
	case "gemini":
		logger.Info(ctx, "using_llm_provider", "provider", "gemini", "model", cfg.Gemini.Model)
		return llm.NewGeminiClient(cfg.Gemini.APIKey, cfg.Gemini.Model, cfg.Timeouts.HTTPClient)
	default:
		logger.Info(ctx, "using_llm_provider", "provider", "mock")
		return mocks.NewMockLLMProvider()
	}
}

func createSessionProvider(ctx context.Context, cfg *config.Config, logger *logging.Logger) providers.SessionProvider {
	provider := cfg.SessionProvider

	// Auto-detect provider based on available configuration
	if provider == "auto" {
		if cfg.Intelligence.BaseURL != "" {
			provider = "intelligence"
		} else {
			provider = "mock"
		}
	}

	switch provider {
	case "intelligence":
		logger.Info(ctx, "using_session_provider", "provider", "intelligence", "url", cfg.Intelligence.BaseURL)
		return session.NewIntelligenceClient(session.IntelligenceConfig{
			BaseURL: cfg.Intelligence.BaseURL,
			Timeout: cfg.Timeouts.HTTPClient,
		})
	default:
		logger.Info(ctx, "using_session_provider", "provider", "mock")
		return mocks.NewMockSessionProvider()
	}
}

func createCategoryIndexProvider(ctx context.Context, cfg *config.Config, logger *logging.Logger) providers.CategoryIndexProvider {
	// Prefer env-var CIDs — works without on-chain registration
	if cfg.CategoryCIDs.IsConfigured() {
		logger.Info(ctx, "using_category_index_provider", "provider", "env_cids",
			"food", cfg.CategoryCIDs.Food != "",
			"health", cfg.CategoryCIDs.Health != "",
			"address", cfg.CategoryCIDs.Address != "",
			"payment", cfg.CategoryCIDs.Payment != "",
		)
		return mocks.NewCIDCategoryIndexProvider(
			cfg.CategoryCIDs.Food,
			cfg.CategoryCIDs.Health,
			cfg.CategoryCIDs.Address,
			cfg.CategoryCIDs.Payment,
		)
	}
	// Fall back to on-chain lookup via intelligence service
	if cfg.Intelligence.BaseURL != "" {
		logger.Info(ctx, "using_category_index_provider", "provider", "intelligence", "url", cfg.Intelligence.BaseURL)
		return categoryindex.NewIntelligenceClient(categoryindex.IntelligenceConfig{
			BaseURL: cfg.Intelligence.BaseURL,
			Timeout: cfg.Timeouts.HTTPClient,
		})
	}
	logger.Info(ctx, "using_category_index_provider", "provider", "mock")
	return mocks.NewMockCategoryIndexProvider()
}

func createFilesystemProvider(ctx context.Context, cfg *config.Config, logger *logging.Logger) providers.FilesystemProvider {
	provider := cfg.FilesystemProvider

	// Auto-detect provider based on available configuration
	if provider == "auto" {
		if cfg.Pinata.GatewayURL != "" || cfg.Pinata.JWT != "" {
			provider = "pinata"
		} else {
			provider = "mock"
		}
	}

	switch provider {
	case "pinata":
		logger.Info(ctx, "using_filesystem_provider", "provider", "pinata", "gateway", cfg.Pinata.GatewayURL)
		return filesystem.NewPinataClient(filesystem.PinataConfig{
			GatewayURL: cfg.Pinata.GatewayURL,
			GatewayKey: cfg.Pinata.GatewayKey,
			JWT:        cfg.Pinata.JWT,
			IsPrivate:  cfg.Pinata.IsPrivate,
			Timeout:    cfg.Timeouts.HTTPClient,
		})
	default:
		logger.Info(ctx, "using_filesystem_provider", "provider", "mock")
		return mocks.NewMockFilesystemProvider()
	}
}
