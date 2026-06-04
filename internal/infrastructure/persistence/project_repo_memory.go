package persistence

import (
	"context"
	"maps"
	"sort"
	"sync"
	"time"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
)

// ProjectRepoMemory — потокобезопасная in-memory реализация ProjectRepository.
// Используется как fallback, когда DATABASE_URL не задан.
type ProjectRepoMemory struct {
	mu    sync.RWMutex
	items map[string]*domain.Project
}

func NewProjectRepoMemory() *ProjectRepoMemory {
	return &ProjectRepoMemory{items: make(map[string]*domain.Project)}
}

var _ ports.ProjectRepository = (*ProjectRepoMemory)(nil)

func clone(p *domain.Project) *domain.Project {
	cp := *p
	if p.Files != nil {
		cp.Files = make(map[string]string, len(p.Files))
		maps.Copy(cp.Files, p.Files)
	}

	return &cp
}

func (r *ProjectRepoMemory) ListByOwner(_ context.Context, ownerID string) ([]*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Project
	for _, p := range r.items {
		if p.OwnerID == ownerID {
			c := clone(p)
			c.Files = nil // summaries omit files
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })

	return out, nil
}

func (r *ProjectRepoMemory) GetByID(_ context.Context, id string) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}

	return clone(p), nil
}

func (r *ProjectRepoMemory) Create(_ context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[p.ID] = clone(p)

	return nil
}

func (r *ProjectRepoMemory) Update(_ context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[p.ID]; !ok {
		return domain.ErrNotFound
	}
	p.UpdatedAt = time.Now().UTC()
	r.items[p.ID] = clone(p)

	return nil
}

func (r *ProjectRepoMemory) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)

	return nil
}

func (r *ProjectRepoMemory) CountByOwner(_ context.Context, ownerID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, p := range r.items {
		if p.OwnerID == ownerID {
			n++
		}
	}

	return n, nil
}

func (r *ProjectRepoMemory) CountPublishedByOwner(_ context.Context, ownerID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, p := range r.items {
		if p.OwnerID == ownerID && p.IsPublic {
			n++
		}
	}

	return n, nil
}
