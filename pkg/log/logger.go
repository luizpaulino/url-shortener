package log

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {
	// JSON handler no stdout
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}
