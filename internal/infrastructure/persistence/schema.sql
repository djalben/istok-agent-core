-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
--  Istok Agent Core — Layer 1 schema (PostgreSQL)
--  Idempotent: safe to run on every startup.
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name  TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS projects (
    id            TEXT PRIMARY KEY,
    owner_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    framework     TEXT NOT NULL DEFAULT '',
    is_public     BOOLEAN NOT NULL DEFAULT FALSE,
    slug          TEXT,
    thumbnail_url TEXT,
    prompt        TEXT NOT NULL DEFAULT '',
    files         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_projects_owner_id   ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_updated_at ON projects(updated_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_slug ON projects(slug) WHERE slug IS NOT NULL;
