package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/istok/agent-core/internal/application/usecases"
)

// ProjectsHandler обрабатывает запросы к проектам
type ProjectsHandler struct {
	projectGenerator *usecases.ProjectGeneratorService
}

// NewProjectsHandler создает новый handler
func NewProjectsHandler(projectGenerator *usecases.ProjectGeneratorService) *ProjectsHandler {
	return &ProjectsHandler{
		projectGenerator: projectGenerator,
	}
}

// ProjectStats статистика проекта
type ProjectStats struct {
	ProjectID           string `json:"projectId"`
	Model               string `json:"model"`
	ModelVersion        string `json:"modelVersion"`
	ResponseTimeMs      int    `json:"responseTimeMs"`
	CrawlerNodesFound   int    `json:"crawlerNodesFound"`
	GeneratedFilesCount int    `json:"generatedFilesCount"`
	TokensUsed          int    `json:"tokensUsed"`
	CostRub             int    `json:"costRub"`
	Status              string `json:"status"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
}

// GeneratedFile сгенерированный файл
type GeneratedFile struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Size     int    `json:"sizeBytes"`
	Preview  string `json:"preview"`
}

// Project полный проект
type Project struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Stats      ProjectStats    `json:"stats"`
	Messages   []AgentMessage  `json:"messages"`
	Files      []GeneratedFile `json:"files"`
	PreviewURL *string         `json:"previewUrl"`
}

// HandleGetProject обрабатывает GET /api/v1/projects/:id
func (h *ProjectsHandler) HandleGetProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	projectID := extractProjectID(r.URL.Path)
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "Не указан ID проекта")
		return
	}

	// Возвращаем моковый проект
	project := Project{
		ID:   projectID,
		Name: "Конкурентный Анализ #1",
		Stats: ProjectStats{
			ProjectID:           projectID,
			Model:               "Claude 3.5 Sonnet",
			ModelVersion:        "3.5.0",
			ResponseTimeMs:      142,
			CrawlerNodesFound:   847,
			GeneratedFilesCount: 23,
			TokensUsed:          18420,
			CostRub:             2340,
			Status:              "ready",
			CreatedAt:           time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			UpdatedAt:           time.Now().Format(time.RFC3339),
		},
		Messages: []AgentMessage{
			{
				ID:        "msg_1",
				ProjectID: projectID,
				Role:      "system",
				Content:   "Сессия инициализирована. Готов к работе.",
				Timestamp: time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				Status:    "complete",
			},
		},
		Files: []GeneratedFile{
			{
				Path:     "app/layout.tsx",
				Language: "tsx",
				Size:     1240,
				Preview:  "export default function RootLayout({ children }…",
			},
			{
				Path:     "app/page.tsx",
				Language: "tsx",
				Size:     3870,
				Preview:  "import { Hero } from '@/components/hero'…",
			},
		},
		PreviewURL: nil,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(project)
}

// HandleGetProjectStats обрабатывает GET /api/v1/projects/:id/stats
func (h *ProjectsHandler) HandleGetProjectStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	projectID := extractProjectID(r.URL.Path)
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "Не указан ID проекта")
		return
	}

	stats := ProjectStats{
		ProjectID:           projectID,
		Model:               "Claude 3.5 Sonnet",
		ModelVersion:        "3.5.0",
		ResponseTimeMs:      142,
		CrawlerNodesFound:   847,
		GeneratedFilesCount: 23,
		TokensUsed:          18420,
		CostRub:             2340,
		Status:              "ready",
		CreatedAt:           time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		UpdatedAt:           time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}
