package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// New creates a new structured logger based on configuration
func New(cfg Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// NewDefault creates a new logger with default configuration (JSON, Info level)
func NewDefault() *slog.Logger {
	return New(Config{
		Level:  "info",
		Format: "json",
	})
}

// With returns a new logger with additional attributes
func With(logger *slog.Logger, args ...any) *slog.Logger {
	return logger.With(args...)
}
