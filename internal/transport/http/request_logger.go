package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/djalben/istok-agent-core/internal/ports"
)

const requestIDKey ctxKey = "requestID"

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}

func requestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)

	return id
}

func logFrom(ctx context.Context) *slog.Logger {
	return ports.LoggerFromContext(ctx)
}

// requestLoggerMiddleware прокидывает request-scoped logger в context (как xplr).
func (s *Server) requestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := newRequestID()
		l := slog.Default().With(
			"request_id", reqID,
			"method", r.Method,
			"path", r.URL.Path,
		)
		ctx := ports.WithContextLogger(r.Context(), l)
		ctx = context.WithValue(ctx, requestIDKey, reqID)
		r = r.WithContext(ctx)

		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		attrs := []any{
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_ip", clientIP(r),
		}
		if uid, ok := userIDFromContext(ctx); ok {
			attrs = append(attrs, "user_id", uid)
		}

		l.InfoContext(ctx, "http request completed", attrs...)
		if wrapped.statusCode >= 400 {
			l.ErrorContext(ctx, "http request failed", attrs...)
		}
	})
}
