package app

import (
	"os"

	"log/slog"
)

func SetupLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
}
