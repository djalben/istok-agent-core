package application

import (
	"context"
	"log/slog"

	"github.com/djalben/istok-agent-core/internal/ports"
)

func applog(ctx context.Context) *slog.Logger {
	return ports.LoggerFromContext(ctx)
}
