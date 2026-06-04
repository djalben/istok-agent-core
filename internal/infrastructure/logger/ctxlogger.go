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
