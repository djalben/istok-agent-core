package domain

import "time"

// Project — сущность сгенерированного проекта (Layer 1: Auth & DB).
//
// Files хранится как JSONB в Postgres (map[путь]содержимое).
// В списочных ответах (ProjectSummary) поле Files опускается (nil + omitempty),
// в детальном ответе (ProjectDetail) — заполняется.
type Project struct {
	ID           string            `json:"id"`
	OwnerID      string            `json:"-"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Framework    string            `json:"framework"`
	IsPublic     bool              `json:"is_public"`
	Slug         *string           `json:"slug"`
	ThumbnailURL *string           `json:"thumbnail_url"`
	Prompt       string            `json:"prompt,omitempty"`
	Files        map[string]string `json:"files,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ProjectPatch — частичное обновление проекта (PATCH /projects/:id).
// nil-поле означает «не менять».
type ProjectPatch struct {
	Name        *string
	Description *string
	Framework   *string
	IsPublic    *bool
	FolderID    *string
	WorkspaceID *string
	Files       map[string]string
}
