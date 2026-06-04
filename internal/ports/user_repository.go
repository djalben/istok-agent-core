package ports

import (
	"context"

	"github.com/djalben/istok-agent-core/internal/domain"
)

// UserRepository — контракт хранилища пользователей.
// Реализации: Postgres (infrastructure/persistence) и in-memory fallback.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
}
