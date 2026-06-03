package http

import "net/http"

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Folders & Workspaces — STUB handlers (Layer 1).
//  Возвращают минимальные seed-данные; таблиц пока нет.
//  Контракт совпадает с FolderListResponse / WorkspaceListResponse фронта.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type folderDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProjectCount int    `json:"project_count"`
}

type workspaceDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	IsPersonal bool   `json:"is_personal"`
}

type WorkspaceHandler struct{}

func NewWorkspaceHandler() *WorkspaceHandler { return &WorkspaceHandler{} }

// HandleFolders — GET /api/v1/folders (stub: пустой список).
func (h *WorkspaceHandler) HandleFolders(w http.ResponseWriter, r *http.Request) {
	_ = writeJSON(w, http.StatusOK, map[string][]folderDTO{
		"folders": {},
	})
}

// HandleWorkspaces — GET /api/v1/workspaces (stub: личное пространство).
func (h *WorkspaceHandler) HandleWorkspaces(w http.ResponseWriter, r *http.Request) {
	_ = writeJSON(w, http.StatusOK, map[string][]workspaceDTO{
		"workspaces": {
			{ID: "personal", Name: "Личное пространство", Role: "owner", IsPersonal: true},
		},
	})
}
