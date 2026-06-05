package ports

import (
	"context"
	"log/slog"
	"sync/atomic"
)

type slogCtxKey struct{}

var rootLogger atomic.Pointer[slog.Logger]

// discardLogger — no-op fallback до SetRootLogger (не slog.Default).
var discardLogger = slog.New(slog.DiscardHandler)

// SetRootLogger устанавливает application-wide logger (вызывается из cmd/server при старте).
func SetRootLogger(l *slog.Logger) {
	if l == nil {
		return
	}

	rootLogger.Store(l)
}

// RootLogger возвращает root logger, установленный через SetRootLogger.
func RootLogger() *slog.Logger {
	if l := rootLogger.Load(); l != nil {
		return l
	}

	return discardLogger
}

// WithContextLogger сохраняет request-scoped *slog.Logger в context.
func WithContextLogger(ctx context.Context, l *slog.Logger) context.Context {
	if l == nil {
		return ctx
	}

	return context.WithValue(ctx, slogCtxKey{}, l)
}

// LoggerFromContext возвращает logger из context или RootLogger().
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return RootLogger()
	}

	if l, ok := ctx.Value(slogCtxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}

	return RootLogger()
}
