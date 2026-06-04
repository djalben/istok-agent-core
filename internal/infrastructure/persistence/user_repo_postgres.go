package persistence

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
)

// UserRepoPostgres implements ports.UserRepository on PostgreSQL.
type UserRepoPostgres struct {
	pool *pgxpool.Pool
}

func NewUserRepoPostgres(pool *pgxpool.Pool) *UserRepoPostgres {
	return &UserRepoPostgres{pool: pool}
}

var _ ports.UserRepository = (*UserRepoPostgres)(nil)

func (r *UserRepoPostgres) Create(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, display_name, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Email, u.PasswordHash, u.DisplayName, u.CreatedAt,
	)
	return err
}

func (r *UserRepoPostgres) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.scanOne(ctx,
		`SELECT id, email, password_hash, display_name, created_at FROM users WHERE email = $1`, email)
}

func (r *UserRepoPostgres) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return r.scanOne(ctx,
		`SELECT id, email, password_hash, display_name, created_at FROM users WHERE id = $1`, id)
}

func (r *UserRepoPostgres) scanOne(ctx context.Context, query, arg string) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx, query, arg).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
