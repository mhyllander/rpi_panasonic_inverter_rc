package utils

import (
	"log/slog"
	"os"
)

func SetLoggerOpts(level string) *slog.HandlerOptions {
	var opts slog.HandlerOptions = slog.HandlerOptions{}
	switch level {
	case "debug":
		opts.Level = slog.LevelDebug
	case "info":
		opts.Level = slog.LevelInfo
	case "warn":
		opts.Level = slog.LevelWarn
	case "error":
		opts.Level = slog.LevelError
	default:
		opts.Level = slog.LevelInfo
	}
	return &opts
}

func InitLogger(logLevel string) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, SetLoggerOpts(logLevel)))
	slog.SetDefault(logger)
}
