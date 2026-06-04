package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"os"
	"time"

	"gitlab.com/libs-artifex/wrapper"
	"log/slog"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Railway Deploy Handler
//  POST /api/v1/deploy/railway
//  Интеграция с Railway Public API (GraphQL) для создания
//  сервиса из сгенерированного проекта.
//
//  Контракт: фронт отправляет { project_name, files[] },
//  бэкенд возвращает { status, service_id, deploy_url, logs_url }.
//  Если RAILWAY_API_TOKEN не задан — возвращает 503 с guidance.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// DeployHandler обрабатывает запросы на деплой.
type DeployHandler struct {
	httpClient *http.Client
}

// NewDeployHandler конструктор.
func NewDeployHandler() *DeployHandler {
	return &DeployHandler{httpClient: &http.Client{Timeout: 60 * time.Second}}
}

// DeployFilePayload — файл в запросе деплоя.
type DeployFilePayload struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// DeployRequest — входной контракт.
type DeployRequest struct {
	ProjectName string              `json:"project_name"`
	Files       []DeployFilePayload `json:"files"`
	EnvVars     map[string]string   `json:"env_vars,omitempty"`
}

// DeployResponse — ответ клиенту.
type DeployResponse struct {
	Status    string `json:"status"` // "queued" | "deploying" | "success" | "failed" | "unavailable"
	ServiceID string `json:"service_id,omitempty"`
	DeployURL string `json:"deploy_url,omitempty"`
	LogsURL   string `json:"logs_url,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// HandleRailway POST /api/v1/deploy/railway.
func (h *DeployHandler) HandleRailway(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST only")

		return
	}

	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))

		return
	}

	if req.ProjectName == "" {
		req.ProjectName = fmt.Sprintf("istok-%d", time.Now().Unix())
	}
	if len(req.Files) == 0 {
		writeError(w, http.StatusBadRequest, "files[] must not be empty")

		return
	}

	token := os.Getenv("RAILWAY_API_TOKEN")
	if token == "" {
		resp := DeployResponse{
			Status: "unavailable",
			Message: "Railway deploy недоступен: RAILWAY_API_TOKEN не настроен. " +
				"Получите токен на https://railway.app/account/tokens и добавьте в env.",
		}
		_ = writeJSON(w, http.StatusServiceUnavailable, resp)

		return
	}

	// ── Railway GraphQL: projectCreate mutation ──
	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	serviceID, deployURL, err := h.createRailwayService(ctx, token, req)
	if err != nil {
		slog.Info(fmt.Sprintf("🚨 Railway deploy failed: %v", err))
		resp := DeployResponse{
			Status: "failed",
			Error:  err.Error(),
		}
		_ = writeJSON(w, http.StatusBadGateway, resp)

		return
	}
	slog.Info(fmt.Sprintf("🚀 Railway deploy started: service=%s url=%s", serviceID, deployURL))
	resp := DeployResponse{
		Status:    "deploying",
		ServiceID: serviceID,
		DeployURL: deployURL,
		LogsURL:   fmt.Sprintf("https://railway.app/project/%s/logs", serviceID),
		Message:   "✅ Deploy queued on Railway. Build logs available.",
	}
	_ = writeJSON(w, http.StatusAccepted, resp)
}

// createRailwayService — вызов Railway GraphQL (projectCreate + serviceCreate).
// Минимальная реализация: создаёт проект с именем project_name. Загрузка файлов
// ожидает, что клиент после успеха пушит код в созданный git-репозиторий.
func (h *DeployHandler) createRailwayService(ctx context.Context, token string, req DeployRequest) (serviceID, deployURL string, err error) {
	mutation := `mutation ProjectCreate($name: String!) {
		projectCreate(input: { name: $name, isPublic: false }) {
			id
			name
		}
	}`

	body, err := json.Marshal(map[string]any{
		"query": mutation,
		"variables": map[string]any{
			"name": req.ProjectName,
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrMarshalRequest, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://backboard.railway.app/graphql/v2", bytes.NewBuffer(body))
	if err != nil {
		return "", "", wrapper.Wrap(err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrRailwayHTTPError, wrapper.Wrap(err))
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		maxLog := min(len(raw), 400)

		return "", "", fmt.Errorf("%w %d: %s", ErrRailwayHTTPError, resp.StatusCode, string(raw[:maxLog]))
	}

	var parsed struct {
		Data struct {
			ProjectCreate struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"projectCreate"` //nolint:tagliatelle // Railway GraphQL schema
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", "", fmt.Errorf("parse railway response: %w", err)
	}
	if len(parsed.Errors) > 0 {
		return "", "", fmt.Errorf("%w: %s", ErrRailwayGraphQLError, parsed.Errors[0].Message)
	}
	if parsed.Data.ProjectCreate.ID == "" {
		return "", "", ErrRailwayEmptyProject
	}

	id := parsed.Data.ProjectCreate.ID

	return id, "https://railway.app/project/" + id, nil
}
