package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
)

// AuthClaims — полезная нагрузка JWT.
type AuthClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthService — бизнес-логика аутентификации (signup/login/verify).
// Зависит только от порта UserRepository (Dependency Rule).
type AuthService struct {
	users     ports.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

// NewAuthService. Если jwtSecret пуст — генерируем случайный (dev-режим;
// токены не переживут перезапуск). В проде задавайте JWT_SECRET.
func NewAuthService(users ports.UserRepository, jwtSecret string) *AuthService {
	secret := jwtSecret
	if secret == "" {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		secret = hex.EncodeToString(b)
	}
	return &AuthService{
		users:     users,
		jwtSecret: []byte(secret),
		tokenTTL:  7 * 24 * time.Hour,
	}
}

// Signup создаёт пользователя и возвращает его вместе с JWT.
func (s *AuthService) Signup(ctx context.Context, email, password, displayName string) (*domain.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, "", errors.New("неверный формат email")
	}
	if len(password) < 6 {
		return nil, "", errors.New("пароль должен быть не менее 6 символов")
	}
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, "", domain.ErrEmailExists
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	if displayName == "" {
		displayName = strings.Split(email, "@")[0]
	}
	u := &domain.User{
		ID:           domain.GenerateID(),
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  displayName,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, "", err
	}
	token, err := s.issue(u)
	return u, token, err
}

// Login проверяет пароль и возвращает пользователя + JWT.
func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", domain.ErrInvalidCreds
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, "", domain.ErrInvalidCreds
	}
	token, err := s.issue(u)
	return u, token, err
}

// GetByID возвращает пользователя по id (для /auth/me и /user/profile).
func (s *AuthService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.users.FindByID(ctx, id)
}

// Verify валидирует JWT и возвращает claims.
func (s *AuthService) Verify(tokenString string) (*AuthClaims, error) {
	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}
	return claims, nil
}

func (s *AuthService) issue(u *domain.User) (string, error) {
	now := time.Now()
	claims := &AuthClaims{
		UserID: u.ID,
		Email:  u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "istok-agent",
			Subject:   u.ID,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}
