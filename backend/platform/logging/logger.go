package logging

import (
	"log/slog"
	"os"
	"strings"
)

// NewLogger creates a new structured logger using slog.
// Level is parsed from string: "debug", "info", "warn", "error".
func NewLogger(level string) *slog.Logger {
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: logLevel == slog.LevelDebug,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}
