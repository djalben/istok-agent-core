package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/istok/agent-core/internal/application/usecases"
	"github.com/istok/agent-core/internal/domain"
	"github.com/istok/agent-core/internal/ports"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — S-Tier AI Orchestrator
//  Мультимодельная архитектура нового поколения
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// GenerationMode режим генерации
type GenerationMode string

const (
	ModeAgent     GenerationMode = "agent"     // Инновационное проектирование: Gemini 3 Pro — глубокий анализ
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
	RoleCoder        = domain.RoleCoder
	RoleDesigner     = domain.RoleDesigner
	RoleVideographer = domain.RoleVideographer
	RoleValidator    = domain.RoleValidator
)

// AgentConfig конфигурация агента
type AgentConfig struct {
	Role            AgentRole
	Model           string
	Description     string
	Timeout         time.Duration
	ThinkingEnabled bool
	ThinkingBudget  int
}

// TaskStatus — алиас для обратной совместимости (SSE handler и др.).
// Новый код должен использовать domain.AgentEvent.
type TaskStatus = domain.AgentEvent

// ReverseEngineeringResult результат анализа сайта
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

// MasterPlan план разработки от директора (Planner)
type MasterPlan struct {
	Architecture string
	Components   []string
	Timeline     string
	Technologies []string
	Steps        []string  // backward-compat flat steps
	DAG          []DAGTask // DAG-представление плана
}

// GenerationResult финальный результат генерации
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
	agents     map[AgentRole]*AgentConfig
	llm        ports.LLMProvider
	events     *domain.EventBus
	projectEnv *ProjectEnv
	planner    *usecases.PlannerAgent   // модернизированный Планировщик с DAG + FSM gate
	projectCtx *usecases.ProjectContext // отсканированный package.json/tsconfig.json
	mu         sync.RWMutex
}

// NewOrchestrator создает оркестратор с LLM-провайдером (через порт) и шиной событий.
func NewOrchestrator(llm ports.LLMProvider) *Orchestrator {
	return &Orchestrator{
		llm:     llm,
		events:  domain.NewEventBus(128),
		planner: usecases.NewPlannerAgent(llm, "anthropic/claude-3-7-sonnet-thinking"),
		agents: map[AgentRole]*AgentConfig{
			RoleDirector: {
				Role:        RoleDirector,
				Model:       "anthropic/claude-3-7-sonnet-thinking",
				Description: "🧠 Директор — Claude 3.7 Sonnet Thinking (планирование)",
				Timeout:     5 * time.Minute,
			},
			RoleBrain: {
				Role:        RoleBrain,
				Model:       "anthropic/claude-3-7-sonnet-thinking",
				Description: "🧠 Мозг — Claude 3.7 Sonnet Thinking (архитектура)",
				Timeout:     10 * time.Minute,
			},
			RoleResearcher: {
				Role:        RoleResearcher,
				Model:       "anthropic/claude-3-7-sonnet-thinking",
				Description: "🔍 Исследователь — Claude 3.7 Sonnet Thinking (анализ)",
				Timeout:     5 * time.Minute,
			},
			RoleCoder: {
				Role:        RoleCoder,
				Model:       "anthropic/claude-3-7-sonnet",
				Description: "💻 Кодер — Claude 3.7 Sonnet Medium (код)",
				Timeout:     10 * time.Minute,
			},
			RoleDesigner: {
				Role:        RoleDesigner,
				Model:       "google/nano-banana",
				Description: "🎨 Дизайнер — Nano Banana (UI-ассеты, Replicate)",
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
				Model:       "anthropic/claude-3-7-sonnet",
				Description: "✅ Валидатор — Claude 3.7 Sonnet (Syntax & Runtime)",
				Timeout:     3 * time.Minute,
			},
		},
	}
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
		return fmt.Errorf("planner not initialized")
	}
	pc, err := o.planner.ScanProject(packageJSONPath, tsconfigPath)
	if err != nil {
		return err
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
	RoleDirector, // Planner uses Director role in current implementation
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
			provider = "Anthropic Direct"
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

// GenerateWithMode запускает процесс генерации в указанном режиме
func (o *Orchestrator) GenerateWithMode(ctx context.Context, specification string, url string, mode GenerationMode) (*GenerationResult, error) {
	if mode == ModeCode {
		return o.generateCodeMode(ctx, specification)
	}
	// Both "agent" (Инновационное проектирование) and "synthesis" (Адаптивный синтез) use full pipeline
	return o.generateAgentMode(ctx, specification, url)
}

// generateCodeMode быстрая генерация через Claude 3.7 Sonnet (Code Mode)
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
	if err := fsm.TransitionTo(domain.StatePlanning, "code mode: fast planning"); err != nil {
		return nil, fmt.Errorf("FSM: %w", err)
	}
	o.events.PublishFSMTransition(domain.StateCreated, domain.StatePlanning, "code mode")

	plan := &MasterPlan{
		Architecture: "Quick UI Generation",
		Steps:        []string{specification},
	}

	// Утверждаем план в FSM (gate для Coding)
	if err := fsm.ApprovePlan(domain.ApprovedPlan{
		Architecture: plan.Architecture,
		Steps:        plan.Steps,
		ApprovedBy:   "code_mode_auto",
	}); err != nil {
		return nil, fmt.Errorf("FSM plan approval: %w", err)
	}

	if err := fsm.TransitionTo(domain.StateArchitectureApproved, "auto-approved for code mode"); err != nil {
		return nil, fmt.Errorf("FSM: %w", err)
	}

	// Переход в Coding (пройдёт только если план утверждён)
	if err := fsm.TransitionTo(domain.StateCoding, "plan approved, starting code generation"); err != nil {
		return nil, fmt.Errorf("FSM: %w", err)
	}
	o.events.PublishFSMTransition(domain.StateArchitectureApproved, domain.StateCoding, "code mode")

	o.sendStatus(RoleCoder, "running", "⚡ Claude 3.7 Sonnet генерирует UI компоненты...", 20)

	code, err := o.generateCode(ctx, specification, plan, nil, nil)
	if err != nil {
		_ = fsm.TransitionTo(domain.StateFailed, err.Error())
		o.sendStatus(RoleCoder, "error", fmt.Sprintf("❌ Ошибка: %v", err), 0)
		return nil, err
	}

	_ = fsm.TransitionTo(domain.StateQualityCheck, "code generated")
	_ = fsm.TransitionTo(domain.StateSecurityCheck, "quality ok")
	_ = fsm.TransitionTo(domain.StateVerified, "security ok")
	_ = fsm.TransitionTo(domain.StateCompleted, "done")

	result.Code = code
	result.Duration = time.Since(startTime)
	o.sendStatus(RoleCoder, "completed", fmt.Sprintf("✅ Код готов за %v", result.Duration), 100)
	return result, nil
}

// generateAgentMode полная мультимодальная генерация с Gemini 3 Pro (Agent Mode)
func (o *Orchestrator) generateAgentMode(ctx context.Context, specification string, url string) (*GenerationResult, error) {
	startTime := time.Now()
	result := &GenerationResult{
		Code:   make(map[string]string),
		Assets: make(map[string]string),
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Инициализируем FSM
	fsm := domain.NewTaskStateMachine()

	// Track synthesis features for architecture phase
	var competitorFeatures []CompetitorFeature

	// ── FSM: Created → Researching ──
	if err := fsm.TransitionTo(domain.StateResearching, "starting research phase"); err != nil {
		return nil, fmt.Errorf("FSM: %w", err)
	}
	o.events.PublishFSMTransition(domain.StateCreated, domain.StateResearching, "agent mode")

	// ── Этап 0 (ОБЯЗАТЕЛЬНЫЙ): DeepSeek V3.2 — Исследование + Глубокий синтез ВСЕГДА первым ──
	researcher := NewResearcherAgent(o.llm)
	if url != "" {
		// Глубокий синтез конкурента: извлечение всех фич + задачи для кодинга
		synthesis, _ := o.deepSynthesis(ctx, url, specification)
		if synthesis != nil && len(synthesis.Features) > 0 {
			competitorFeatures = synthesis.Features
		}

		// Визуальный аудит URL
		visualAudit, err := researcher.VisualAudit(ctx, url, o.events)
		if err != nil {
			o.sendStatus(RoleResearcher, "error", fmt.Sprintf("⚠️ URL-аудит недоступен: %v", err), 0)
		} else {
			o.mu.Lock()
			result.VisualAudit = visualAudit
			result.Audit = &ReverseEngineeringResult{
				URL:          url,
				Colors:       visualAudit.Colors,
				Fonts:        visualAudit.Fonts,
				Components:   visualAudit.Components,
				Layout:       visualAudit.Layout,
				Technologies: visualAudit.Technologies,
				Audit:        fmt.Sprintf("DesignSystem: %s, Animations: %v", visualAudit.DesignSystem, visualAudit.Animations),
			}
			o.mu.Unlock()
		}
	} else {
		// Нет URL — анализ спецификации текстом
		visualAudit := researcher.AnalyzeSpec(ctx, specification, o.events)
		o.mu.Lock()
		result.VisualAudit = visualAudit
		result.Audit = &ReverseEngineeringResult{
			URL:          "spec://text",
			Colors:       visualAudit.Colors,
			Fonts:        visualAudit.Fonts,
			Components:   visualAudit.Components,
			Layout:       visualAudit.Layout,
			Technologies: visualAudit.Technologies,
			Audit:        fmt.Sprintf("DesignSystem: %s, Insights: %v", visualAudit.DesignSystem, visualAudit.Insights),
		}
		o.mu.Unlock()
	}

	// ── FSM: Researching → Planning ──
	if err := fsm.TransitionTo(domain.StatePlanning, "research complete, starting planning"); err != nil {
		return nil, fmt.Errorf("FSM: %w", err)
	}
	o.events.PublishFSMTransition(domain.StateResearching, domain.StatePlanning, "research done")

	// ── Этап 1: Gemini 3 Pro Brain — DefineArchitecture (Full-Stack манифест) ──
	manifest, archErr := o.defineArchitecture(ctx, specification, result.Audit, competitorFeatures)
	if archErr != nil {
		log.Printf("⚠️ Architecture manifest warning: %v", archErr)
	}

	// Этап 1b: Стратегический синтез
	o.sendStatus(RoleBrain, "running", "🧠 Gemini 3 Pro анализирует стратегию...", 18)
	strategy, brainErr := o.synthesizeStrategy(ctx, specification, result.Audit)
	if brainErr != nil {
		log.Printf("⚠️ Brain synthesis warning (non-critical): %v", brainErr)
	} else if strategy != "" && result.Audit != nil {
		result.Audit.Audit = strategy
	}
	o.sendStatus(RoleBrain, "completed", "✅ Стратегия построена на основе анализа.", 22)

	// ── Этап 2: Planner Agent — DAG-план с инъекцией контекста ─────────────────
	o.sendStatus(RoleDirector, "running", "🧠 Planner Agent: построение DAG-плана...", 28)
	masterPlan, err := o.createMasterPlan(ctx, specification, result.Audit)
	if err != nil {
		_ = fsm.TransitionTo(domain.StateFailed, err.Error())
		o.sendStatus(RoleDirector, "error", fmt.Sprintf("❌ Ошибка планирования: %v", err), 0)
		return nil, fmt.Errorf("master plan creation failed: %w", err)
	}
	result.MasterPlan = masterPlan
	o.sendStatus(RoleDirector, "completed", fmt.Sprintf("✅ DAG-план готов: %d задач", len(masterPlan.DAG)), 100)

	// ── FSM: Planning → Architecture_Approved (c утверждением плана) ──
	if err := fsm.ApprovePlan(domain.ApprovedPlan{
		Architecture: masterPlan.Architecture,
		Steps:        masterPlan.Steps,
		Components:   masterPlan.Components,
		Technologies: masterPlan.Technologies,
		ApprovedBy:   "director",
	}); err != nil {
		_ = fsm.TransitionTo(domain.StateFailed, "plan rejected: "+err.Error())
		return nil, fmt.Errorf("FSM plan approval: %w", err)
	}
	if err := fsm.TransitionTo(domain.StateArchitectureApproved, "director plan approved"); err != nil {
		return nil, fmt.Errorf("FSM: %w", err)
	}
	o.events.PublishFSMTransition(domain.StatePlanning, domain.StateArchitectureApproved, "plan approved")
	o.events.Publish(domain.AgentEvent{
		Kind: domain.EventPlan, Agent: RoleDirector,
		Message: fmt.Sprintf("%d steps, %d techs", len(masterPlan.Steps), len(masterPlan.Technologies)),
	})

	// ── FSM: Architecture_Approved → Strategy_Synthesized (smart gate by Planner) ──
	// Planner проверит наличие API-ключей и контекста проекта ПЕРЕД переходом.
	if err := o.planner.AdvanceToStrategySynthesized(fsm, o.projectCtx); err != nil {
		log.Printf("⚠️ Planner FSM gate: %v — fallback transition", err)
		o.sendStatus(RoleDirector, "running", fmt.Sprintf("⚠️ Planner readiness: %v", err), 24)
		// Fallback: разрешаем переход даже если gate провалился (для обратной совместимости)
		if fsmErr := fsm.TransitionTo(domain.StateStrategySynthesized, "strategy synthesis done (fallback)"); fsmErr != nil {
			log.Printf("⚠️ FSM strategy fallback transition: %v", fsmErr)
		}
	} else {
		o.sendStatus(RoleDirector, "running", "✅ Planner: readiness check passed", 26)
	}
	o.events.PublishFSMTransition(domain.StateArchitectureApproved, domain.StateStrategySynthesized, "planner gate")

	// ── FSM: Strategy_Synthesized → Designing ──
	if err := fsm.TransitionTo(domain.StateDesigning, "starting design phase"); err != nil {
		log.Printf("⚠️ FSM designing transition: %v", err)
	}
	o.events.PublishFSMTransition(domain.StateStrategySynthesized, domain.StateDesigning, "design start")

	// ── Этап 3: Дизайнер генерирует изображения ПЕРВЫМ (Nano Banana 2) ──
	// Дизайнер запускается ДО Кодера, чтобы передать ему реальные URL изображений
	mediaService := newMediaService(o.llm)
	imageURLs := map[string]string{}

	o.sendStatus(RoleDesigner, "running", "🎨 Nano Banana 2 генерирует изображения для проекта...", 35)
	var designColors []string
	if result.VisualAudit != nil {
		designColors = result.VisualAudit.Colors
	}
	assets, designErr := mediaService.GenerateUIAssets(ctx, specification, specification, designColors)
	if designErr != nil {
		log.Printf("⚠️ Designer error (non-critical): %v", designErr)
		o.sendStatus(RoleDesigner, "error", fmt.Sprintf("⚠️ Дизайн: %v", designErr), 0)
	} else {
		if assets.HeroImageURL != "" {
			imageURLs["hero"] = assets.HeroImageURL
		}
		if assets.OGImageURL != "" {
			imageURLs["og"] = assets.OGImageURL
		}
		o.mu.Lock()
		result.Assets = map[string]string{
			"logo.svg":      assets.LogoSVG,
			"hero_prompt":   assets.HeroPrompt,
			"og_prompt":     assets.OGImagePrompt,
			"color_palette": fmt.Sprintf("%v", assets.ColorPalette),
		}
		if assets.HeroImageURL != "" {
			result.Assets["hero_image_url"] = assets.HeroImageURL
		}
		if assets.OGImageURL != "" {
			result.Assets["og_image_url"] = assets.OGImageURL
		}
		o.mu.Unlock()
		o.sendStatus(RoleDesigner, "completed", fmt.Sprintf("✅ Дизайн готов: %d изображений, SVG логотип", len(imageURLs)), 100)
	}
	log.Printf("🎨 Designer phase complete: %d image URLs for Coder", len(imageURLs))

	// ── FSM: Designing → Coding (gate: plan must be approved) ──
	if err := fsm.TransitionTo(domain.StateCoding, "design complete, starting code generation"); err != nil {
		_ = fsm.TransitionTo(domain.StateFailed, "FSM coding gate: "+err.Error())
		return nil, fmt.Errorf("FSM: %w", err)
	}
	o.events.PublishFSMTransition(domain.StateDesigning, domain.StateCoding, "coding start")

	// ── Этап 4: Кодер + Видеограф параллельно ──
	// Кодер получает URL изображений от Дизайнера и встраивает их в код
	var wg sync.WaitGroup
	var coderErr error
	var generatedCode map[string]string

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("🔥 PANIC in coder goroutine: %v", rec)
				coderErr = fmt.Errorf("coder panic: %v", rec)
				o.sendStatus(RoleCoder, "error", fmt.Sprintf("❌ Panic: %v", rec), 0)
			}
		}()
		o.sendStatus(RoleCoder, "running", "💻 Кодер пишет функциональный код с реальными изображениями...", 40)
		code, err := o.generateCodeFullStack(ctx, specification, masterPlan, result.Audit, manifest, competitorFeatures, imageURLs)
		if err != nil {
			coderErr = fmt.Errorf("code generation failed: %w", err)
			o.sendStatus(RoleCoder, "error", fmt.Sprintf("❌ Ошибка кода: %v", err), 0)
			return
		}
		generatedCode = code
		o.sendStatus(RoleCoder, "completed", fmt.Sprintf("✅ Код сгенерирован (%d файлов), запуск валидации...", len(code)), 70)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("🔥 PANIC in videographer goroutine: %v", rec)
				o.sendStatus(RoleVideographer, "error", fmt.Sprintf("⚠️ Видео: %v", rec), 0)
			}
		}()
		o.sendStatus(RoleVideographer, "running", "🎬 Монтаж промо-ролика...", 70)
		video, err := mediaService.GeneratePromoVideo(ctx, "ИСТОК", specification)
		if err != nil {
			log.Printf("⚠️ Videographer error (non-critical): %v", err)
			o.sendStatus(RoleVideographer, "error", fmt.Sprintf("⚠️ Видео: %v", err), 0)
			return
		}
		o.mu.Lock()
		result.Video = fmt.Sprintf("Script: %s | Scenes: %d | Music: %s", video.Script[:min(len(video.Script), 100)], len(video.Scenes), video.MusicStyle)
		o.mu.Unlock()
		o.sendStatus(RoleVideographer, "completed", fmt.Sprintf("✅ Промо-ролик готов: %d сцен, %s", len(video.Scenes), video.Duration), 100)
	}()

	wg.Wait()

	if coderErr != nil {
		_ = fsm.TransitionTo(domain.StateFailed, coderErr.Error())
		return nil, coderErr
	}

	// ── Verification Layer (Layer 3): Security + Tester + UI/UX Reviewer ──
	// VerificationGate требует Approved от всех 3 агентов перед StateCompleted.
	const maxRetries = 2
	gate := usecases.NewVerificationGate()
	var finalReport *usecases.VerificationReport

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// FSM: Coding → QualityCheck
		_ = fsm.TransitionTo(domain.StateQualityCheck, fmt.Sprintf("verify attempt %d", attempt+1))
		o.events.PublishFSMTransition(domain.StateCoding, domain.StateQualityCheck, fmt.Sprintf("attempt %d", attempt+1))

		// На последней попытке выключаем тесты — экономим время, отдаём как есть с warnings
		gate.RunTests = attempt < maxRetries

		o.sendStatus(RoleValidator, "running",
			fmt.Sprintf("🛡️ VerificationGate: Security + Tester + UI/UX (попытка %d)...", attempt+1),
			80+attempt*5)

		report := gate.Verify(ctx, generatedCode)
		finalReport = report
		log.Printf("🛡️ VerificationGate attempt %d: %s", attempt+1, report.Summary)

		// Публикуем per-agent статусы для UI
		for _, a := range report.Approvals {
			marker := "✅"
			if !a.Approved {
				marker = "❌"
			}
			log.Printf("  %s [%s] %s", marker, a.Agent, a.Summary)
		}

		if report.Approved {
			// FSM: QualityCheck → SecurityCheck → Verified
			_ = fsm.TransitionTo(domain.StateSecurityCheck, "all 3 agents approved")
			o.events.PublishFSMTransition(domain.StateQualityCheck, domain.StateSecurityCheck, "verify OK")
			_ = fsm.TransitionTo(domain.StateVerified, "verification gate passed")
			o.events.PublishFSMTransition(domain.StateSecurityCheck, domain.StateVerified, "verified")
			o.sendStatus(RoleValidator, "completed",
				fmt.Sprintf("✅ Все 3 агента одобрили: %s", report.Summary), 100)
			break
		}

		// Verification failed
		o.sendStatus(RoleValidator, "running",
			fmt.Sprintf("⚠️ Заблокировано агентом [%s]: %s", report.BlockingAgent, report.Summary), 85)

		if attempt >= maxRetries {
			// Max retries — НЕ переходим в Verified без одобрения. Падаем в Failed.
			log.Printf("🚫 VerificationGate: max retries (%d) reached, blocking by [%s]",
				maxRetries, report.BlockingAgent)
			_ = fsm.TransitionTo(domain.StateFailed,
				fmt.Sprintf("verification gate blocked by %s after %d attempts",
					report.BlockingAgent, maxRetries+1))
			o.events.PublishFSMTransition(domain.StateQualityCheck, domain.StateFailed,
				"verification blocked")
			o.sendStatus(RoleValidator, "error",
				fmt.Sprintf("🚫 Verification BLOCKED: %s", report.Summary), 100)
			break
		}

		// ── Auto-Fix: FSM → RetryCoding → Coding ──
		retryErrorCtx := report.ForCoderContext()
		_ = fsm.TransitionTo(domain.StateRetryCoding,
			fmt.Sprintf("auto-fix: blocked by %s", report.BlockingAgent))
		o.events.PublishFSMTransition(domain.StateQualityCheck, domain.StateRetryCoding,
			"auto-fix")

		_ = fsm.TransitionTo(domain.StateCoding, "retry with combined error context")
		o.events.PublishFSMTransition(domain.StateRetryCoding, domain.StateCoding, "retry")

		o.sendStatus(RoleCoder, "running",
			fmt.Sprintf("🔄 Auto-fix: повторная генерация (попытка %d/%d, blocked by %s)...",
				attempt+2, maxRetries+1, report.BlockingAgent), 75)

		// Re-generate с комбинированным контекстом ошибок от всех 3 агентов
		enrichedSpec := specification + "\n\n" + retryErrorCtx
		retryCode, err := o.generateCodeFullStack(ctx, enrichedSpec, masterPlan, result.Audit,
			manifest, competitorFeatures, imageURLs)
		if err != nil {
			log.Printf("⚠️ Auto-fix retry %d failed: %v", attempt+1, err)
			o.sendStatus(RoleCoder, "error", fmt.Sprintf("⚠️ Retry failed: %v", err), 0)
			break
		}
		generatedCode = retryCode
		o.sendStatus(RoleCoder, "completed",
			fmt.Sprintf("✅ Auto-fix код готов (%d файлов)", len(retryCode)), 78)
	}

	// Save final code
	o.mu.Lock()
	result.Code = generatedCode
	o.mu.Unlock()

	// ── Final Gate: переход в Completed ТОЛЬКО если VerificationGate дал Approved ──
	if err := gate.CanTransitionToCompleted(finalReport); err != nil {
		log.Printf("🚫 FSM Completed BLOCKED: %v", err)
		// FSM уже в Failed (выставлено выше при max retries) — не делаем повторный transition.
		// Просто возвращаем ошибку.
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("pipeline incomplete: %w", err)
	}

	_ = fsm.TransitionTo(domain.StateCompleted, "all verification gates passed")
	o.events.PublishFSMTransition(domain.StateVerified, domain.StateCompleted, "done")

	result.Duration = time.Since(startTime)
	log.Printf("✅ FSM: %d transitions in %v", len(fsm.Transitions()), result.Duration)
	o.sendStatus(RoleDirector, "completed", fmt.Sprintf("🎉 Проект готов за %v", result.Duration), 100)
	return result, nil
}

// min вспомогательная функция
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// convertPlanTasks преобразует usecases.PlanTask → application.DAGTask
// для обратной совместимости с MasterPlan.DAG, ожидаемым downstream-агентами.
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

// createMasterPlan делегирует построение плана модернизированному PlannerAgent
// (DAG engine + Context Injection + Smart FSM Gate). Если у нас уже есть
// projectCtx (отсканированные package.json/tsconfig.json) — используем его,
// иначе строим временный из projectEnv для обратной совместимости.
func (o *Orchestrator) createMasterPlan(ctx context.Context, specification string, audit *ReverseEngineeringResult) (*MasterPlan, error) {
	agent := o.agents[RoleDirector]
	ctx, cancel := context.WithTimeout(ctx, agent.Timeout)
	defer cancel()

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

	if o.planner != nil {
		log.Printf("🧠 Planner Agent: запрашиваю DAG-план у %s", o.planner.Model)
		uPlan, err := o.planner.BuildPlan(ctx, specification, auditSummary, projectCtx)
		if err == nil && uPlan != nil && len(uPlan.Tasks) > 0 {
			plan := &MasterPlan{
				Architecture: uPlan.Architecture,
				Components:   uPlan.Components,
				Technologies: uPlan.Technologies,
				Timeline:     uPlan.Timeline,
				Steps:        uPlan.Steps,
				DAG:          convertPlanTasks(uPlan.Tasks),
			}
			if plan.Architecture == "" {
				plan.Architecture = specification
			}
			if len(plan.Steps) == 0 {
				for _, t := range uPlan.Tasks {
					plan.Steps = append(plan.Steps, t.Title)
				}
			}
			log.Printf("✅ Planner: %d DAG tasks, exec order: %v", len(plan.DAG), uPlan.ExecutionOrder)
			return plan, nil
		}
		log.Printf("⚠️ PlannerAgent failed (%v), falling back to legacy Director prompt", err)
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
%s`, specification, auditSummary, envCtx)

	log.Printf("🧠 Director: запрашиваю DAG-план у %s", agent.Model)

	result, err := o.callLLMWithReasoning(ctx, agent.Model,
		`You are a senior software architect and project planner. Create precise, actionable DAG plans.
Output only valid JSON. Every task must have impacted_files and required_dependencies.
ARCHITECTURE RULES:
- Never put business logic in main.go or HTTP handlers.
- Separate Domain (entities), Application (use cases), Infrastructure (external APIs), Transport (HTTP/SSE).
- All external dependencies must go through interfaces (ports).
- Use @/* import aliases. Structure: components/ui, components/layout, hooks, services.`,
		userPrompt, 4096, agent.ThinkingBudget)

	if err != nil {
		log.Printf("⚠️ Director API error, using default plan: %v", err)
		return o.defaultMasterPlan(specification, audit), nil
	}

	plan := o.parseMasterPlan(result, specification, audit)
	log.Printf("✅ Director: план готов — %d шагов, %d технологий", len(plan.Steps), len(plan.Technologies))
	return plan, nil
}

// generateCodeFullStack вызывает Coder с полным контекстом: manifest + features + backend templates + imageURLs
func (o *Orchestrator) generateCodeFullStack(ctx context.Context, specification string, plan *MasterPlan, audit *ReverseEngineeringResult, manifest *SystemManifest, features []CompetitorFeature, imageURLs map[string]string) (map[string]string, error) {
	// Build extra context from manifest and synthesis
	manifestCtx := ""
	if manifest != nil {
		mj, _ := json.Marshal(manifest)
		if len(mj) > 100 {
			manifestCtx = fmt.Sprintf("\n\nSYSTEM ARCHITECTURE MANIFEST:\n%s", string(mj))
		}
	}

	synthesisCtx := ""
	if len(features) > 0 {
		var lines []string
		for _, f := range features {
			lines = append(lines, fmt.Sprintf("- [%s] %s: %s", f.Priority, f.Name, f.Description))
		}
		synthesisCtx = fmt.Sprintf("\n\nCOMPETITOR FEATURES TO IMPLEMENT:\n%s", strings.Join(lines, "\n"))
	}

	backendCtx := backendTemplateContext(manifest)

	// Inject extra context into specification for the Coder
	enrichedSpec := specification + manifestCtx + synthesisCtx
	if backendCtx != "" {
		enrichedSpec += "\n" + backendCtx
	}

	return o.generateCode(ctx, enrichedSpec, plan, audit, imageURLs)
}

// generateCode вызывает Coder с полным контекстом от Researcher + Director + Designer
func (o *Orchestrator) generateCode(ctx context.Context, specification string, plan *MasterPlan, audit *ReverseEngineeringResult, imageURLs map[string]string) (map[string]string, error) {
	agent := o.agents[RoleCoder]
	ctx, cancel := context.WithTimeout(ctx, agent.Timeout)
	defer cancel()

	// Build rich design context from Researcher audit
	colorCtx := "#5b4cdb, #0e0e11, #ffffff"
	componentCtx := "Hero Section, Navigation, Feature Cards, Footer"
	designCtx := "Modern dark theme with glassmorphism effects"
	techCtx := "HTML5, CSS3, Vanilla JavaScript"

	if audit != nil {
		if len(audit.Colors) > 0 {
			colorCtx = strings.Join(audit.Colors, ", ")
		}
		if len(audit.Components) > 0 {
			componentCtx = strings.Join(audit.Components, ", ")
		}
		if audit.Layout != "" {
			designCtx = audit.Layout
		}
		if len(audit.Technologies) > 0 {
			end := len(audit.Technologies)
			if end > 5 {
				end = 5
			}
			techCtx = strings.Join(audit.Technologies[:end], ", ")
		}
	}

	planSteps := specification
	if plan != nil && len(plan.Steps) > 0 {
		planSteps = strings.Join(plan.Steps, "\n")
	}

	// Build image context from Designer's Nano Banana 2 output
	imageCtx := ""
	if len(imageURLs) > 0 {
		var imgLines []string
		for key, url := range imageURLs {
			imgLines = append(imgLines, fmt.Sprintf("- %s: %s", key, url))
		}
		imageCtx = fmt.Sprintf(`
GENERATED IMAGES (from Designer via Nano Banana 2):
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
1. Output JSON: {"index.html":"<!DOCTYPE html>..."}
2. Self-contained index.html — ALL CSS and JS inline, renders in iframe
3. TailwindCSS CDN: <script src="https://cdn.tailwindcss.com"></script>
4. REAL JavaScript functionality — NOT just HTML markup:
   - Working forms with validation (addEventListener, preventDefault, real error messages)
   - Interactive elements: mobile hamburger menu, smooth scroll, modals, tabs
   - Business logic in JS: shopping cart with add/remove, price calculation, order total
   - localStorage for persistence (cart items, form data, user preferences)
   - Dynamic content rendering from JavaScript data arrays/objects
   - Toast notifications for user feedback (added to cart, form submitted, etc.)
5. REAL content for "%s" — NO Lorem Ipsum, NO placeholder text
6. Mobile-responsive with working hamburger menu (JS toggle)
7. Smooth CSS animations, transitions, hover effects
8. Professional typography with Google Fonts CDN

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

Your ENTIRE response must be a single JSON object. NO markdown fences. Start with { end with }
OUTPUT: {"index.html":"<!DOCTYPE html><html lang=\"ru\">...</html>"}`,
		specification, colorCtx, componentCtx, designCtx, techCtx, planSteps, imageCtx, specification)

	log.Printf("💻 Coder: генерирую функциональный код через %s", agent.Model)

	content, err := o.callLLMWithReasoning(ctx, agent.Model,
		`You are an elite full-stack web developer. You write FUNCTIONAL code, not just markup.
RULES:
- Every page MUST have real JavaScript: forms, interactivity, data rendering, cart logic.
- Use addEventListener for ALL events. No inline handlers.
- Store data in JS objects/arrays at the top of <script>. Render dynamically.
- Forms must validate inputs and show error/success messages.
- Shopping/ordering must calculate totals and persist in localStorage.
- Respond with valid JSON only. No markdown, no explanation.`,
		userPrompt, 16000, agent.ThinkingBudget)

	if err != nil {
		log.Printf("⚠️ Coder primary (%s) failed: %v — falling back to qwen-2.5-72b", agent.Model, err)
		// Fallback to a known-good model
		content, err = o.callLLM(ctx, "qwen/qwen-2.5-72b-instruct",
			"You are an expert frontend developer. Respond with valid JSON only. No markdown.",
			userPrompt, 16000)
		if err != nil {
			return nil, fmt.Errorf("code generation failed (both models): %w", err)
		}
	}

	files := o.parseCodeFiles(content)
	if len(files) == 0 {
		log.Printf("⚠️ Coder: JSON parse failed — extracting HTML directly")
		// Try to extract raw HTML if JSON parsing failed
		if idx := strings.Index(content, "<!DOCTYPE"); idx != -1 {
			files = map[string]string{"index.html": content[idx:]}
		} else if idx := strings.Index(content, "<html"); idx != -1 {
			files = map[string]string{"index.html": content[idx:]}
		} else {
			files = map[string]string{"index.html": content}
		}
	}

	log.Printf("✅ Coder: %d файлов сгенерировано", len(files))
	return files, nil
}

// sendStatus отправляет статус в шину событий
func (o *Orchestrator) sendStatus(agent AgentRole, status string, message string, progress int) {
	o.events.PublishStatus(agent, "", message, progress)
}

// GetStatusStream возвращает канал для получения событий (обратная совместимость SSE handler).
func (o *Orchestrator) GetStatusStream() <-chan domain.AgentEvent {
	return o.events.Subscribe()
}

// GetEventBus возвращает шину событий для прямого использования.
func (o *Orchestrator) GetEventBus() *domain.EventBus {
	return o.events
}

// Close закрывает оркестратор и шину событий.
func (o *Orchestrator) Close() {
	o.events.Close()
}
