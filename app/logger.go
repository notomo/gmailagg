package app

import (
	"os"
	"strconv"

	"log/slog"
)

func SetupLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if v, _ := strconv.ParseBool(os.Getenv("DEBUG")); v {
		opts.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}
