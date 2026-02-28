package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	Log = slog.New(handler)
}

// SetLevel sets the global log level dynamically.
func SetLevel(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	Log = slog.New(handler)
}
