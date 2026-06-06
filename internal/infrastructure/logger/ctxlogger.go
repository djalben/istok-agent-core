package logger

import (
	"context"
	"log/slog"

	"github.com/djalben/istok-agent-core/internal/ports"
)

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return ports.WithContextLogger(ctx, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	return ports.LoggerFromContext(ctx)
}

// rootLogger — root-logger процесса, задаётся из main на старте.
var rootLogger *slog.Logger

// SetRoot задаёт root-logger процесса (wire из main → infrastructure/logger).
func SetRoot(l *slog.Logger) {
	if l != nil {
		rootLogger = l
	}
}

// Root возвращает root-logger процесса или slog.Default(), если он ещё не задан.
// Сигнатура func() *slog.Logger совместима с ports.SetFallbackLogger.
func Root() *slog.Logger {
	if rootLogger == nil {
		return slog.Default()
	}

	return rootLogger
}
