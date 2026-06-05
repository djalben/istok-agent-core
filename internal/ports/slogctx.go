package ports

import (
	"context"
	"log/slog"
)

type slogCtxKey struct{}

var fallbackRoot = func() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

// SetFallbackLogger задаёт провайдер root-logger для LoggerFromContext (wire из main → infrastructure/logger).
func SetFallbackLogger(fn func() *slog.Logger) {
	if fn != nil {
		fallbackRoot = fn
	}
}

// WithContextLogger сохраняет request-scoped *slog.Logger в context.
func WithContextLogger(ctx context.Context, l *slog.Logger) context.Context {
	if l == nil {
		return ctx
	}

	return context.WithValue(ctx, slogCtxKey{}, l)
}

// LoggerFromContext возвращает logger из context или fallback root.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return fallbackRoot()
	}

	if l, ok := ctx.Value(slogCtxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}

	return fallbackRoot()
}
