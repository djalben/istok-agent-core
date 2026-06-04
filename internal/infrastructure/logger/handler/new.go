package handler

import (
	"io"
	"log/slog"
	"os"

	"github.com/djalben/istok-agent-core/internal/infrastructure/logger"
)

// Create — создание handler с выводом в stdout.
func Create(isPlain bool, level string) slog.Handler {
	return CreateWithWriter(isPlain, level, os.Stdout)
}

// CreateWithWriter — создание handler с произвольным io.Writer (tee для Watcher и т.п.).
func CreateWithWriter(isPlain bool, level string, writer io.Writer) slog.Handler {
	opts := &slog.HandlerOptions{Level: logger.GetLevel(level)}
	if isPlain {
		return slog.NewTextHandler(writer, opts)
	}

	return New(writer, opts)
}
