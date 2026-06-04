package usecases

import (
	"context"
	"maps"
	"sort"
	"sync"
	"time"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
)

// memoryProjectRepo — in-memory ProjectRepository для unit-тестов usecases (без infrastructure).
type memoryProjectRepo struct {
	mu    sync.RWMutex
	items map[string]*domain.Project
}

func newMemoryProjectRepo() ports.ProjectRepository {
	return &memoryProjectRepo{items: make(map[string]*domain.Project)}
}

func cloneProject(p *domain.Project) *domain.Project {
	cp := *p
	if p.Files != nil {
		cp.Files = make(map[string]string, len(p.Files))
		maps.Copy(cp.Files, p.Files)
	}

	return &cp
}

func (r *memoryProjectRepo) ListByOwner(_ context.Context, ownerID string) ([]*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Project
	for _, p := range r.items {
		if p.OwnerID == ownerID {
			c := cloneProject(p)
			c.Files = nil
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })

	return out, nil
}

func (r *memoryProjectRepo) GetByID(_ context.Context, id string) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}

	return cloneProject(p), nil
}

func (r *memoryProjectRepo) Create(_ context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[p.ID] = cloneProject(p)

	return nil
}

func (r *memoryProjectRepo) Update(_ context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[p.ID]; !ok {
		return domain.ErrNotFound
	}
	p.UpdatedAt = time.Now().UTC()
	r.items[p.ID] = cloneProject(p)

	return nil
}

func (r *memoryProjectRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)

	return nil
}

func (r *memoryProjectRepo) CountByOwner(_ context.Context, ownerID string) (int, error) {
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

func (r *memoryProjectRepo) CountPublishedByOwner(_ context.Context, ownerID string) (int, error) {
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
