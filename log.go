package grdep

import (
	"io"
	"log/slog"
)

func NewLogger(w io.Writer, level slog.Leveler) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
	}))
	defaultLogger = logger
	defaultLogLevel = level
	return logger
}

var (
	defaultLogger   *slog.Logger
	defaultLogLevel slog.Leveler
)

func IsDebug() bool {
	return defaultLogLevel == slog.LevelDebug
}

func L() *slog.Logger {
	return defaultLogger
}

func OnDebug(f func()) {
	if IsDebug() {
		f()
	}
}
