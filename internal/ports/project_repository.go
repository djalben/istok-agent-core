package ports

import (
	"context"

	"github.com/istok/agent-core/internal/domain"
)

// ProjectRepository — контракт хранилища проектов.
// List/Get/Count всегда скоупятся по ownerID (изоляция пользователей).
type ProjectRepository interface {
	ListByOwner(ctx context.Context, ownerID string) ([]*domain.Project, error)
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	Create(ctx context.Context, p *domain.Project) error
	Update(ctx context.Context, p *domain.Project) error
	Delete(ctx context.Context, id string) error

	// Aggregates for the profile page.
	CountByOwner(ctx context.Context, ownerID string) (int, error)
	CountPublishedByOwner(ctx context.Context, ownerID string) (int, error)
}
