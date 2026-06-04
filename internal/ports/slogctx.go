package ports

import (
	"context"
	"log/slog"
)

type slogCtxKey struct{}

// WithContextLogger сохраняет request-scoped *slog.Logger в context.
func WithContextLogger(ctx context.Context, l *slog.Logger) context.Context {
	if l == nil {
		return ctx
	}

	return context.WithValue(ctx, slogCtxKey{}, l)
}

// LoggerFromContext возвращает logger из context или slog.Default().
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}

	if l, ok := ctx.Value(slogCtxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}

	return slog.Default()
}
