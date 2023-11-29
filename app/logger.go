package app

import (
	"os"
	"strconv"

	"log/slog"
)

func SetupLogger() {
	opts := &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		ReplaceAttr: cloudLoggingAttributes(),
	}
	if v, _ := strconv.ParseBool(os.Getenv("DEBUG")); v {
		opts.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}

var (
	cloudLoggingSeverityError   = "ERROR"
	cloudLoggingSeverityWarning = "WARNING"
	cloudLoggingSeverityInfo    = "INFO"
	cloudLoggingSeverityDebug   = "DEBUG"
)

func cloudLoggingAttributes() func([]string, slog.Attr) slog.Attr {
	levelToSeverity := map[string]string{
		"ERROR": cloudLoggingSeverityError,
		"WARN":  cloudLoggingSeverityWarning,
		"INFO":  cloudLoggingSeverityInfo,
		"DEBUG": cloudLoggingSeverityDebug,
	}
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "level" {
			a.Key = "severity"
			a.Value = slog.StringValue(levelToSeverity[a.Value.String()])
		}
		return a
	}
}
