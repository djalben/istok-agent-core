package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — S-Tier AI Orchestrator
//  Мультимодельная архитектура нового поколения
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// GenerationMode режим генерации
type GenerationMode string

const (
	ModeAgent     GenerationMode = "agent"     // Инновационное проектирование: Claude Opus — глубокий анализ
	ModeCode      GenerationMode = "code"      // Быстрая генерация UI
	ModeSynthesis GenerationMode = "synthesis" // Адаптивный синтез конкурентов
)

// AgentRole определяет роль агента в системе
type AgentRole string

const (
	RoleDirector     AgentRole = "director"     // Claude Opus 4.6 - Логика и декомпозиция
	RoleBrain        AgentRole = "brain"        // Claude Opus 4.6 + Reasoning - Глубокий анализ
	RoleResearcher   AgentRole = "researcher"   // DeepSeek V3.2 - Адаптивный синтез конкурентов
	RoleCoder        AgentRole = "coder"        // Claude Opus 4.6 - Clean Code
	RoleDesigner     AgentRole = "designer"     // Gemini 3 Pro - UI ассеты
	RoleVideographer AgentRole = "videographer" // Gemini 3.1 Flash Lite - Промо-видео
	RoleValidator    AgentRole = "validator"    // Валидатор — синтаксическая и runtime проверка
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

// TaskStatus статус выполнения задачи
type TaskStatus struct {
	Agent     AgentRole
	Status    string
	Message   string
	Progress  int
	Timestamp time.Time
	Error     error
}

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

// MasterPlan план разработки от директора
type MasterPlan struct {
	Architecture string
	Components   []string
	Timeline     string
	Technologies []string
	Steps        []string
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

// Orchestrator управляет пулом AI агентов
type Orchestrator struct {
	agents       map[AgentRole]*AgentConfig
	statusStream chan TaskStatus
	mu           sync.RWMutex
	apiKey       string
}

// NewOrchestrator создает новый оркестратор
func NewOrchestrator() *Orchestrator {
	return NewOrchestratorWithKey("")
}

// NewOrchestratorWithKey создает оркестратор с API ключом
func NewOrchestratorWithKey(apiKey string) *Orchestrator {
	return &Orchestrator{
		agents: map[AgentRole]*AgentConfig{
			RoleDirector: {
				Role:            RoleDirector,
				Model:           "anthropic/claude-opus-4.6",
				Description:     "🧠 Директор — Claude Opus 4.6 Reasoning",
				Timeout:         5 * time.Minute,
				ThinkingEnabled: true,
				ThinkingBudget:  10000,
			},
			RoleBrain: {
				Role:            RoleBrain,
				Model:           "anthropic/claude-opus-4.6",
				Description:     "🧠 Мозг — Claude Opus 4.6 Extended Reasoning",
				Timeout:         10 * time.Minute,
				ThinkingEnabled: true,
				ThinkingBudget:  20000,
			},
			RoleResearcher: {
				Role:        RoleResearcher,
				Model:       "deepseek/deepseek-v3.2-speciale",
				Description: "🔍 Исследователь — DeepSeek V3.2 Адаптивный синтез",
				Timeout:     5 * time.Minute,
			},
			RoleCoder: {
				Role:            RoleCoder,
				Model:           "anthropic/claude-opus-4.6",
				Description:     "💻 Кодер — Claude Opus 4.6 Clean Code",
				Timeout:         10 * time.Minute,
				ThinkingEnabled: true,
				ThinkingBudget:  8000,
			},
			RoleDesigner: {
				Role:        RoleDesigner,
				Model:       "google/gemini-3-pro-image-preview",
				Description: "🎨 Дизайнер — Gemini 3 Pro UI-ассеты",
				Timeout:     5 * time.Minute,
			},
			RoleVideographer: {
				Role:        RoleVideographer,
				Model:       "google/gemini-3.1-flash-lite-preview",
				Description: "🎬 Видеограф — Gemini 3.1 Flash Lite промо-видео",
				Timeout:     15 * time.Minute,
			},
			RoleValidator: {
				Role:        RoleValidator,
				Model:       "anthropic/claude-opus-4.6",
				Description: "✅ Валидатор — Syntax & Runtime проверка",
				Timeout:     3 * time.Minute,
			},
		},
		statusStream: make(chan TaskStatus, 100),
		apiKey:       apiKey,
	}
}

// GenerateWithMode запускает процесс генерации в указанном режиме
func (o *Orchestrator) GenerateWithMode(ctx context.Context, specification string, url string, mode GenerationMode) (*GenerationResult, error) {
	if mode == ModeCode {
		return o.generateCodeMode(ctx, specification)
	}
	// Both "agent" (Инновационное проектирование) and "synthesis" (Адаптивный синтез) use full pipeline
	return o.generateAgentMode(ctx, specification, url)
}

// generateCodeMode быстрая генерация через DeepSeek-V3 (Code Mode)
func (o *Orchestrator) generateCodeMode(ctx context.Context, specification string) (*GenerationResult, error) {
	startTime := time.Now()
	result := &GenerationResult{
		Code:   make(map[string]string),
		Assets: make(map[string]string),
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	o.sendStatus(RoleCoder, "running", "⚡ Claude Opus 4.6 генерирует UI компоненты...", 20)

	plan := &MasterPlan{
		Architecture: "Quick UI Generation",
		Steps:        []string{specification},
	}

	code, err := o.generateCode(ctx, specification, plan, nil)
	if err != nil {
		o.sendStatus(RoleCoder, "error", fmt.Sprintf("❌ Ошибка: %v", err), 0)
		return nil, err
	}

	result.Code = code
	result.Duration = time.Since(startTime)
	o.sendStatus(RoleCoder, "completed", fmt.Sprintf("✅ Код готов за %v", result.Duration), 100)
	return result, nil
}

// generateAgentMode полная мультимодальная генерация с Claude Thinking (Agent Mode)
func (o *Orchestrator) generateAgentMode(ctx context.Context, specification string, url string) (*GenerationResult, error) {
	startTime := time.Now()
	result := &GenerationResult{
		Code:   make(map[string]string),
		Assets: make(map[string]string),
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Track synthesis features for architecture phase
	var competitorFeatures []CompetitorFeature

	// ── Этап 0 (ОБЯЗАТЕЛЬНЫЙ): DeepSeek V3.2 — Исследование + Глубокий синтез ВСЕГДА первым ──
	researcher := NewResearcherAgent(o.apiKey)
	if url != "" {
		// Глубокий синтез конкурента: извлечение всех фич + задачи для кодинга
		synthesis, _ := o.deepSynthesis(ctx, url, specification)
		if synthesis != nil && len(synthesis.Features) > 0 {
			competitorFeatures = synthesis.Features
		}

		// Визуальный аудит URL
		visualAudit, err := researcher.VisualAudit(ctx, url, o.statusStream)
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
		visualAudit := researcher.AnalyzeSpec(ctx, specification, o.statusStream)
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

	// ── Этап 1: Claude Opus 4.6 Brain — DefineArchitecture (Full-Stack манифест) ──
	manifest, archErr := o.defineArchitecture(ctx, specification, result.Audit, competitorFeatures)
	if archErr != nil {
		log.Printf("⚠️ Architecture manifest warning: %v", archErr)
	}

	// Этап 1b: Стратегический синтез
	o.sendStatus(RoleBrain, "running", "🧠 Claude Opus 4.6 анализирует стратегию...", 18)
	strategy, brainErr := o.synthesizeStrategy(ctx, specification, result.Audit)
	if brainErr != nil {
		log.Printf("⚠️ Brain synthesis warning (non-critical): %v", brainErr)
	} else if strategy != "" && result.Audit != nil {
		result.Audit.Audit = strategy
	}
	o.sendStatus(RoleBrain, "completed", "✅ Стратегия построена на основе анализа.", 22)

	// ── Этап 2: Director — Мастер-план ────────────────────────────────────────
	o.sendStatus(RoleDirector, "running", "🧠 Claude Opus 4.6 проектирует мастер-план...", 28)
	masterPlan, err := o.createMasterPlan(ctx, specification, result.Audit)
	if err != nil {
		o.sendStatus(RoleDirector, "error", fmt.Sprintf("❌ Ошибка планирования: %v", err), 0)
		return nil, fmt.Errorf("master plan creation failed: %w", err)
	}
	result.MasterPlan = masterPlan
	o.sendStatus(RoleDirector, "completed", "✅ Мастер-план спроектирован", 100)

	// ── Этап 3: Параллельный запуск Кодера + Дизайнера + Видеографа ──
	mediaService := newMediaService(o.apiKey)
	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Горутина 1: Coder пишет код с контекстом манифеста + синтеза + шаблонов
	wg.Add(1)
	go func() {
		defer wg.Done()
		o.sendStatus(RoleCoder, "running", "💻 Claude Opus 4.6 пишет производственный код...", 40)
		code, err := o.generateCodeFullStack(ctx, specification, masterPlan, result.Audit, manifest, competitorFeatures)
		if err != nil {
			errChan <- fmt.Errorf("code generation failed: %w", err)
			o.sendStatus(RoleCoder, "error", fmt.Sprintf("❌ Ошибка кода: %v", err), 0)
			return
		}
		code = o.validateAndHeal(ctx, code, specification)
		o.sendStatus(RoleValidator, "running", "🛡️ ValidatorAgent проверяет код...", 92)
		code = o.validatorCheck(ctx, code, specification)
		o.sendStatus(RoleValidator, "completed", "✅ Код прошёл валидацию", 100)
		o.mu.Lock()
		result.Code = code
		o.mu.Unlock()
		o.sendStatus(RoleCoder, "completed", fmt.Sprintf("✅ Код написан, проверен, валидирован (%d файлов)", len(code)), 100)
	}()

	// Горутина 2: Nano Banana 2 генерирует UI-ассеты
	wg.Add(1)
	go func() {
		defer wg.Done()
		o.sendStatus(RoleDesigner, "running", "🎨 Gemini 3 Pro рендерит UI-ассеты...", 55)
		var colors []string
		if result.VisualAudit != nil {
			colors = result.VisualAudit.Colors
		}
		assets, err := mediaService.GenerateUIAssets(ctx, "ИСТОК", specification, colors)
		if err != nil {
			errChan <- fmt.Errorf("asset generation failed: %w", err)
			o.sendStatus(RoleDesigner, "error", fmt.Sprintf("❌ Ошибка ассетов: %v", err), 0)
			return
		}
		o.mu.Lock()
		result.Assets = map[string]string{
			"logo.svg":      assets.LogoSVG,
			"hero_prompt":   assets.HeroPrompt,
			"og_prompt":     assets.OGImagePrompt,
			"color_palette": fmt.Sprintf("%v", assets.ColorPalette),
		}
		o.mu.Unlock()
		o.sendStatus(RoleDesigner, "completed", fmt.Sprintf("✅ UI-ассеты готовы: %d цветов, SVG логотип", len(assets.ColorPalette)), 100)
	}()

	// Горутина 3: Veo монтирует промо-ролик
	wg.Add(1)
	go func() {
		defer wg.Done()
		o.sendStatus(RoleVideographer, "running", "🎬 Gemini 3.1 Flash Lite монтирует промо-ролик...", 70)
		video, err := mediaService.GeneratePromoVideo(ctx, "ИСТОК", specification)
		if err != nil {
			errChan <- fmt.Errorf("video generation failed: %w", err)
			o.sendStatus(RoleVideographer, "error", fmt.Sprintf("❌ Ошибка видео: %v", err), 0)
			return
		}
		o.mu.Lock()
		result.Video = fmt.Sprintf("Script: %s | Scenes: %d | Music: %s", video.Script[:min(len(video.Script), 100)], len(video.Scenes), video.MusicStyle)
		o.mu.Unlock()
		o.sendStatus(RoleVideographer, "completed", fmt.Sprintf("✅ Промо-ролик готов: %d сцен, %s", len(video.Scenes), video.Duration), 100)
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	result.Duration = time.Since(startTime)
	o.sendStatus(RoleDirector, "completed", fmt.Sprintf("🎉 Мультимодальный проект готов за %v", result.Duration), 100)
	return result, nil
}

// min вспомогательная функция
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// createMasterPlan вызывает Claude (Director) для создания реального плана разработки
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

	userPrompt := fmt.Sprintf(`Create a concise technical master plan for this web project.

SPECIFICATION:
%s

DESIGN AUDIT (from Researcher Agent):
%s

Output ONLY a valid JSON object — no markdown, no explanation:
{
  "architecture": "concise architecture description tailored to the specification",
  "components": ["Component1", "Component2", "Component3"],
  "technologies": ["Technology1", "Technology2"],
  "timeline": "estimated timeline",
  "steps": ["Step 1: ...", "Step 2: ...", "Step 3: ..."]
}`, specification, auditSummary)

	log.Printf("🧠 Director: запрашиваю план у %s", agent.Model)

	result, err := o.callLLMWithReasoning(ctx, agent.Model,
		"You are a senior software architect. Create precise, actionable plans. Output only valid JSON.",
		userPrompt, 2048, agent.ThinkingBudget)

	if err != nil {
		log.Printf("⚠️ Director API error, using default plan: %v", err)
		return o.defaultMasterPlan(specification, audit), nil
	}

	plan := o.parseMasterPlan(result, specification, audit)
	log.Printf("✅ Director: план готов — %d шагов, %d технологий", len(plan.Steps), len(plan.Technologies))
	return plan, nil
}

// generateCodeFullStack вызывает Coder с полным контекстом: manifest + features + backend templates
func (o *Orchestrator) generateCodeFullStack(ctx context.Context, specification string, plan *MasterPlan, audit *ReverseEngineeringResult, manifest *SystemManifest, features []CompetitorFeature) (map[string]string, error) {
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

	return o.generateCode(ctx, enrichedSpec, plan, audit)
}

// generateCode вызывает Coder с полным контекстом от Researcher + Director
func (o *Orchestrator) generateCode(ctx context.Context, specification string, plan *MasterPlan, audit *ReverseEngineeringResult) (map[string]string, error) {
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

	userPrompt := fmt.Sprintf(`You are a world-class frontend developer. Build a STUNNING, production-ready web project.

PROJECT SPECIFICATION:
%s

DESIGN SYSTEM (from Researcher Agent):
- Color Palette: %s
- Key Components to include: %s
- Layout & Style: %s
- Technology hints: %s

IMPLEMENTATION STEPS (from Director Agent):
%s

REQUIREMENTS:
1. Output a JSON object mapping filename to file content (strings)
2. MUST include "index.html" — completely self-contained, ALL CSS and JS inline, renders in iframe immediately
3. Use TailwindCSS CDN (https://cdn.tailwindcss.com) for styling — it is reliable
4. Design must be VISUALLY STUNNING: modern gradients, smooth CSS animations, glassmorphism, professional typography
5. Use REAL content specific to "%s" — NO Lorem Ipsum, NO placeholder text, real sections and copy
6. Make it fully mobile-responsive
7. Include: hero section, features/benefits, call-to-action, footer — adapted to the project type
8. CRITICAL: Your ENTIRE response must be a single JSON object. NO markdown fences. Start with { end with }

OUTPUT FORMAT:
{"index.html":"<!DOCTYPE html><html lang=\"en\">...</html>"}`,
		specification, colorCtx, componentCtx, designCtx, techCtx, planSteps, specification)

	log.Printf("💻 Coder: генерирую код через %s", agent.Model)

	content, err := o.callLLMWithReasoning(ctx, agent.Model,
		"You are an expert frontend developer. Respond with valid JSON only. No markdown.",
		userPrompt, 32000, agent.ThinkingBudget)

	if err != nil {
		log.Printf("⚠️ Coder primary (%s) failed: %v — falling back to qwen-2.5-72b", agent.Model, err)
		// Fallback to a known-good model
		content, err = o.callLLM(ctx, "qwen/qwen-2.5-72b-instruct",
			"You are an expert frontend developer. Respond with valid JSON only. No markdown.",
			userPrompt, 8000)
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

// sendStatus отправляет статус в поток
func (o *Orchestrator) sendStatus(agent AgentRole, status string, message string, progress int) {
	select {
	case o.statusStream <- TaskStatus{
		Agent:     agent,
		Status:    status,
		Message:   message,
		Progress:  progress,
		Timestamp: time.Now(),
	}:
	default:
		// Канал заполнен, пропускаем
	}
}

// GetStatusStream возвращает канал для получения статусов
func (o *Orchestrator) GetStatusStream() <-chan TaskStatus {
	return o.statusStream
}

// Close закрывает оркестратор
func (o *Orchestrator) Close() {
	close(o.statusStream)
}
