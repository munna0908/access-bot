package logging

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog for structured logging.
type Logger struct {
	logger *slog.Logger
}

// NewLogger creates a new logger with the specified level.
func NewLogger(level string) *Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{logger: logger}
}

// RequestLogger provides logging context for a specific request.
type RequestLogger struct {
	logger        *slog.Logger
	requestID     string
	participantID string
	sessionID     string
}

// WithRequest creates a request-scoped logger.
func (l *Logger) WithRequest(requestID, participantID, sessionID string) *RequestLogger {
	return &RequestLogger{
		logger:        l.logger,
		requestID:     requestID,
		participantID: participantID,
		sessionID:     sessionID,
	}
}

// Info logs an info message with request context.
func (rl *RequestLogger) Info(ctx context.Context, msg string, attrs ...any) {
	baseAttrs := []any{
		"request_id", rl.requestID,
		"participant_id", rl.participantID,
		"session_id", rl.sessionID,
	}
	rl.logger.InfoContext(ctx, msg, append(baseAttrs, attrs...)...)
}

// Error logs an error message with request context.
func (rl *RequestLogger) Error(ctx context.Context, msg string, attrs ...any) {
	baseAttrs := []any{
		"request_id", rl.requestID,
		"participant_id", rl.participantID,
		"session_id", rl.sessionID,
	}
	rl.logger.ErrorContext(ctx, msg, append(baseAttrs, attrs...)...)
}

// Debug logs a debug message with request context.
func (rl *RequestLogger) Debug(ctx context.Context, msg string, attrs ...any) {
	baseAttrs := []any{
		"request_id", rl.requestID,
		"participant_id", rl.participantID,
		"session_id", rl.sessionID,
	}
	rl.logger.DebugContext(ctx, msg, append(baseAttrs, attrs...)...)
}

// Warn logs a warning message with request context.
func (rl *RequestLogger) Warn(ctx context.Context, msg string, attrs ...any) {
	baseAttrs := []any{
		"request_id", rl.requestID,
		"participant_id", rl.participantID,
		"session_id", rl.sessionID,
	}
	rl.logger.WarnContext(ctx, msg, append(baseAttrs, attrs...)...)
}

// LogRequestResult logs the final result of a request.
func (rl *RequestLogger) LogRequestResult(ctx context.Context, status string, usedCategories []string, reason *string) {
	attrs := []any{
		"status", status,
		"used_categories", usedCategories,
	}
	if reason != nil {
		attrs = append(attrs, "reason", *reason)
	}
	rl.Info(ctx, "request_completed", attrs...)
}

// Info logs an info message.
func (l *Logger) Info(ctx context.Context, msg string, attrs ...any) {
	l.logger.InfoContext(ctx, msg, attrs...)
}

// Error logs an error message.
func (l *Logger) Error(ctx context.Context, msg string, attrs ...any) {
	l.logger.ErrorContext(ctx, msg, attrs...)
}

// Debug logs a debug message.
func (l *Logger) Debug(ctx context.Context, msg string, attrs ...any) {
	l.logger.DebugContext(ctx, msg, attrs...)
}

// Warn logs a warning message.
func (l *Logger) Warn(ctx context.Context, msg string, attrs ...any) {
	l.logger.WarnContext(ctx, msg, attrs...)
}
