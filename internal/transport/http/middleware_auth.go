package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
)

type ctxKey string

const userIDKey ctxKey = "userID"

// AuthMiddleware валидирует Bearer JWT и кладёт userID в контекст запроса.
// Оборачивает защищённые handler'ы (см. server.go).
func AuthMiddleware(auth *usecases.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "Токен не предоставлен")
			return
		}
		claims, err := auth.Verify(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Неверный или истёкший токен")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next(w, r.WithContext(ctx))
	}
}

// bearerToken извлекает токен из заголовка "Authorization: Bearer <token>".
func bearerToken(r *http.Request) (string, bool) {
	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

// userIDFromContext возвращает id аутентифицированного пользователя.
func userIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}
