package http

import (
	"context"
	"log/slog"
)

func logSSERequestMeta(ctx context.Context) {
	logFrom(ctx).InfoContext(ctx, "sse request arrived")
}

func logRateLimitExceeded(ctx context.Context) {
	logFrom(ctx).WarnContext(ctx, "rate limit exceeded")
}

func logConcurrencyLimitReached(ctx context.Context) {
	logFrom(ctx).WarnContext(ctx, "concurrency limit reached")
}

func sseLog(ctx context.Context) *slog.Logger {
	return logFrom(ctx)
}
