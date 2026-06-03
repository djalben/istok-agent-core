package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/istok/agent-core/internal/application/usecases"
	"github.com/istok/agent-core/internal/domain"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Projects Handler (delivery) — CRUD + remix.
//  Все маршруты за AuthMiddleware: userID берётся из контекста.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type ProjectsHandler struct {
	svc *usecases.ProjectService
}

func NewProjectsHandler(svc *usecases.ProjectService) *ProjectsHandler {
	return &ProjectsHandler{svc: svc}
}

type projectListResponse struct {
	Projects []*domain.Project `json:"projects"`
}

type createProjectBody struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Framework   string            `json:"framework"`
	Prompt      string            `json:"prompt"`
	IsPublic    bool              `json:"is_public"`
	Files       map[string]string `json:"files"`
}

type updateProjectBody struct {
	Name        *string           `json:"name"`
	Description *string           `json:"description"`
	Framework   *string           `json:"framework"`
	IsPublic    *bool             `json:"is_public"`
	FolderID    *string           `json:"folder_id"`
	WorkspaceID *string           `json:"workspace_id"`
	Files       map[string]string `json:"files"`
}

type remixBody struct {
	Name           string `json:"name"`
	IncludeHistory bool   `json:"include_history"`
}

// HandleList — GET /api/v1/projects
func (h *ProjectsHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	projects, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if projects == nil {
		projects = []*domain.Project{}
	}
	_ = writeJSON(w, http.StatusOK, projectListResponse{Projects: projects})
}

// HandleGet — GET /api/v1/projects/{id}
func (h *ProjectsHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	p, err := h.svc.Get(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, p)
}

// HandleCreate — POST /api/v1/projects
func (h *ProjectsHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	var body createProjectBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	p, err := h.svc.Create(r.Context(), userID, usecases.CreateProjectInput{
		Name:        body.Name,
		Description: body.Description,
		Framework:   body.Framework,
		Prompt:      body.Prompt,
		IsPublic:    body.IsPublic,
		Files:       body.Files,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	_ = writeJSON(w, http.StatusCreated, p)
}

// HandleUpdate — PATCH /api/v1/projects/{id}
func (h *ProjectsHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	var body updateProjectBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	patch := domain.ProjectPatch{
		Name:        body.Name,
		Description: body.Description,
		Framework:   body.Framework,
		IsPublic:    body.IsPublic,
		FolderID:    body.FolderID,
		WorkspaceID: body.WorkspaceID,
		Files:       body.Files,
	}
	p, err := h.svc.Update(r.Context(), userID, r.PathValue("id"), patch)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, p)
}

// HandleRemix — POST /api/v1/projects/{id}/remix
func (h *ProjectsHandler) HandleRemix(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	var body remixBody
	_ = json.NewDecoder(r.Body).Decode(&body) // тело опционально
	p, err := h.svc.Remix(r.Context(), userID, r.PathValue("id"), body.Name, body.IncludeHistory)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	_ = writeJSON(w, http.StatusCreated, p)
}

// HandleDelete — DELETE /api/v1/projects/{id}
func (h *ProjectsHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	if err := h.svc.Delete(r.Context(), userID, r.PathValue("id")); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeDomainError маппит доменные ошибки в HTTP-статусы (общий хелпер пакета).
func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "Не найдено")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusForbidden, "Доступ запрещён")
	case errors.Is(err, domain.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "Не авторизован")
	case errors.Is(err, domain.ErrEmailExists):
		writeError(w, http.StatusConflict, "Email уже зарегистрирован")
	case errors.Is(err, domain.ErrInvalidCreds):
		writeError(w, http.StatusUnauthorized, "Неверный email или пароль")
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}
