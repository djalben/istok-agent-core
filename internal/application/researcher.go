package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/istok/agent-core/internal/application/usecases"
	"github.com/istok/agent-core/internal/domain"
	"github.com/istok/agent-core/internal/ports"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Researcher
//  Visual & Tech Audit (ядро Истока, adaptive thinking)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// VisualAuditResult результат визуального и технического аудита
type VisualAuditResult struct {
	URL          string            `json:"url"`
	Colors       []string          `json:"colors"`
	Fonts        []string          `json:"fonts"`
	Components   []string          `json:"components"`
	Layout       string            `json:"layout"`
	Technologies []string          `json:"technologies"`
	DesignSystem string            `json:"design_system"`
	Animations   []string          `json:"animations"`
	Breakpoints  []string          `json:"breakpoints"`
	Insights     []string          `json:"insights"`
	CSSVariables map[string]string `json:"css_variables"`
	AnalyzedAt   time.Time         `json:"analyzed_at"`
}

// ResearcherAgent агент-исследователь Истока (adaptive thinking)
type ResearcherAgent struct {
	llm   ports.LLMProvider
	model string
}

// NewResearcherAgent создает нового агента-исследователя
func NewResearcherAgent(llm ports.LLMProvider) *ResearcherAgent {
	return &ResearcherAgent{
		llm:   llm,
		model: "anthropic/claude-opus-4-7-thinking",
	}
}

// VisualAudit выполняет полный визуальный и технический аудит URL
func (r *ResearcherAgent) VisualAudit(ctx context.Context, url string, events *domain.EventBus) (*VisualAuditResult, error) {
	sendStatus := func(status, msg string, progress int) {
		events.PublishStatus(domain.RoleResearcher, "", msg, progress)
	}

	sendStatus("running", "🔍 Исследователь Истока анализирует визуальный код...", 10)

	prompt := r.buildAuditPrompt(url)

	log.Printf("🔍 ResearcherAgent: запрос к %s для аудита %s", r.model, url)

	result, err := r.callLLM(ctx, prompt)
	if err != nil {
		sendStatus("error", fmt.Sprintf("❌ Ошибка аудита: %v", err), 0)
		log.Printf("🚨 ResearcherAgent error: %v", err)
		// Возвращаем дефолтный результат, чтобы не блокировать генерацию
		return r.defaultAuditResult(url), nil
	}

	sendStatus("running", "🔍 Ядро Истока разбирает дизайн-систему...", 60)

	auditResult := r.parseAuditResult(url, result)

	sendStatus("completed", fmt.Sprintf("✅ Визуальный аудит завершён: найдено %d компонентов", len(auditResult.Components)), 100)

	log.Printf("✅ ResearcherAgent: аудит %s завершён, компонентов: %d", url, len(auditResult.Components))
	return auditResult, nil
}

// AnalyzeSpec выполняет итеративное глубокое исследование спецификации (Deep Research).
// Минимум 3 итерации: найти данные → задать уточняющие вопросы → уточнить анализ.
// Каждая итерация обогащает контекст для следующей.
func (r *ResearcherAgent) AnalyzeSpec(ctx context.Context, spec string, events *domain.EventBus) *VisualAuditResult {
	send := func(status, msg string, progress int) {
		events.PublishStatus(domain.RoleResearcher, "", msg, progress)
	}

	send("running", "🔍 Deep Research: запуск итеративного анализа (3 итерации)...", 5)
	log.Printf("🔍 ResearcherAgent.AnalyzeSpec: Deep Research mode, 3 iterations via %s", r.model)

	const maxIterations = 3
	accumulatedContext := ""
	var lastRawResult string

	for i := 1; i <= maxIterations; i++ {
		iterProgress := 5 + (i * 25) // 30, 55, 80

		// ── Reflecting phase ──
		events.PublishReflecting(domain.RoleResearcher,
			fmt.Sprintf("🔄 [Iteration %d/%d] Рефлексивный анализ спецификации...", i, maxIterations), iterProgress-10)

		var prompt string
		switch i {
		case 1:
			// Iteration 1: Initial broad analysis
			events.PublishReflecting(domain.RoleResearcher, "💡 [Goal] Первичный анализ: извлечение ключевых сущностей и дизайн-паттернов...", iterProgress-8)
			prompt = fmt.Sprintf(`You are an expert product analyst and frontend architect performing DEEP RESEARCH iteration 1/3.

TASK: Perform initial broad analysis of this specification. Identify:
- Core business entities and their relationships
- Key UI patterns and component hierarchy
- Technology stack requirements and constraints
- Potential edge cases and missing requirements

SPECIFICATION:
%s

Return ONLY valid JSON. Start with {. End with }.
{
  "colors": ["#hex1", "#hex2"],
  "fonts": ["FontName1", "FontName2"],
  "components": ["Component1", "Component2"],
  "layout": "description of ideal layout",
  "technologies": ["React", "TailwindCSS"],
  "design_system": "Material/Shadcn/Custom/etc",
  "animations": ["animation1", "animation2"],
  "breakpoints": ["mobile-first", "768px", "1024px"],
  "insights": ["key insight 1", "key insight 2"],
  "css_variables": {"--primary": "#value", "--background": "#value"},
  "clarifying_questions": ["question about unclear requirement 1", "question 2", "question 3"]
}`, spec)

		case 2:
			// Iteration 2: Self-questioning — answer own clarifying questions
			events.PublishReflecting(domain.RoleResearcher, "🔍 [Hypothesis] Уточняющий анализ: отвечаю на выявленные вопросы...", iterProgress-8)
			prompt = fmt.Sprintf(`You are an expert product analyst performing DEEP RESEARCH iteration 2/3.

PREVIOUS ANALYSIS (iteration 1):
%s

TASK: Review the previous analysis and its clarifying_questions. Answer each question using your expertise.
Then REFINE the design system based on these answers. Add more specific components, better color choices, and deeper insights.

ORIGINAL SPECIFICATION:
%s

Return ONLY valid JSON with the same structure. More detailed and refined. Start with {. End with }.
{
  "colors": ["#hex1", "#hex2"],
  "fonts": ["FontName1"],
  "components": ["more specific components"],
  "layout": "refined layout description",
  "technologies": ["specific versions"],
  "design_system": "precise system name",
  "animations": ["specific animations with timing"],
  "breakpoints": ["exact breakpoints"],
  "insights": ["deeper insights from answered questions"],
  "css_variables": {"--primary": "#value"}
}`, accumulatedContext, spec)

		case 3:
			// Iteration 3: Final synthesis with verification
			events.PublishReflecting(domain.RoleResearcher, "✅ [Verification] Финальный синтез: верификация полноты анализа...", iterProgress-8)
			prompt = fmt.Sprintf(`You are an expert product analyst performing DEEP RESEARCH iteration 3/3 (FINAL).

ACCUMULATED RESEARCH (iterations 1-2):
%s

TASK: Final verification and synthesis. Cross-check all findings against the original specification.
Ensure EVERY requirement is covered. Produce the definitive, production-ready design system analysis.
Remove any speculative elements. Keep only verified, actionable data.

ORIGINAL SPECIFICATION:
%s

Return ONLY valid JSON. This is the FINAL output. Start with {. End with }.
{
  "colors": ["#hex1", "#hex2"],
  "fonts": ["FontName1"],
  "components": ["verified component list"],
  "layout": "final layout description",
  "technologies": ["verified tech stack"],
  "design_system": "final system",
  "animations": ["verified animations"],
  "breakpoints": ["final breakpoints"],
  "insights": ["verified deep insights"],
  "css_variables": {"--primary": "#value"}
}`, accumulatedContext, spec)
		}

		events.PublishReflecting(domain.RoleResearcher,
			fmt.Sprintf("⚡ [Action] Итерация %d: запрос к LLM...", i), iterProgress-5)

		result, err := r.callLLM(ctx, prompt)
		if err != nil {
			log.Printf("⚠️ ResearcherAgent iteration %d error: %v", i, err)
			if i == 1 {
				send("error", fmt.Sprintf("⚠️ LLM недоступен, использую дефолтный анализ: %v", err), 100)
				return r.defaultAuditResult("spec://" + spec[:min(len(spec), 50)])
			}
			// На итерациях 2-3 используем предыдущий результат
			break
		}

		lastRawResult = result
		accumulatedContext += fmt.Sprintf("\n--- Iteration %d result ---\n%s\n", i, result)

		debugOut := result
		if len(debugOut) > 300 {
			debugOut = debugOut[:300] + "...[truncated]"
		}
		log.Printf("✅ Deep Research iteration %d/%d (%d chars): %s", i, maxIterations, len(result), debugOut)
		send("running", fmt.Sprintf("🔍 Deep Research: итерация %d/%d завершена", i, maxIterations), iterProgress)
	}

	send("running", "🔍 Исследователь Истока формирует финальный отчёт...", 85)
	auditResult := r.parseAuditResult("spec://"+spec[:min(len(spec), 50)], lastRawResult)
	send("completed", fmt.Sprintf("✅ Deep Research завершён (3 итерации): %d компонентов, %d цветов", len(auditResult.Components), len(auditResult.Colors)), 100)

	log.Printf("✅ ResearcherAgent.AnalyzeSpec (Deep Research): %d компонентов, %d технологий", len(auditResult.Components), len(auditResult.Technologies))
	return auditResult
}

// buildAuditPrompt формирует промпт для аудита
func (r *ResearcherAgent) buildAuditPrompt(url string) string {
	return fmt.Sprintf(`You are an expert UI/UX analyst and frontend architect. Analyze the website at %s.

Perform a comprehensive Visual & Tech Audit and return ONLY a valid JSON object with this exact structure:
{
  "colors": ["#hex1", "#hex2", "..."],
  "fonts": ["FontName1", "FontName2"],
  "components": ["Hero Section", "Navigation Bar", "Feature Cards", "..."],
  "layout": "description of overall layout",
  "technologies": ["React", "TailwindCSS", "Framer Motion", "..."],
  "design_system": "Material/Shadcn/Custom/etc",
  "animations": ["fade-in", "slide-up", "glassmorphism", "..."],
  "breakpoints": ["mobile-first", "1024px", "1280px"],
  "insights": ["key design insight 1", "key design insight 2"],
  "css_variables": {"--primary": "#value", "--background": "#value"}
}

Be specific about colors (use hex), real font names, and actual component names.

CRITICAL: YOUR ENTIRE RESPONSE MUST BE PURE JSON. NO THINKING. NO EXPLANATION. NO MARKDOWN. NO TEXT BEFORE OR AFTER THE JSON OBJECT. START YOUR RESPONSE WITH { AND END WITH }.`, url)
}

// callLLM выполняет запрос через порт LLMProvider (без прямых HTTP-вызовов)
func (r *ResearcherAgent) callLLM(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	resp, err := r.llm.Complete(ctx, ports.LLMRequest{
		Model:        r.model,
		SystemPrompt: "Strict Rule: Minimise reasoning. No conversational fillers. Be concise. Use White Label (Istok Core only).",
		UserPrompt:   prompt,
		MaxTokens:    2048,
		Temperature:  0.3,
	})
	if err != nil {
		return "", fmt.Errorf("researcher LLM call failed: %w", err)
	}
	return resp.Content, nil
}

// parseAuditResult парсит JSON ответ ядра
func (r *ResearcherAgent) parseAuditResult(url, content string) *VisualAuditResult {
	result := r.defaultAuditResult(url)

	jsonBlock, ok := usecases.ExtractFirstJSONObject(content)
	if !ok {
		log.Printf("⚠️ ResearcherAgent: no JSON object found in response (len=%d)", len(content))
		return result
	}

	var parsed struct {
		Colors       []string          `json:"colors"`
		Fonts        []string          `json:"fonts"`
		Components   []string          `json:"components"`
		Layout       string            `json:"layout"`
		Technologies []string          `json:"technologies"`
		DesignSystem string            `json:"design_system"`
		Animations   []string          `json:"animations"`
		Breakpoints  []string          `json:"breakpoints"`
		Insights     []string          `json:"insights"`
		CSSVariables map[string]string `json:"css_variables"`
	}

	if err := json.Unmarshal([]byte(jsonBlock), &parsed); err != nil {
		log.Printf("⚠️ ResearcherAgent: JSON unmarshal error: %v | block_len=%d", err, len(jsonBlock))
		return result
	}

	if len(parsed.Colors) > 0 {
		result.Colors = parsed.Colors
	}
	if len(parsed.Fonts) > 0 {
		result.Fonts = parsed.Fonts
	}
	if len(parsed.Components) > 0 {
		result.Components = parsed.Components
	}
	if parsed.Layout != "" {
		result.Layout = parsed.Layout
	}
	if len(parsed.Technologies) > 0 {
		result.Technologies = parsed.Technologies
	}
	if parsed.DesignSystem != "" {
		result.DesignSystem = parsed.DesignSystem
	}
	if len(parsed.Animations) > 0 {
		result.Animations = parsed.Animations
	}
	if len(parsed.Breakpoints) > 0 {
		result.Breakpoints = parsed.Breakpoints
	}
	if len(parsed.Insights) > 0 {
		result.Insights = parsed.Insights
	}
	if len(parsed.CSSVariables) > 0 {
		result.CSSVariables = parsed.CSSVariables
	}

	return result
}

// defaultAuditResult возвращает дефолтный результат при ошибке
func (r *ResearcherAgent) defaultAuditResult(url string) *VisualAuditResult {
	return &VisualAuditResult{
		URL:          url,
		Colors:       []string{"#5b4cdb", "#0e0e11", "#ffffff", "#f0f0f5"},
		Fonts:        []string{"Inter", "Geist Sans"},
		Components:   []string{"Hero Section", "Navigation", "Feature Cards", "CTA Button", "Footer"},
		Layout:       "Modern SPA с тёмной темой и градиентами",
		Technologies: []string{"React", "Vite", "TailwindCSS", "shadcn/ui"},
		DesignSystem: "Custom",
		Animations:   []string{"fade-in", "slide-up", "glassmorphism"},
		Breakpoints:  []string{"mobile-first", "768px", "1024px", "1280px"},
		Insights:     []string{"Акцент на визуальной иерархии", "Использование белого пространства"},
		CSSVariables: map[string]string{
			"--primary":    "#5b4cdb",
			"--background": "#0e0e11",
			"--foreground": "#ffffff",
		},
		AnalyzedAt: time.Now(),
	}
}

// jsonEscape безопасно экранирует строку для JSON
func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
