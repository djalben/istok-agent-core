package http

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Auth Handler
//  JWT-based аутентификация без внешних зависимостей
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// User представляет пользователя в системе
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // не отправляем в JSON
	DisplayName  string    `json:"display_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	users     map[string]*User // email -> User
	jwtSecret []byte
	mu        sync.RWMutex
}

// Claims для JWT токена
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// SignupRequest запрос на регистрацию
type SignupRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// LoginRequest запрос на вход
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse ответ с токеном
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// NewAuthHandler создает новый handler аутентификации
func NewAuthHandler() *AuthHandler {
	// Генерируем случайный JWT secret
	secret := make([]byte, 32)
	rand.Read(secret)

	return &AuthHandler{
		users:     make(map[string]*User),
		jwtSecret: secret,
	}
}

// HandleSignup обрабатывает POST /api/v1/auth/signup
func (h *AuthHandler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	// Валидация
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Email и пароль обязательны")
		return
	}

	if len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "Пароль должен быть не менее 6 символов")
		return
	}

	if !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "Неверный формат email")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Проверяем, существует ли пользователь
	if _, exists := h.users[req.Email]; exists {
		writeError(w, http.StatusConflict, "Пользователь с таким email уже существует")
		return
	}

	// Хешируем пароль
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка создания пользователя")
		return
	}

	// Создаем пользователя
	user := &User{
		ID:           generateID(),
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		DisplayName:  req.DisplayName,
		CreatedAt:    time.Now(),
	}

	if user.DisplayName == "" {
		user.DisplayName = strings.Split(req.Email, "@")[0]
	}

	h.users[req.Email] = user

	// Генерируем JWT токен
	token, err := h.generateToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	// Возвращаем токен и данные пользователя
	response := AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleLogin обрабатывает POST /api/v1/auth/login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	// Валидация
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Email и пароль обязательны")
		return
	}

	h.mu.RLock()
	user, exists := h.users[req.Email]
	h.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}

	// Генерируем JWT токен
	token, err := h.generateToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	// Возвращаем токен и данные пользователя
	response := AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleMe обрабатывает GET /api/v1/auth/me (получение текущего пользователя)
func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем токен из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "Токен не предоставлен")
		return
	}

	// Формат: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		writeError(w, http.StatusUnauthorized, "Неверный формат токена")
		return
	}

	tokenString := parts[1]

	// Парсим и валидируем токен
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи")
		}
		return h.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		writeError(w, http.StatusUnauthorized, "Неверный токен")
		return
	}

	// Получаем пользователя
	h.mu.RLock()
	user, exists := h.users[claims.Email]
	h.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusUnauthorized, "Пользователь не найден")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// generateToken генерирует JWT токен для пользователя
func (h *AuthHandler) generateToken(user *User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour * 7) // 7 дней

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "istok-agent",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

// generateID генерирует случайный ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
