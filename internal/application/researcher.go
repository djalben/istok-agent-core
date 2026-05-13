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

// AnalyzeSpec выполняет итеративное глубокое исследование спецификации (минимум 3 итерации).
// Паттерн: [Анализ] → [Уточняющие вопросы] → [Глубокое уточнение] → [Финальный отчёт]
func (r *ResearcherAgent) AnalyzeSpec(ctx context.Context, spec string, events *domain.EventBus) *VisualAuditResult {
	send := func(status, msg string, progress int) {
		events.PublishStatus(domain.RoleResearcher, "", msg, progress)
	}

	send("running", "🔍 Deep Research: итерация 1/3 — первичный анализ...", 5)
	events.PublishReflection(domain.RoleResearcher, "Starting iterative deep research (3 passes)")

	// ── Iteration 1: Initial Analysis ──
	iteration1Prompt := fmt.Sprintf(`You are an expert product analyst and frontend architect.
Analyze this project specification. Identify:
1. Core UI components needed
2. Color palette and typography
3. Technology stack implications
4. Potential architectural challenges

SPECIFICATION:
%s

Output a JSON object:
{
  "components": ["Component1", "Component2"],
  "colors": ["#hex1", "#hex2"],
  "fonts": ["Font1", "Font2"],
  "technologies": ["Tech1", "Tech2"],
  "challenges": ["challenge1", "challenge2"],
  "questions": ["What about X?", "Should Y use Z?"]
}

CRITICAL: PURE JSON ONLY. Start with {.`, spec)

	log.Printf("🔍 DeepResearch[1/3]: первичный анализ через %s", r.model)
	result1, err := r.callLLM(ctx, iteration1Prompt)
	if err != nil {
		send("error", fmt.Sprintf("⚠️ LLM недоступен: %v", err), 100)
		log.Printf("⚠️ DeepResearch[1/3] error: %v", err)
		return r.defaultAuditResult("spec://" + spec[:min(len(spec), 50)])
	}
	log.Printf("✅ DeepResearch[1/3]: %d chars", len(result1))

	// ── Iteration 2: Clarifying Questions + Deeper Analysis ──
	send("running", "🔍 Deep Research: итерация 2/3 — уточняющий анализ...", 35)
	events.PublishReflection(domain.RoleResearcher, fmt.Sprintf("Iteration 1 complete (%d chars). Starting clarification pass.", len(result1)))

	iteration2Prompt := fmt.Sprintf(`Based on your initial analysis of a project, now perform a DEEPER investigation.

ORIGINAL SPEC:
%s

YOUR INITIAL ANALYSIS:
%s

Now answer these self-generated questions and refine your analysis:
1. What specific shadcn/ui components map to each identified UI component?
2. What is the optimal color palette considering accessibility (WCAG AA contrast)?
3. What animations enhance UX without hurting performance?
4. What is the ideal responsive breakpoint strategy?
5. What CSS custom properties should be defined for the design system?

Output an ENHANCED JSON (same structure, but more detailed and refined):
{
  "colors": ["#hex1", "#hex2", "..."],
  "fonts": ["FontName1", "FontName2"],
  "components": ["specific shadcn-based Component1", "Component2", "..."],
  "layout": "detailed layout description with responsive strategy",
  "technologies": ["React", "TailwindCSS", "..."],
  "design_system": "shadcn/ui + specific customizations",
  "animations": ["specific animation with timing", "..."],
  "breakpoints": ["mobile-first", "768px", "1024px", "1280px"],
  "insights": ["deep insight 1", "deep insight 2", "..."],
  "css_variables": {"--primary": "#value", "--background": "#value", "..."}
}

CRITICAL: PURE JSON ONLY. Start with {.`, spec, result1)

	log.Printf("🔍 DeepResearch[2/3]: уточняющий анализ через %s", r.model)
	result2, err := r.callLLM(ctx, iteration2Prompt)
	if err != nil {
		log.Printf("⚠️ DeepResearch[2/3] error: %v — using iteration 1 result", err)
		result2 = result1
	} else {
		log.Printf("✅ DeepResearch[2/3]: %d chars", len(result2))
	}

	// ── Iteration 3: Final Synthesis + Verification ──
	send("running", "🔍 Deep Research: итерация 3/3 — финальный синтез...", 65)
	events.PublishReflection(domain.RoleResearcher, fmt.Sprintf("Iteration 2 complete (%d chars). Final synthesis pass.", len(result2)))

	iteration3Prompt := fmt.Sprintf(`You are performing a FINAL verification pass on a design system analysis.

ORIGINAL SPEC:
%s

REFINED ANALYSIS:
%s

VERIFICATION CHECKLIST:
1. Are all colors accessible (AA contrast ratio ≥ 4.5:1 for text)?
2. Do components cover ALL user flows in the spec?
3. Is the tech stack internally consistent (no conflicting packages)?
4. Are breakpoints sufficient for all target devices?
5. Do CSS variables form a complete, coherent design token system?

Fix any issues found. Output the FINAL, production-ready design system JSON:
{
  "colors": ["#hex1", "#hex2", "..."],
  "fonts": ["FontName1", "FontName2"],
  "components": ["Component1", "Component2", "..."],
  "layout": "final layout description",
  "technologies": ["React", "TailwindCSS", "..."],
  "design_system": "final design system name",
  "animations": ["animation1", "animation2"],
  "breakpoints": ["mobile-first", "768px", "1024px", "1280px"],
  "insights": ["verified insight 1", "verified insight 2"],
  "css_variables": {"--primary": "#value", "--background": "#value"}
}

CRITICAL: PURE JSON ONLY. Start with {.`, spec[:min(len(spec), 1000)], result2)

	log.Printf("🔍 DeepResearch[3/3]: финальная верификация через %s", r.model)
	result3, err := r.callLLM(ctx, iteration3Prompt)
	if err != nil {
		log.Printf("⚠️ DeepResearch[3/3] error: %v — using iteration 2 result", err)
		result3 = result2
	} else {
		log.Printf("✅ DeepResearch[3/3]: %d chars", len(result3))
	}

	// Parse final result
	send("running", "🔍 Deep Research: формирование финального отчёта...", 90)
	auditResult := r.parseAuditResult("spec://"+spec[:min(len(spec), 50)], result3)
	send("completed", fmt.Sprintf("✅ Deep Research (3 итерации): %d компонентов, %d цветов", len(auditResult.Components), len(auditResult.Colors)), 100)

	log.Printf("✅ DeepResearch COMPLETE: %d компонентов, %d технологий (3 iterations)", len(auditResult.Components), len(auditResult.Technologies))
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
