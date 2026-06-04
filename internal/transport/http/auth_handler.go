package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"github.com/djalben/istok-agent-core/internal/domain"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Auth Handler (delivery)
//  Тонкий слой: парсинг HTTP ↔ вызов AuthService.
//  Вся бизнес-логика — в application/usecases/auth_service.go.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type AuthHandler struct {
	auth *usecases.AuthService
}

func NewAuthHandler(auth *usecases.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type signupRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

// HandleSignup — POST /api/v1/auth/signup
func (h *AuthHandler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	user, token, err := h.auth.Signup(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		if errors.Is(err, domain.ErrEmailExists) {
			writeError(w, http.StatusConflict, "Пользователь с таким email уже существует")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = writeJSON(w, http.StatusOK, authResponse{Token: token, User: user})
}

// HandleLogin — POST /api/v1/auth/login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	user, token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}
	_ = writeJSON(w, http.StatusOK, authResponse{Token: token, User: user})
}

// HandleMe — GET /api/v1/auth/me (за AuthMiddleware).
func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	user, err := h.auth.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Пользователь не найден")
		return
	}
	_ = writeJSON(w, http.StatusOK, user)
}
