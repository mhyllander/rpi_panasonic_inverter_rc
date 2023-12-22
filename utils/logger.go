package utils

import "log/slog"

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
