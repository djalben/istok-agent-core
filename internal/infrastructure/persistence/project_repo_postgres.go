package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// ProjectRepoPostgres implements ports.ProjectRepository on PostgreSQL.
type ProjectRepoPostgres struct {
	pool *pgxpool.Pool
}

func NewProjectRepoPostgres(pool *pgxpool.Pool) *ProjectRepoPostgres {
	return &ProjectRepoPostgres{pool: pool}
}

var _ ports.ProjectRepository = (*ProjectRepoPostgres)(nil)

// ListByOwner returns lightweight summaries (no files) ordered by recency.
func (r *ProjectRepoPostgres) ListByOwner(ctx context.Context, ownerID string) ([]*domain.Project, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_id, name, description, framework, is_public, slug, thumbnail_url, created_at, updated_at
		 FROM projects WHERE owner_id = $1 ORDER BY updated_at DESC`, ownerID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	defer rows.Close()

	var out []*domain.Project
	for rows.Next() {
		var p domain.Project
		err := rows.Scan(&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.Framework,
			&p.IsPublic, &p.Slug, &p.ThumbnailURL, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, wrapper.Wrap(err)
		}
		out = append(out, &p)
	}

	return out, wrapper.Wrap(rows.Err())
}

// GetByID returns the full project including files.
func (r *ProjectRepoPostgres) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	var p domain.Project
	var filesRaw []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner_id, name, description, framework, is_public, slug, thumbnail_url, prompt, files, created_at, updated_at
		 FROM projects WHERE id = $1`, id).
		Scan(&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.Framework, &p.IsPublic,
			&p.Slug, &p.ThumbnailURL, &p.Prompt, &filesRaw, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if len(filesRaw) > 0 {
		_ = json.Unmarshal(filesRaw, &p.Files)
	}
	if p.Files == nil {
		p.Files = map[string]string{}
	}

	return &p, nil
}

func (r *ProjectRepoPostgres) Create(ctx context.Context, p *domain.Project) error {
	filesJSON, err := json.Marshal(p.Files)
	if err != nil {
		return wrapper.Wrap(err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO projects
		   (id, owner_id, name, description, framework, is_public, slug, thumbnail_url, prompt, files, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		p.ID, p.OwnerID, p.Name, p.Description, p.Framework, p.IsPublic,
		p.Slug, p.ThumbnailURL, p.Prompt, filesJSON, p.CreatedAt, p.UpdatedAt)

	return wrapper.Wrap(err)
}

// Update persists mutable fields and bumps updated_at.
func (r *ProjectRepoPostgres) Update(ctx context.Context, p *domain.Project) error {
	filesJSON, err := json.Marshal(p.Files)
	if err != nil {
		return wrapper.Wrap(err)
	}
	p.UpdatedAt = time.Now().UTC()
	ct, err := r.pool.Exec(ctx,
		`UPDATE projects
		 SET name=$2, description=$3, framework=$4, is_public=$5, slug=$6, thumbnail_url=$7, prompt=$8, files=$9, updated_at=$10
		 WHERE id=$1`,
		p.ID, p.Name, p.Description, p.Framework, p.IsPublic, p.Slug, p.ThumbnailURL, p.Prompt, filesJSON, p.UpdatedAt)
	if err != nil {
		return wrapper.Wrap(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *ProjectRepoPostgres) Delete(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return wrapper.Wrap(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *ProjectRepoPostgres) CountByOwner(ctx context.Context, ownerID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM projects WHERE owner_id = $1`, ownerID).Scan(&n)

	return n, wrapper.Wrap(err)
}

func (r *ProjectRepoPostgres) CountPublishedByOwner(ctx context.Context, ownerID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM projects WHERE owner_id = $1 AND is_public = TRUE`, ownerID).Scan(&n)

	return n, wrapper.Wrap(err)
}
