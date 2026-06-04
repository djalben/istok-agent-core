package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — S-Tier AI Orchestrator
//  Мультимодельная архитектура нового поколения
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// GenerationMode режим генерации.
type GenerationMode string

const (
	ModeAgent     GenerationMode = "agent"     // Инновационное проектирование: ядро Истока — глубокий анализ
	ModeCode      GenerationMode = "code"      // Быстрая генерация UI
	ModeSynthesis GenerationMode = "synthesis" // Адаптивный синтез конкурентов
)

// AgentRole — алиас на domain.AgentRole для обратной совместимости внутри application слоя.
type AgentRole = domain.AgentRole

// Константы ролей — алиасы на domain-константы.
const (
	RoleDirector     = domain.RoleDirector
	RoleBrain        = domain.RoleBrain
	RoleResearcher   = domain.RoleResearcher
	RoleArchitect    = domain.RoleArchitect
	RolePlanner      = domain.RolePlanner
	RoleCoder        = domain.RoleCoder
	RoleDesigner     = domain.RoleDesigner
	RoleVideographer = domain.RoleVideographer
	RoleValidator    = domain.RoleValidator
	RoleSecurity     = domain.RoleSecurity
	RoleTester       = domain.RoleTester
	RoleUIReviewer   = domain.RoleUIReviewer
)

// AgentConfig конфигурация агента.
type AgentConfig struct {
	Role            AgentRole
	Model           string
	Description     string
	Timeout         time.Duration
	ThinkingEnabled bool
}

// TaskStatus — алиас для обратной совместимости (SSE handler и др.).
// Новый код должен использовать domain.AgentEvent.
type TaskStatus = domain.AgentEvent

// ReverseEngineeringResult результат анализа сайта.
type ReverseEngineeringResult struct {
	URL          string
	Colors       []string
	Fonts        []string
	Components   []string
	Layout       string
	Technologies []string
	Audit        string
}

// DAGTask — узел DAG-плана разработки.
// Каждая задача знает свои зависимости, затрагиваемые файлы и требуемые пакеты.
type DAGTask struct {
	ID                   string   `json:"id"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	DependsOn            []string `json:"depends_on"`
	ImpactedFiles        []string `json:"impacted_files"`
	RequiredDependencies []string `json:"required_dependencies"`
}

// MasterPlan план разработки от директора (Planner).
type MasterPlan struct {
	Architecture    string
	Components      []string
	Timeline        string
	Technologies    []string
	Steps           []string  // backward-compat flat steps
	DAG             []DAGTask // DAG-представление плана
	ThinkingProcess string    `json:"-"` // скрытое поле: ход мыслей агента (логируется, не отправляется клиенту)
}

// GenerationResult финальный результат генерации.
type GenerationResult struct {
	Code        map[string]string
	Assets      map[string]string
	Video       string
	MasterPlan  *MasterPlan
	Audit       *ReverseEngineeringResult
	VisualAudit *VisualAuditResult
	Duration    time.Duration
}

// Orchestrator управляет пулом AI агентов.
// Зависит от ports.LLMProvider (Dependency Rule) и domain.EventBus (канальный протокол).
type Orchestrator struct {
	agents           map[AgentRole]*AgentConfig
	llm              ports.LLMProvider
	uiMedia          ports.UIMediaService
	events           *domain.EventBus // дефолтная шина (сессии без sessionID: curl/тесты)
	buses            *busRegistry     // per-session шины (изоляция параллельных генераций)
	projectEnv       *ProjectEnv
	planner          *usecases.PlannerAgent   // модернизированный Планировщик с DAG + FSM gate
	projectCtx       *usecases.ProjectContext // отсканированный package.json/tsconfig.json
	sessionCache     *SessionCache            // tier checkpoints for resume
	approvalRegistry *ApprovalRegistry        // human-in-the-loop approval channels
	fundsRegistry    *FundsRegistry           // pause/resume on insufficient funds
	mu               sync.RWMutex
}

// NewOrchestrator создает оркестратор с LLM-провайдером (через порт) и шиной событий.
func NewOrchestrator(llm ports.LLMProvider, uiMedia ports.UIMediaService) *Orchestrator {
	return &Orchestrator{
		llm:              llm,
		uiMedia:          uiMedia,
		events:           domain.NewEventBus(256),
		buses:            newBusRegistry(256),
		sessionCache:     NewSessionCache(30 * time.Minute),
		approvalRegistry: NewApprovalRegistry(15 * time.Minute),
		fundsRegistry:    NewFundsRegistry(2 * time.Hour),
		planner:          usecases.NewPlannerAgent(llm, "anthropic/claude-sonnet-4-6-thinking"),
		agents: map[AgentRole]*AgentConfig{
			RoleDirector: {
				Role:        RoleDirector,
				Model:       "anthropic/claude-sonnet-4-6-thinking",
				Description: "🧠 Директор — Ядро Истока (планирование)",
				Timeout:     5 * time.Minute,
			},
			RoleBrain: {
				Role:        RoleBrain,
				Model:       "anthropic/claude-sonnet-4-6-thinking",
				Description: "🧠 Мозг — Ядро Истока (архитектура)",
				Timeout:     10 * time.Minute,
			},
			RoleResearcher: {
				Role:        RoleResearcher,
				Model:       "anthropic/claude-sonnet-4-6-thinking",
				Description: "🔍 Исследователь — Ядро Истока (анализ)",
				Timeout:     5 * time.Minute,
			},
			RoleCoder: {
				Role:        RoleCoder,
				Model:       "anthropic/claude-sonnet-4-6",
				Description: "💻 Кодер — AI Istok (код)",
				Timeout:     10 * time.Minute,
			},
			RoleDesigner: {
				Role:        RoleDesigner,
				Model:       "google/nano-banana",
				Description: "🎨 Дизайнер — визуальные ассеты Истока",
				Timeout:     5 * time.Minute,
			},
			RoleVideographer: {
				Role:        RoleVideographer,
				Model:       "google/veo-3",
				Description: "🎬 Видеограф — Veo 3 (промо-видео, Replicate)",
				Timeout:     15 * time.Minute,
			},
			RoleValidator: {
				Role:        RoleValidator,
				Model:       "anthropic/claude-sonnet-4-6",
				Description: "✅ Валидатор — AI Istok (Syntax & Runtime)",
				Timeout:     3 * time.Minute,
			},
		},
	}
}

// GetUIMedia returns the UI media service (composition root injects infrastructure adapter).
func (o *Orchestrator) GetUIMedia() ports.UIMediaService {
	return o.uiMedia
}

// GetLLM returns the LLM provider instance.
func (o *Orchestrator) GetLLM() ports.LLMProvider {
	return o.llm
}

// GetSessionCache returns the session checkpoint cache for resume support.
func (o *Orchestrator) GetSessionCache() *SessionCache {
	return o.sessionCache
}

// GetApprovalRegistry returns the human-in-the-loop approval registry.
func (o *Orchestrator) GetApprovalRegistry() *ApprovalRegistry {
	return o.approvalRegistry
}

// GetFundsRegistry returns the insufficient-funds pause/resume registry.
func (o *Orchestrator) GetFundsRegistry() *FundsRegistry {
	return o.fundsRegistry
}

// SetProjectEnv устанавливает результат ProjectScanner для передачи агентам.
// Вызывается перед GenerateWithMode, если есть package.json/tsconfig.json для сканирования.
func (o *Orchestrator) SetProjectEnv(env *ProjectEnv) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.projectEnv = env
}

// SetProjectContext устанавливает контекст проекта (package.json + tsconfig.json),
// который PlannerAgent использует для инъекции точных версий и path-алиасов в промпт.
func (o *Orchestrator) SetProjectContext(pc *usecases.ProjectContext) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.projectCtx = pc
}

// ScanProjectFiles читает package.json и tsconfig.json через PlannerAgent
// и сохраняет результат в Orchestrator.projectCtx. Возвращает ошибку только если
// чтение упало; пустые файлы тихо игнорируются.
func (o *Orchestrator) ScanProjectFiles(packageJSONPath, tsconfigPath string) error {
	if o.planner == nil {
		return ErrPlannerNotInitialized
	}
	pc, err := o.planner.ScanProject(packageJSONPath, tsconfigPath)
	if err != nil {
		return wrapper.Wrap(err)
	}
	o.SetProjectContext(pc)

	return nil
}

// Planner возвращает PlannerAgent для прямого использования транспортным слоем
// (например, чтобы вызвать ValidateReadiness перед стартом генерации).
func (o *Orchestrator) Planner() *usecases.PlannerAgent { return o.planner }

// AgentDescriptor — публичное описание одного агента (для HTTP-контракта).
type AgentDescriptor struct {
	Role        string
	Model       string
	Provider    string
	Description string
	Thinking    bool
	TimeoutSec  int
}

// AgentPipelineOrder — каноничный порядок выполнения агентов пайплайна.
// Идентичен константе AGENT_PIPELINE в web/src/hooks/useGeneration.ts.
var AgentPipelineOrder = []AgentRole{
	RoleDirector,
	RoleResearcher,
	RoleBrain,
	RolePlanner,
	RoleCoder,
	RoleDesigner,
	RoleValidator,
	RoleVideographer,
}

// CanonicalPipeline — строковые идентификаторы агентов в порядке выполнения,
// расширенный список (включая роли, не имеющие отдельного AgentConfig: planner,
// tester, ui_reviewer, security). Должен совпадать с AGENT_PIPELINE на фронте.
var CanonicalPipeline = []string{
	"director",
	"researcher",
	"brain",
	"architect",
	"planner",
	"coder",
	"designer",
	"validator",
	"security",
	"tester",
	"ui_reviewer",
	"videographer",
}

// AgentDescriptors возвращает публичное описание всех агентов (для /agents/status).
// Провайдер определяется по префиксу модели (anthropic/* → Anthropic Direct,
// google/*, black-forest-labs/* → Replicate).
func (o *Orchestrator) AgentDescriptors() []AgentDescriptor {
	o.mu.RLock()
	defer o.mu.RUnlock()

	result := make([]AgentDescriptor, 0, len(o.agents))
	for role, cfg := range o.agents {
		thinking := strings.Contains(strings.ToLower(cfg.Model), "thinking")
		provider := "Local"
		switch {
		case strings.HasPrefix(cfg.Model, "anthropic/"), strings.HasPrefix(cfg.Model, "claude-"):
			provider = "Istok Core"
		case strings.HasPrefix(cfg.Model, "google/"),
			strings.HasPrefix(cfg.Model, "black-forest-labs/"),
			strings.HasPrefix(cfg.Model, "ideogram-ai/"):
			provider = "Replicate"
		}
		result = append(result, AgentDescriptor{
			Role:        string(role),
			Model:       cfg.Model,
			Provider:    provider,
			Description: cfg.Description,
			Thinking:    thinking,
			TimeoutSec:  int(cfg.Timeout.Seconds()),
		})
	}

	return result
}

// GenerateWithMode запускает процесс генерации в указанном режиме.
func (o *Orchestrator) GenerateWithMode(ctx context.Context, specification string, url string, mode GenerationMode) (*GenerationResult, error) {
	if mode == ModeCode {
		return o.generateCodeMode(ctx, specification)
	}
	// Both "agent" (Инновационное проектирование) and "synthesis" (Адаптивный синтез) use full pipeline
	return o.generateAgentMode(ctx, specification, url)
}

// SubscribeSession регистрирует изолированную шину сессии и возвращает её канал.
// Должна вызываться ДО старта генерации, чтобы ранние события не потерялись.
// Освобождение — через ReleaseSession (в defer горутины генерации).
func (o *Orchestrator) SubscribeSession(sessionID string) <-chan domain.AgentEvent {
	if sessionID == "" {
		return o.events.Subscribe()
	}

	return o.buses.acquire(sessionID).Subscribe()
}

// ReleaseSession удаляет шину сессии из реестра.
// Вызывается горутиной генерации при завершении (последний publisher).
func (o *Orchestrator) ReleaseSession(sessionID string) {
	if sessionID != "" {
		o.buses.release(sessionID)
	}
}

// GetStatusStream возвращает канал дефолтной шины (обратная совместимость).
func (o *Orchestrator) GetStatusStream() <-chan domain.AgentEvent {
	return o.events.Subscribe()
}

// GetEventBus возвращает дефолтную шину событий для прямого использования.
func (o *Orchestrator) GetEventBus() *domain.EventBus {
	return o.events
}

// Close закрывает оркестратор и шину событий.
func (o *Orchestrator) Close() {
	o.events.Close()
}

// generateCodeMode быстрая генерация через ядро Истока (Code Mode).
func (o *Orchestrator) generateCodeMode(ctx context.Context, specification string) (*GenerationResult, error) {
	startTime := time.Now()
	result := &GenerationResult{
		Code:   make(map[string]string),
		Assets: make(map[string]string),
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	// Инициализируем FSM
	fsm := domain.NewTaskStateMachine()

	// Code mode: Created → Planning (авто-план) → Coding → Completed
	err := fsm.TransitionTo(domain.StatePlanning, "code mode: fast planning")
	if err != nil {
		return nil, fmt.Errorf("FSM: %w", err)
	}
	o.busFromCtx(ctx).PublishFSMTransition(domain.StateCreated, domain.StatePlanning, "code mode")

	// ── Director milestone (UI Agent Pulse): планирование ──
	o.sendStatus(ctx, RoleDirector, "running", "🧠 Директор Истока составляет план...", 5)

	plan := &MasterPlan{
		Architecture: "Quick UI Generation",
		Steps:        []string{specification},
	}

	o.sendStatus(ctx, RoleDirector, "completed", "✅ План готов — передаю Кодеру", 15)

	// Утверждаем план в FSM (gate для Coding)
	err = fsm.ApprovePlan(domain.ApprovedPlan{
		Architecture: plan.Architecture,
		Steps:        plan.Steps,
		ApprovedBy:   "code_mode_auto",
	})
	if err != nil {
		return nil, fmt.Errorf("FSM plan approval: %w", err)
	}

	err = fsm.TransitionTo(domain.StateArchitectureApproved, "auto-approved for code mode")
	if err != nil {
		return nil, fmt.Errorf("FSM: %w", err)
	}

	// ArchitectureApproved → StrategySynthesized → Coding.
	// FSM не допускает прямой переход architecture_approved → coding (см. allowedTransitions).
	err = fsm.TransitionTo(domain.StateStrategySynthesized, "code mode: strategy skipped")
	if err != nil {
		return nil, fmt.Errorf("FSM: %w", err)
	}

	// Переход в Coding (пройдёт только если план утверждён)
	err = fsm.TransitionTo(domain.StateCoding, "plan approved, starting code generation")
	if err != nil {
		return nil, fmt.Errorf("FSM: %w", err)
	}
	o.busFromCtx(ctx).PublishFSMTransition(domain.StateStrategySynthesized, domain.StateCoding, "code mode")

	o.sendStatus(ctx, RoleCoder, "running", "⚡ Кодер Истока генерирует UI компоненты...", 20)

	code, err := o.generateCode(ctx, specification, plan, nil, nil)
	if err != nil {
		_ = fsm.TransitionTo(domain.StateFailed, err.Error())
		o.sendStatus(ctx, RoleCoder, "error", fmt.Sprintf("❌ Ошибка: %v", err), 0)

		return nil, err
	}

	_ = fsm.TransitionTo(domain.StateQualityCheck, "code generated")
	_ = fsm.TransitionTo(domain.StateSecurityCheck, "quality ok")
	_ = fsm.TransitionTo(domain.StateVerified, "security ok")
	_ = fsm.TransitionTo(domain.StateCompleted, "done")

	result.Code = code
	result.Duration = time.Since(startTime)
	o.sendStatus(ctx, RoleCoder, "completed", fmt.Sprintf("✅ Код готов за %v", result.Duration), 100)

	return result, nil
}

func convertPlanTasks(tasks []usecases.PlanTask) []DAGTask {
	out := make([]DAGTask, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, DAGTask{
			ID:                   t.ID,
			Title:                t.Title,
			Description:          t.Description,
			DependsOn:            t.DependsOn,
			ImpactedFiles:        t.ImpactedFiles,
			RequiredDependencies: t.RequiredDependencies,
		})
	}

	return out
}

func (o *Orchestrator) tryPlannerMasterPlan(
	ctx context.Context,
	specification, auditSummary, thinkingLog string,
	projectCtx *usecases.ProjectContext,
) (*MasterPlan, bool) {
	if o.planner == nil {
		return nil, false
	}
	applog(ctx).InfoContext(ctx, "planner agent request", "model", o.planner.Model)
	uPlan, err := o.planner.BuildPlan(ctx, specification, auditSummary, projectCtx)
	if err != nil || uPlan == nil || len(uPlan.Tasks) == 0 {
		applog(ctx).WarnContext(ctx, "planner agent failed, legacy fallback", "error", err)

		return nil, false
	}
	plan := &MasterPlan{
		Architecture:    uPlan.Architecture,
		Components:      uPlan.Components,
		Technologies:    uPlan.Technologies,
		Timeline:        uPlan.Timeline,
		Steps:           uPlan.Steps,
		DAG:             convertPlanTasks(uPlan.Tasks),
		ThinkingProcess: thinkingLog,
	}
	if plan.Architecture == "" {
		plan.Architecture = specification
	}
	if len(plan.Steps) == 0 {
		for _, t := range uPlan.Tasks {
			plan.Steps = append(plan.Steps, t.Title)
		}
	}
	applog(ctx).InfoContext(ctx, "planner plan ready", "dag_tasks", len(plan.DAG), "exec_order", uPlan.ExecutionOrder)

	return plan, true
}

// createMasterPlan делегирует построение плана модернизированному PlannerAgent
// (DAG engine + Context Injection + Smart FSM Gate). Если у нас уже есть
// projectCtx (отсканированные package.json/tsconfig.json) — используем его,
// иначе строим временный из projectEnv для обратной совместимости.
func (o *Orchestrator) createMasterPlan(ctx context.Context, specification string, audit *ReverseEngineeringResult) (*MasterPlan, error) {
	agent := o.agents[RoleDirector]
	ctx, cancel := context.WithTimeout(ctx, agent.Timeout)
	defer cancel()

	// ── Phase 0: Reflective Reasoning (Thought Chain) ──
	directorReflection := "Plan implementation DAG for: " + specification[:min(len(specification), 300)]
	directorTC, _ := o.ThoughtChain(ctx, RoleDirector, directorReflection)
	directorReflectionCtx := ThoughtChainContext(directorTC)

	// Build audit summary for Director context
	auditSummary := "No visual audit available."
	if audit != nil {
		auditSummary = fmt.Sprintf(
			"Colors: %v | Components: %v | Layout: %s | Technologies: %v | DesignSystem: %s",
			audit.Colors, audit.Components, audit.Layout, audit.Technologies, audit.Audit,
		)
	}

	// ── Path 1: Delegate to PlannerAgent if we have project context ──
	o.mu.RLock()
	projectCtx := o.projectCtx
	projectEnv := o.projectEnv
	o.mu.RUnlock()

	// Bridge: если projectCtx не задан, но projectEnv есть — конвертируем
	if projectCtx == nil && projectEnv != nil {
		projectCtx = &usecases.ProjectContext{
			PackageName:    projectEnv.PackageName,
			Dependencies:   projectEnv.Dependencies,
			DevDeps:        projectEnv.DevDeps,
			Scripts:        projectEnv.Scripts,
			PackageManager: projectEnv.PackageManager,
			TSTarget:       projectEnv.TSTarget,
			TSModule:       projectEnv.TSModule,
			TSPaths:        projectEnv.TSPaths,
			TSStrict:       projectEnv.TSStrict,
			TSBaseURL:      projectEnv.TSBaseURL,
			Loaded:         projectEnv.PackageName != "" || len(projectEnv.Dependencies) > 0,
		}
	}

	// Capture thinking process for logging (hidden from client JSON via json:"-")
	thinkingLog := ""
	if directorTC != nil && directorTC.RawChain != "" {
		thinkingLog = directorTC.RawChain
		applog(ctx).DebugContext(ctx, "director thinking process", "chars", len(thinkingLog))
	}

	if plan, ok := o.tryPlannerMasterPlan(ctx, specification, auditSummary, thinkingLog, projectCtx); ok {
		return plan, nil
	}

	// ── Path 2 (legacy fallback): прямой LLM-вызов через Director ──
	envCtx := ""
	if projectEnv != nil {
		envCtx = projectEnv.ForPrompt()
	}

	userPrompt := fmt.Sprintf(`Create a FUNCTIONAL implementation plan as a DAG (Directed Acyclic Graph).
Each task must specify which files it touches and which packages it needs.

SPECIFICATION:
%s

DESIGN AUDIT (from Researcher Agent):
%s

Output ONLY a valid JSON object — no markdown, no explanation:
{
  "architecture": "architecture description with key data structures and business logic",
  "components": ["Component1 (with interaction description)", "Component2", ...],
  "technologies": ["vite", "react", "tailwindcss", "shadcn/ui", "@tanstack/react-query", ...],
  "timeline": "estimated timeline",
  "steps": ["Step 1: ...", "Step 2: ...", "..."],
  "dag": [
    {
      "id": "T1",
      "title": "Project scaffold & routing",
      "description": "Initialize Vite+React project, configure TanStack Router with @/* aliases",
      "depends_on": [],
      "impacted_files": ["package.json", "tsconfig.json", "vite.config.ts", "src/routes/__root.tsx"],
      "required_dependencies": ["vite", "react", "react-dom", "@tanstack/react-router"]
    },
    {
      "id": "T2",
      "title": "UI shell: layout + navigation",
      "description": "Build AppLayout with Sidebar, Header, MobileNav using shadcn components",
      "depends_on": ["T1"],
      "impacted_files": ["src/components/layout/AppLayout.tsx", "src/components/layout/Sidebar.tsx", "src/components/layout/Header.tsx"],
      "required_dependencies": ["@radix-ui/react-slot", "class-variance-authority", "lucide-react"]
    },
    {
      "id": "T3",
      "title": "Data layer: hooks + services",
      "description": "Create TanStack Query hooks and API service functions",
      "depends_on": ["T1"],
      "impacted_files": ["src/hooks/useAuth.ts", "src/services/api.ts", "src/services/auth.ts"],
      "required_dependencies": ["@tanstack/react-query"]
    },
    {
      "id": "T4",
      "title": "Feature pages",
      "description": "Build all route pages consuming hooks from T3 and layout from T2",
      "depends_on": ["T2", "T3"],
      "impacted_files": ["src/pages/Dashboard.tsx", "src/pages/Settings.tsx"],
      "required_dependencies": []
    }
  ]
}

CRITICAL RULES:
1. Each task MUST have "impacted_files" (exact paths it creates/modifies) and "required_dependencies" (npm packages).
2. "depends_on" references other task IDs — forms a DAG. No cycles allowed.
3. Tasks with no dependencies (depends_on=[]) can run in parallel.
4. Each step must describe FUNCTIONAL behavior, not just visual layout.
Bad: "Create hero section" Good: "Create hero with CTA button that smooth-scrolls to menu section"
%s%s`, specification, auditSummary, envCtx, directorReflectionCtx)
	applog(ctx).InfoContext(ctx, "director plan request", "model", agent.Model)

	result, err := o.callLLMWithReasoning(ctx, agent.Model,
		`You are a senior software architect and project planner. Create precise, actionable DAG plans.
Output only valid JSON. Every task must have impacted_files and required_dependencies.
ARCHITECTURE RULES:
- Never put business logic in main.go or HTTP handlers.
- Separate Domain (entities), Application (use cases), Infrastructure (external APIs), Transport (HTTP/SSE).
- All external dependencies must go through interfaces (ports).
- Use @/* import aliases. Structure: components/ui, components/layout, hooks, services.`,
		userPrompt, 4096)
	if err != nil {
		applog(ctx).ErrorContext(ctx, "director legacy LLM failed, default plan", "error", err)

		return o.defaultMasterPlan(specification, audit), nil
	}

	// DEBUG: raw Director LLM output before parsing
	debugDir := result
	if len(debugDir) > 500 {
		debugDir = debugDir[:500] + "...[truncated]"
	}
	applog(ctx).DebugContext(ctx, "director raw LLM output", "chars", len(result), "preview", debugDir)

	plan := o.parseMasterPlan(result, specification, audit)
	plan.ThinkingProcess = thinkingLog
	applog(ctx).InfoContext(ctx, "director plan ready", "steps", len(plan.Steps), "technologies", len(plan.Technologies))

	return plan, nil
}

// generateCodeFullStack вызывает Coder с полным контекстом: manifest + features + backend templates + imageURLs.
// Если manifest содержит FileMap с 5+ файлами — используется chunked generation (по группам).
// Иначе — single-file fallback (index.html).
func (o *Orchestrator) generateCodeFullStack(ctx context.Context, specification string, plan *MasterPlan, audit *ReverseEngineeringResult, manifest *SystemManifest, features []CompetitorFeature, imageURLs map[string]string) (map[string]string, error) {
	// ── Path 1: Chunked multi-file generation from FileMap ──
	if manifest != nil && len(manifest.FileMap) >= 5 {
		applog(ctx).InfoContext(ctx, "coder chunked path", "file_map_entries", len(manifest.FileMap))
		o.sendStatus(ctx, RoleCoder, "running", fmt.Sprintf("📦 Многофайловая генерация: %d файлов из архитектуры...", len(manifest.FileMap)), 42)

		files, err := o.generateCodeChunked(ctx, specification, manifest, plan, audit, features, imageURLs)
		if err == nil && len(files) > 0 {
			injectInspectorProvider(files)
			applog(ctx).InfoContext(ctx, "chunked coder success", "files", len(files))

			return files, nil
		}
		applog(ctx).WarnContext(ctx, "chunked coder failed, single-file fallback", "error", err)
		o.sendStatus(ctx, RoleCoder, "running", "⚠️ Переключение на монолитную генерацию...", 45)
	}

	// ── Path 2: Single-file fallback (index.html) ──
	manifestCtx := ""
	if manifest != nil {
		mj, err := json.Marshal(manifest)
		if err == nil && len(mj) > 100 {
			manifestCtx = "\n\nSYSTEM ARCHITECTURE MANIFEST:\n" + string(mj)
		}
	}

	synthesisCtx := ""
	if len(features) > 0 {
		var lines []string
		for _, f := range features {
			lines = append(lines, fmt.Sprintf("- [%s] %s: %s", f.Priority, f.Name, f.Description))
		}
		synthesisCtx = "\n\nCOMPETITOR FEATURES TO IMPLEMENT:\n" + strings.Join(lines, "\n")
	}

	backendCtx := backendTemplateContext(manifest)

	enrichedSpec := specification + manifestCtx + synthesisCtx
	if backendCtx != "" {
		enrichedSpec += "\n" + backendCtx
	}

	return o.generateCode(ctx, enrichedSpec, plan, audit, imageURLs)
}

type coderVisualContext struct {
	colors, components, design, tech string
}

func coderVisualContextFromAudit(audit *ReverseEngineeringResult) coderVisualContext {
	c := coderVisualContext{
		colors:     "#5b4cdb, #0e0e11, #ffffff",
		components: "Hero Section, Navigation, Feature Cards, Footer",
		design:     "Modern dark theme with glassmorphism effects",
		tech:       "HTML5, CSS3, Vanilla JavaScript",
	}
	if audit == nil {
		return c
	}
	if len(audit.Colors) > 0 {
		c.colors = strings.Join(audit.Colors, ", ")
	}
	if len(audit.Components) > 0 {
		c.components = strings.Join(audit.Components, ", ")
	}
	if audit.Layout != "" {
		c.design = audit.Layout
	}
	if len(audit.Technologies) == 0 {
		return c
	}
	end := min(len(audit.Technologies), 5)
	c.tech = strings.Join(audit.Technologies[:end], ", ")

	return c
}

// generateCode вызывает Coder с полным контекстом от Researcher + Director + Designer.
func (o *Orchestrator) generateCode(ctx context.Context, specification string, plan *MasterPlan, audit *ReverseEngineeringResult, imageURLs map[string]string) (map[string]string, error) {
	agent := o.agents[RoleCoder]
	ctx, cancel := context.WithTimeout(ctx, agent.Timeout)
	defer cancel()

	vis := coderVisualContextFromAudit(audit)
	colorCtx, componentCtx, designCtx, techCtx := vis.colors, vis.components, vis.design, vis.tech

	planSteps := specification
	if plan != nil && len(plan.Steps) > 0 {
		planSteps = strings.Join(plan.Steps, "\n")
	}

	// Build image context from Designer's visual core output
	imageCtx := ""
	if len(imageURLs) > 0 {
		imgLines := make([]string, 0, len(imageURLs))
		for key, url := range imageURLs {
			imgLines = append(imgLines, fmt.Sprintf("- %s: %s", key, url))
		}
		imageCtx = fmt.Sprintf(`
GENERATED IMAGES (from Designer):
%s
IMPORTANT: Use these REAL image URLs in <img> tags. Do NOT use placeholder images or unsplash.`, strings.Join(imgLines, "\n"))
	}

	userPrompt := fmt.Sprintf(`Build a PRODUCTION-READY web application with REAL functionality.

PROJECT: %s

DESIGN SYSTEM (from Researcher):
- Colors: %s
- Components: %s
- Layout: %s
- Tech: %s

IMPLEMENTATION PLAN (from Director):
%s
%s
CRITICAL REQUIREMENTS:
1. Self-contained index.html — ALL CSS and JS inline, renders in iframe
2. TailwindCSS CDN: <script src="https://cdn.tailwindcss.com"></script>
3. REAL JavaScript functionality — NOT just HTML markup:
   - Working forms with validation (addEventListener, preventDefault, real error messages)
   - Interactive elements: mobile hamburger menu, smooth scroll, modals, tabs
   - Business logic in JS: shopping cart with add/remove, price calculation, order total
   - localStorage for persistence (cart items, form data, user preferences)
   - Dynamic content rendering from JavaScript data arrays/objects
   - Toast notifications for user feedback (added to cart, form submitted, etc.)
4. REAL content for "%s" — NO Lorem Ipsum, NO placeholder text
5. Mobile-responsive with working hamburger menu (JS toggle)
6. Smooth CSS animations, transitions, hover effects
7. Professional typography with Google Fonts CDN

FUNCTIONALITY BY PROJECT TYPE (adapt to specification):
- Coffee shop/Restaurant: menu with categories and prices, "Add to Cart" buttons, cart sidebar with quantity +/-, order form with total calculation, working contact form with validation, opening hours section
- Online store: product grid from JS data, filter/sort, cart with localStorage, checkout form, quantity controls
- Portfolio/Agency: contact form with validation, project gallery with category filter, smooth scroll navigation
- SaaS/Landing: pricing toggle (monthly/yearly), FAQ accordion, lead capture form, feature comparison tabs
- Blog/News: article cards from JS data, category filter, search functionality, reading time estimate

STRUCTURE REQUIREMENTS:
- All event listeners via addEventListener (NO inline onclick)
- Organize JS: data objects at top, utility functions, component renderers, event handlers, init function
- Use semantic HTML5 tags (nav, main, section, article, footer)
- Include meta viewport tag for mobile

CRITICAL OUTPUT FORMAT — XML artifact protocol:
Wrap each file in <file path="..."> tags with raw unescaped code inside:
<file path="index.html">
<!DOCTYPE html><html lang="ru">...</html>
</file>

Output ONLY <file> blocks. No JSON. No markdown. No explanation.`,
		specification, colorCtx, componentCtx, designCtx, techCtx, planSteps, imageCtx, specification)
	applog(ctx).InfoContext(ctx, "coder single-file generation", "model", agent.Model)

	content, err := o.callLLMWithReasoning(ctx, agent.Model,
		`You are an elite full-stack web developer. You write FUNCTIONAL code, not just markup.
RULES:
- Every page MUST have real JavaScript: forms, interactivity, data rendering, cart logic.
- Use addEventListener for ALL events. No inline handlers.
- Store data in JS objects/arrays at the top of <script>. Render dynamically.
- Forms must validate inputs and show error/success messages.
- Shopping/ordering must calculate totals and persist in localStorage.
- CRITICAL: Output files using XML artifact tags: <file path="filename">raw code</file>
- NO JSON wrapping. NO escaping. NO markdown fences.`,
		userPrompt, 16000)
	if err != nil {
		applog(ctx).WarnContext(ctx, "coder primary model failed, fallback", "model", agent.Model, "error", err)
		content, err = o.callLLM(ctx, "qwen/qwen-2.5-72b-instruct",
			"You are an expert frontend developer. Output files using XML artifact tags: <file path=\"filename\">raw code</file>. No JSON. No markdown.",
			userPrompt, 16000)
		if err != nil {
			return nil, fmt.Errorf("code generation failed (both models): %w", err)
		}
	}

	files := o.parseCodeFiles(ctx, content)
	if len(files) == 0 {
		applog(ctx).WarnContext(ctx, "coder parse failed, extracting HTML")
		// Try to extract raw HTML if JSON parsing failed
		if idx := strings.Index(content, "<!DOCTYPE"); idx != -1 {
			files = map[string]string{"index.html": content[idx:]}
		} else if idx := strings.Index(content, "<html"); idx != -1 {
			files = map[string]string{"index.html": content[idx:]}
		} else {
			files = map[string]string{"index.html": content}
		}
	}
	applog(ctx).InfoContext(ctx, "coder files generated", "files", len(files))

	return files, nil
}

// translatePlanToBusiness вызывает LLM для перевода технического плана
// в понятный бизнес-язык (экраны, User Stories, функции) без технических терминов.
// При ошибке LLM — fallback на простое форматирование Steps.
func (o *Orchestrator) translatePlanToBusiness(ctx context.Context, specification string, plan *MasterPlan) string {
	// Собираем технический контекст для LLM
	techSummary := fmt.Sprintf("Architecture: %s\nComponents: %s\nSteps:\n",
		plan.Architecture, strings.Join(plan.Components, ", "))
	var techSummarySb1432 strings.Builder
	for i, s := range plan.Steps {
		fmt.Fprintf(&techSummarySb1432, "%d. %s\n", i+1, s)
	}
	techSummary += techSummarySb1432.String()

	systemPrompt := `Ты — Бизнес-Аналитик (Product Manager). Твоя задача — перевести технический план разработки в понятный для обычного человека список функций будущего приложения.

СТРОГИЕ ПРАВИЛА:
1. ЗАПРЕЩЕНО использовать технические термины: БД, SQL, React, API, TypeScript, компонент, эндпоинт, бэкенд, фронтенд, роутинг, стейт, хук, Redux, Vite, Tailwind, SSR, CDN.
2. Пиши ТОЛЬКО на русском языке.
3. Описывай ЭКРАНЫ приложения (что увидит пользователь).
4. Описывай ВОЗМОЖНОСТИ пользователей (User Stories): "Пользователь сможет..."
5. Описывай БИЗНЕС-ЛОГИКУ простыми словами.
6. В конце добавь вопрос: "Хотите добавить или убрать какие-то функции перед началом разработки?"
7. Формат: используй эмодзи для разделов, пиши кратко и по делу.

ФОРМАТ ОТВЕТА:
📱 Экраны приложения
• [Название экрана] — [что на нём]

👤 Что сможет делать пользователь
• [User Story простыми словами]

⚙️ Бизнес-логика
• [Что будет работать "под капотом" простыми словами]

❓ Хотите добавить или убрать какие-то функции?`

	userPrompt := fmt.Sprintf("Исходная идея заказчика:\n%s\n\nТехнический план (перепиши на бизнес-язык):\n%s",
		specification, techSummary)

	result, err := o.callLLM(ctx, "openai/gpt-4.1-mini",
		systemPrompt, userPrompt, 2048)
	if err != nil {
		applog(ctx).WarnContext(ctx, "translatePlanToBusiness failed, fallback", "error", err)
		// Fallback: простое форматирование без технических деталей
		fallback := "📋 Функции вашего приложения:\n\n"
		var fallbackSb1468 strings.Builder
		for i, s := range plan.Steps {
			fmt.Fprintf(&fallbackSb1468, "%d. %s\n", i+1, s)
		}
		fallback += fallbackSb1468.String()
		fallback += "\n❓ Хотите добавить или убрать какие-то функции перед началом разработки?"

		return fallback
	}

	return result
}

// sendStatus отправляет статус в шину событий сессии (из ctx).
func (o *Orchestrator) sendStatus(ctx context.Context, agent AgentRole, _ string, message string, progress int) {
	o.busFromCtx(ctx).PublishStatus(agent, "", message, progress)
}

// busFromCtx возвращает шину событий сессии из ctx.
// Если sessionID отсутствует/не зарегистрирован — возвращает дефолтную шину.
func (o *Orchestrator) busFromCtx(ctx context.Context) *domain.EventBus {
	sessionID, _ := ctx.Value(sessionIDKey{}).(string)
	if sessionID != "" {
		if bus := o.buses.get(sessionID); bus != nil {
			return bus
		}
	}

	return o.events
}

// buildMediaGuidelines creates a markdown instruction block from approved media assets.
// This block is injected into the specification so Coder agents use real URLs and video placeholders.
func buildMediaGuidelines(approvedAssets []domain.MediaAsset, imageURLs map[string]string) string {
	if len(approvedAssets) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "\n\n---\n## 🎨 MEDIA GUIDELINES (ОБЯЗАТЕЛЬНО К ИСПОЛНЕНИЮ)")
	lines = append(lines, "ТЫ ОБЯЗАН использовать указанные ниже URL-адреса для изображений и создать UI-блок для видео-заглушки.")
	lines = append(lines, "НЕ используй placeholder-картинки, unsplash random, или picsum. Только URL ниже.\n")

	hasMedia := false
	for _, a := range approvedAssets {
		if a.Type == "image" {
			// Prefer AI-generated URL from imageURLs map, fallback to PreviewURL (stock)
			url := a.PreviewURL
			if aiURL, ok := imageURLs[a.Placement]; ok && aiURL != "" {
				url = aiURL
			}
			if url == "" {
				continue
			}
			lines = append(lines, fmt.Sprintf("* **Изображение** | Плейсмент: `%s` | URL: `%s` | Описание: %s",
				a.Placement, url, a.Prompt))
			hasMedia = true
		} else if a.Type == "video" {
			lines = append(lines, fmt.Sprintf("* **ВИДЕО-ЗАГЛУШКА** | Плейсмент: `%s` | Описание сценария: %s",
				a.Placement, a.Prompt))
			lines = append(lines, "  → Вставь красивый UI-компонент видеоплеера с надписью «Premium AI Video — Coming Soon».")
			lines = append(lines, "  → Стиль: тёмный фон с градиентом, иконка ▶️ по центру, текст сценария под плеером.")
			hasMedia = true
		}
	}

	if !hasMedia {
		return ""
	}

	lines = append(lines, "\n**ПРАВИЛА:**")
	lines = append(lines, "1. Для каждого изображения используй <img src=\"{URL}\" alt=\"...\" class=\"...\" /> с указанным URL.")
	lines = append(lines, "2. Hero-изображение — на полную ширину секции, object-cover, max-height: 500px.")
	lines = append(lines, "3. OG-изображение — добавь <meta property=\"og:image\" content=\"{URL}\" /> в <head>.")
	lines = append(lines, "4. Видео-заглушка — div с aspect-ratio 16/9, тёмный gradient overlay, иконка play, текст 'Premium AI Video'.")
	lines = append(lines, "")
	lines = append(lines, "ВАЖНО: Эти Media Guidelines НЕ отменяют базовый формат вывода. Ты ОБЯЗАН вернуть код в XML-тегах: <file path=\"...\">код</file>. Используй указанные URL прямо в сыром коде внутри тегов.")
	lines = append(lines, "---")

	return strings.Join(lines, "\n")
}

// stockKeywords extracts 2-3 search keywords from an AI image prompt for Unsplash.
func stockKeywords(prompt string) string {
	lower := strings.ToLower(prompt)
	stopwords := map[string]bool{
		"photorealistic": true, "cinematic": true, "8k": true, "4k": true,
		"ultra": true, "detailed": true, "professional": true, "studio": true,
		"lighting": true, "high": true, "quality": true, "beautiful": true,
		"modern": true, "futuristic": true, "dark": true, "gradient": true,
		"mesh": true, "for": true, "the": true, "and": true, "with": true,
		"app": true, "background": true, "preview": true, "social": true,
		"image": true, "prompt": true, "tones": true, "theme": true,
	}
	words := strings.Fields(lower)
	var kw []string
	for _, w := range words {
		w = strings.Trim(w, ".,!?;:\"'()-")
		if len(w) > 2 && !stopwords[w] && len(kw) < 3 {
			kw = append(kw, w)
		}
	}
	if len(kw) == 0 {
		return "abstract,technology"
	}

	return strings.Join(kw, ",")
}
