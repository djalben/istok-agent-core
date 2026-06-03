package persistence

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// Postgres holds the connection pool and is the entry point for the
// Postgres-backed repositories.
type Postgres struct {
	Pool *pgxpool.Pool
}

// NewPostgres connects to the DB, verifies the connection, and runs the
// idempotent migration (schema.sql) on startup.
func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Postgres{Pool: pool}, nil
}

// Close releases the connection pool.
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
