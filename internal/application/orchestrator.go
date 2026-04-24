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
	ModeAgent     GenerationMode = "agent"     // Инновационное проектирование: Gemini 3 Pro — глубокий анализ
	ModeCode      GenerationMode = "code"      // Быстрая генерация UI
	ModeSynthesis GenerationMode = "synthesis" // Адаптивный синтез конкурентов
)

// AgentRole определяет роль агента в системе
type AgentRole string

const (
	RoleDirector     AgentRole = "director"     // Gemini 3 Pro - Логика и декомпозиция
	RoleBrain        AgentRole = "brain"        // Gemini 3 Pro - Глубокий анализ
	RoleResearcher   AgentRole = "researcher"   // DeepSeek V3.2 - Адаптивный синтез конкурентов
	RoleCoder        AgentRole = "coder"        // Gemini 3 Pro - Clean Code
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
				Role:        RoleDirector,
				Model:       "google/gemini-3-pro",
				Description: "🧠 Директор — Gemini 3 Pro Reasoning",
				Timeout:     5 * time.Minute,
			},
			RoleBrain: {
				Role:        RoleBrain,
				Model:       "google/gemini-3-pro",
				Description: "🧠 Мозг — Gemini 3 Pro Deep Analysis",
				Timeout:     10 * time.Minute,
			},
			RoleResearcher: {
				Role:        RoleResearcher,
				Model:       "deepseek/deepseek-v3.2-speciale",
				Description: "🔍 Исследователь — DeepSeek V3.2 Адаптивный синтез",
				Timeout:     5 * time.Minute,
			},
			RoleCoder: {
				Role:        RoleCoder,
				Model:       "google/gemini-3-pro",
				Description: "💻 Кодер — Gemini 3 Pro Clean Code",
				Timeout:     10 * time.Minute,
			},
			RoleDesigner: {
				Role:        RoleDesigner,
				Model:       "google/gemini-3-pro",
				Description: "🎨 Дизайнер — Gemini 3 Pro UI-ассеты",
				Timeout:     5 * time.Minute,
			},
			RoleVideographer: {
				Role:        RoleVideographer,
				Model:       "google/gemini-3.1-pro",
				Description: "🎬 Видеограф — Gemini 3.1 Pro промо-видео",
				Timeout:     15 * time.Minute,
			},
			RoleValidator: {
				Role:        RoleValidator,
				Model:       "google/gemini-3-pro",
				Description: "✅ Валидатор — Gemini 3 Pro Syntax & Runtime",
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

	o.sendStatus(RoleCoder, "running", "⚡ Gemini 3 Pro генерирует UI компоненты...", 20)

	plan := &MasterPlan{
		Architecture: "Quick UI Generation",
		Steps:        []string{specification},
	}

	code, err := o.generateCode(ctx, specification, plan, nil, nil)
	if err != nil {
		o.sendStatus(RoleCoder, "error", fmt.Sprintf("❌ Ошибка: %v", err), 0)
		return nil, err
	}

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

	// ── Этап 2: Director — Мастер-план ────────────────────────────────────────
	o.sendStatus(RoleDirector, "running", "🧠 Gemini 3 Pro проектирует мастер-план...", 28)
	masterPlan, err := o.createMasterPlan(ctx, specification, result.Audit)
	if err != nil {
		o.sendStatus(RoleDirector, "error", fmt.Sprintf("❌ Ошибка планирования: %v", err), 0)
		return nil, fmt.Errorf("master plan creation failed: %w", err)
	}
	result.MasterPlan = masterPlan
	o.sendStatus(RoleDirector, "completed", "✅ Мастер-план спроектирован", 100)

	// ── Этап 3: Дизайнер генерирует изображения ПЕРВЫМ (Nano Banana 2) ──
	// Дизайнер запускается ДО Кодера, чтобы передать ему реальные URL изображений
	mediaService := newMediaService(o.apiKey)
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

	// ── Этап 4: Кодер + Видеограф параллельно ──
	// Кодер получает URL изображений от Дизайнера и встраивает их в код
	var wg sync.WaitGroup
	var coderErr error

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
		code = o.validateAndHeal(ctx, code, specification)
		o.sendStatus(RoleValidator, "running", "🛡️ ValidatorAgent проверяет код...", 92)
		code = o.validatorCheck(ctx, code, specification)
		o.sendStatus(RoleValidator, "completed", "✅ Код прошёл валидацию", 100)
		o.mu.Lock()
		result.Code = code
		o.mu.Unlock()
		o.sendStatus(RoleCoder, "completed", fmt.Sprintf("✅ Функциональный код готов (%d файлов)", len(code)), 100)
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
		return nil, coderErr
	}

	result.Duration = time.Since(startTime)
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

// createMasterPlan вызывает Gemini (Director) для создания реального плана разработки
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

	userPrompt := fmt.Sprintf(`Create a FUNCTIONAL implementation plan for this web project. Focus on WHAT the code must DO, not just what it looks like.

SPECIFICATION:
%s

DESIGN AUDIT (from Researcher Agent):
%s

Output ONLY a valid JSON object — no markdown, no explanation:
{
  "architecture": "architecture description with key data structures and business logic",
  "components": ["Component1 (with interaction description)", "Component2", ...],
  "technologies": ["TailwindCSS CDN", "Vanilla JS", "localStorage", ...],
  "timeline": "estimated timeline",
  "steps": [
    "Step 1: Define JS data structures (menu items with name/price/category/image, cart state)",
    "Step 2: Build responsive navigation with hamburger menu toggle (JS)",
    "Step 3: Render hero section with real images from Designer",
    "Step 4: Build interactive menu/product grid rendered from JS data arrays",
    "Step 5: Implement cart system with add/remove/quantity and localStorage persistence",
    "Step 6: Create order/contact form with field validation and success feedback",
    "Step 7: Add smooth scroll, animations, toast notifications"
  ]
}

CRITICAL: Each step must describe FUNCTIONAL behavior, not just visual layout.
Bad: "Create hero section" Good: "Create hero with CTA button that smooth-scrolls to menu section"
Bad: "Add menu" Good: "Render menu grid from JS menuItems array with category filter tabs and Add to Cart buttons"`, specification, auditSummary)

	log.Printf("🧠 Director: запрашиваю план у %s", agent.Model)

	result, err := o.callLLMWithReasoning(ctx, agent.Model,
		`You are a senior software architect. Create precise, actionable plans. Output only valid JSON.
ARCHITECTURE RULES:
- Never put business logic in main.go or HTTP handlers.
- Separate Domain (entities), Application (use cases), Infrastructure (external APIs), Transport (HTTP/SSE).
- All external dependencies must go through interfaces (ports).`,
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
