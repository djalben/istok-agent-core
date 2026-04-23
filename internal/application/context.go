package application

import (
	"fmt"
	"strings"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Intelligent Context — Project Snapshot Builder
//  Собирает полный контекст проекта для LLM агентов.
//  Каждый агент видит не только задачу, но и интерфейсы
//  всех существующих модулей.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ProjectSnapshot — полный слепок проекта для LLM контекста.
// Передаётся каждому агенту чтобы он знал архитектуру целиком.
type ProjectSnapshot struct {
	// Specification — исходное ТЗ пользователя
	Specification string

	// URL — анализируемый сайт (если synthesis mode)
	URL string

	// Mode — режим генерации (agent/code/synthesis)
	Mode GenerationMode

	// Architecture — результат работы Architect (если есть)
	Architecture *ArchitectureManifest

	// MasterPlan — план от Director
	MasterPlan *MasterPlan

	// Audit — результат reverse engineering (если synthesis)
	Audit *ReverseEngineeringResult

	// ExistingFiles — текущие файлы проекта (для итеративной доработки)
	ExistingFiles map[string]string

	// Modules — интерфейсы и контракты всех модулей
	Modules []ModuleInfo

	// AgentHistory — что уже сделали предыдущие агенты
	AgentHistory []AgentHistoryEntry
}

// ModuleInfo — описание модуля/интерфейса проекта
type ModuleInfo struct {
	Name        string   // e.g. "CodeGenerator", "Orchestrator"
	Layer       string   // domain / application / infrastructure / transport
	Interfaces  []string // список методов/контрактов
	Description string
}

// AgentHistoryEntry — запись о работе предыдущего агента
type AgentHistoryEntry struct {
	Agent   AgentRole
	Action  string
	Summary string
}

// ArchitectureManifest — результат работы Architect
type ArchitectureManifest struct {
	Stack       []string          `json:"stack"`
	Pages       []string          `json:"pages"`
	Components  []string          `json:"components"`
	ColorScheme map[string]string `json:"color_scheme"`
	Typography  map[string]string `json:"typography"`
	Features    []string          `json:"features"`
}

// NewProjectSnapshot creates empty snapshot with spec and mode.
func NewProjectSnapshot(spec string, url string, mode GenerationMode) *ProjectSnapshot {
	return &ProjectSnapshot{
		Specification: spec,
		URL:           url,
		Mode:          mode,
		ExistingFiles: make(map[string]string),
		Modules:       defaultModules(),
		AgentHistory:  make([]AgentHistoryEntry, 0),
	}
}

// AddAgentResult записывает что сделал агент.
func (ps *ProjectSnapshot) AddAgentResult(agent AgentRole, action, summary string) {
	ps.AgentHistory = append(ps.AgentHistory, AgentHistoryEntry{
		Agent:   agent,
		Action:  action,
		Summary: summary,
	})
}

// ForAgent генерирует контекстный промпт для конкретного агента.
// Включает: архитектуру, модули, историю работы предыдущих агентов.
func (ps *ProjectSnapshot) ForAgent(role AgentRole) string {
	var b strings.Builder

	b.WriteString("## PROJECT CONTEXT\n\n")
	b.WriteString(fmt.Sprintf("Mode: %s\n", ps.Mode))
	b.WriteString(fmt.Sprintf("Specification: %s\n\n", truncate(ps.Specification, 500)))

	if ps.URL != "" {
		b.WriteString(fmt.Sprintf("Target URL: %s\n\n", ps.URL))
	}

	// Architecture
	if ps.Architecture != nil {
		b.WriteString("### Architecture\n")
		b.WriteString(fmt.Sprintf("Stack: %s\n", strings.Join(ps.Architecture.Stack, ", ")))
		b.WriteString(fmt.Sprintf("Pages: %s\n", strings.Join(ps.Architecture.Pages, ", ")))
		b.WriteString(fmt.Sprintf("Components: %s\n\n", strings.Join(ps.Architecture.Components, ", ")))
	}

	// MasterPlan
	if ps.MasterPlan != nil {
		b.WriteString("### Master Plan\n")
		b.WriteString(fmt.Sprintf("Architecture: %s\n", truncate(ps.MasterPlan.Architecture, 200)))
		b.WriteString(fmt.Sprintf("Technologies: %s\n", strings.Join(ps.MasterPlan.Technologies, ", ")))
		b.WriteString(fmt.Sprintf("Components: %s\n\n", strings.Join(ps.MasterPlan.Components, ", ")))
	}

	// Module interfaces — only for Coder/Architect
	if role == RoleCoder || role == RoleBrain {
		b.WriteString("### Module Interfaces (Clean Architecture)\n")
		for _, m := range ps.Modules {
			b.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", m.Name, m.Layer, m.Description))
			for _, iface := range m.Interfaces {
				b.WriteString(fmt.Sprintf("  - %s\n", iface))
			}
		}
		b.WriteString("\n")
	}

	// Agent history
	if len(ps.AgentHistory) > 0 {
		b.WriteString("### Previous Agent Results\n")
		for _, h := range ps.AgentHistory {
			b.WriteString(fmt.Sprintf("- [%s] %s: %s\n", h.Agent, h.Action, truncate(h.Summary, 150)))
		}
		b.WriteString("\n")
	}

	// Existing files summary (for iterative mode)
	if len(ps.ExistingFiles) > 0 {
		b.WriteString("### Existing Files\n")
		for name, content := range ps.ExistingFiles {
			b.WriteString(fmt.Sprintf("- %s (%d chars)\n", name, len(content)))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// defaultModules returns our project's module descriptions.
func defaultModules() []ModuleInfo {
	return []ModuleInfo{
		{
			Name:        "Orchestrator",
			Layer:       "application",
			Description: "S-Tier AI agent coordinator. Manages Director→Researcher→Architect→Coder→Validator pipeline.",
			Interfaces: []string{
				"GenerateWithMode(ctx, spec, url, mode) → GenerationResult",
				"GetStatusStream() → chan TaskStatus",
			},
		},
		{
			Name:        "CodeGenerator",
			Layer:       "ports",
			Description: "Contract for code generation, website analysis, refactoring.",
			Interfaces: []string{
				"GenerateCode(ctx, req) → GenerateCodeResponse",
				"AnalyzeWebsite(ctx, req) → AnalyzeWebsiteResponse",
				"RefactorCode(ctx, req) → RefactorCodeResponse",
				"ValidateOutput(ctx, code, lang) → ValidationResponse",
			},
		},
		{
			Name:        "MediaGenerator",
			Layer:       "ports",
			Description: "Contract for UI asset and video generation.",
			Interfaces: []string{
				"GenerateUIAssets(ctx, spec, plan) → map[string]string",
				"GeneratePromoVideo(ctx, spec, plan) → string",
			},
		},
		{
			Name:        "ResearcherAgent",
			Layer:       "application",
			Description: "DeepSeek V3.2 — competitive analysis, spec enrichment, technology audit.",
			Interfaces: []string{
				"AnalyzeSpec(ctx, spec) → ResearchResult",
				"ReverseEngineer(ctx, url) → ReverseEngineeringResult",
			},
		},
		{
			Name:        "SSE Transport",
			Layer:       "transport",
			Description: "Server-Sent Events handler for real-time generation status streaming.",
			Interfaces: []string{
				"POST /api/v1/generate/stream → SSE stream",
				"GET /api/v1/health → health check",
			},
		},
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
