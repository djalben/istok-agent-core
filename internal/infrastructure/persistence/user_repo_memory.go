package persistence

import (
	"context"
	"strings"
	"sync"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
)

// UserRepoMemory — потокобезопасная in-memory реализация UserRepository.
// Используется как fallback, когда DATABASE_URL не задан.
type UserRepoMemory struct {
	mu      sync.RWMutex
	byID    map[string]*domain.User
	byEmail map[string]*domain.User
}

func NewUserRepoMemory() *UserRepoMemory {
	return &UserRepoMemory{
		byID:    make(map[string]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

var _ ports.UserRepository = (*UserRepoMemory)(nil)

func (r *UserRepoMemory) Create(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := strings.ToLower(u.Email)
	if _, exists := r.byEmail[key]; exists {
		return domain.ErrEmailExists
	}
	cp := *u
	r.byID[u.ID] = &cp
	r.byEmail[key] = &cp

	return nil
}

func (r *UserRepoMemory) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byEmail[strings.ToLower(email)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u

	return &cp, nil
}

func (r *UserRepoMemory) FindByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u

	return &cp, nil
}
