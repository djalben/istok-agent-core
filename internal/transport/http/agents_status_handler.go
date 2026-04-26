package http

import (
	"encoding/json"
	"net/http"

	"github.com/istok/agent-core/internal/application"
	"github.com/istok/agent-core/internal/application/dto"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Agents Status Handler
//  GET /api/v1/agents/status
//  Возвращает пайплайн агентов, модели, провайдеров, thinking-режим.
//  Контракт синхронизирован с Zod-схемой web/src/lib/contracts.ts.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// AgentsStatusHandler отдаёт метаданные всех агентов.
type AgentsStatusHandler struct {
	orchestrator *application.Orchestrator
}

// NewAgentsStatusHandler создаёт handler.
func NewAgentsStatusHandler(orchestrator *application.Orchestrator) *AgentsStatusHandler {
	return &AgentsStatusHandler{orchestrator: orchestrator}
}

// Handle обрабатывает GET /api/v1/agents/status.
func (h *AgentsStatusHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	descriptors := h.orchestrator.AgentDescriptors()
	agents := make([]dto.AgentInfo, 0, len(descriptors))
	for _, d := range descriptors {
		agents = append(agents, dto.AgentInfo{
			Role:        d.Role,
			Model:       d.Model,
			Provider:    d.Provider,
			Description: d.Description,
			Thinking:    d.Thinking,
			TimeoutSec:  d.TimeoutSec,
		})
	}

	response := dto.AgentStatusResponse{
		Agents:      agents,
		FSMStates:   12,
		EventBuffer: 128,
		Pipeline:    application.CanonicalPipeline,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
