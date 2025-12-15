package logger

import (
	"log/slog"
	"os"
)

// New returns a structured slog.Logger configured for JSON output.
func New(service string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				return slog.Attr{Key: "level", Value: a.Value}
			}
			return a
		},
	})
	return slog.New(handler).With("service", service)
}



