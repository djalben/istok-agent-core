package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/istok/agent-core/internal/domain"
	"github.com/istok/agent-core/internal/ports"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Researcher
//  Gemini 2.0 Pro Visual & Tech Audit
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

// ResearcherAgent агент-исследователь на базе DeepSeek V3.2
type ResearcherAgent struct {
	llm   ports.LLMProvider
	model string
}

// NewResearcherAgent создает нового агента-исследователя
func NewResearcherAgent(llm ports.LLMProvider) *ResearcherAgent {
	return &ResearcherAgent{
		llm:   llm,
		model: "deepseek/deepseek-v3.2-speciale",
	}
}

// VisualAudit выполняет полный визуальный и технический аудит URL
func (r *ResearcherAgent) VisualAudit(ctx context.Context, url string, events *domain.EventBus) (*VisualAuditResult, error) {
	sendStatus := func(status, msg string, progress int) {
		events.PublishStatus(domain.RoleResearcher, "", msg, progress)
	}

	sendStatus("running", "🔍 Исследователь Gemini 2.0 анализирует визуальный код...", 10)

	prompt := r.buildAuditPrompt(url)

	log.Printf("🔍 ResearcherAgent: запрос к %s для аудита %s", r.model, url)

	result, err := r.callLLM(ctx, prompt)
	if err != nil {
		sendStatus("error", fmt.Sprintf("❌ Ошибка аудита: %v", err), 0)
		log.Printf("🚨 ResearcherAgent error: %v", err)
		// Возвращаем дефолтный результат, чтобы не блокировать генерацию
		return r.defaultAuditResult(url), nil
	}

	sendStatus("running", "🔍 Gemini 2.0 разбирает дизайн-систему...", 60)

	auditResult := r.parseAuditResult(url, result)

	sendStatus("completed", fmt.Sprintf("✅ Визуальный аудит завершён: найдено %d компонентов", len(auditResult.Components)), 100)

	log.Printf("✅ ResearcherAgent: аудит %s завершён, компонентов: %d", url, len(auditResult.Components))
	return auditResult, nil
}

// AnalyzeSpec анализирует текстовую спецификацию без URL (всегда запускается первым)
func (r *ResearcherAgent) AnalyzeSpec(ctx context.Context, spec string, events *domain.EventBus) *VisualAuditResult {
	send := func(status, msg string, progress int) {
		events.PublishStatus(domain.RoleResearcher, "", msg, progress)
	}

	send("running", "🔍 Gemini 2.0 Pro начал визуальное исследование...", 5)

	prompt := fmt.Sprintf(`You are an expert product analyst and frontend architect.
Analyze this project specification and return ONLY a valid JSON object describing the ideal design system.

CRITICAL: YOUR ENTIRE RESPONSE MUST BE PURE JSON. NO THINKING. NO EXPLANATION. NO MARKDOWN. NO TEXT BEFORE OR AFTER THE JSON OBJECT. START YOUR RESPONSE WITH { AND END WITH }.

SPECIFICATION:
%s

JSON STRUCTURE (output ONLY this, nothing else):
{
  "colors": ["#hex1", "#hex2", "..."],
  "fonts": ["FontName1", "FontName2"],
  "components": ["Component1", "Component2", "..."],
  "layout": "description of ideal layout",
  "technologies": ["React", "TailwindCSS", "..."],
  "design_system": "Material/Shadcn/Custom/etc",
  "animations": ["animation1", "animation2"],
  "breakpoints": ["mobile-first", "768px", "1024px"],
  "insights": ["key insight 1", "key insight 2"],
  "css_variables": {"--primary": "#value", "--background": "#value"}
}`, spec)

	log.Printf("🔍 ResearcherAgent.AnalyzeSpec: анализирую спецификацию через %s", r.model)

	result, err := r.callLLM(ctx, prompt)
	if err != nil {
		send("error", fmt.Sprintf("⚠️ LLM недоступен, использую дефолтный анализ: %v", err), 100)
		log.Printf("⚠️ ResearcherAgent.AnalyzeSpec error: %v", err)
		return r.defaultAuditResult("spec://" + spec[:min(len(spec), 50)])
	}

	send("running", "🔍 Gemini 2.0 формирует JSON-отчёт о дизайне...", 70)
	auditResult := r.parseAuditResult("spec://"+spec[:min(len(spec), 50)], result)
	send("completed", fmt.Sprintf("✅ Исследование завершено: %d компонентов, %d цветов", len(auditResult.Components), len(auditResult.Colors)), 100)

	log.Printf("✅ ResearcherAgent.AnalyzeSpec: %d компонентов, %d технологий", len(auditResult.Components), len(auditResult.Technologies))
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
		Model:       r.model,
		UserPrompt:  prompt,
		MaxTokens:   2048,
		Temperature: 0.3,
	})
	if err != nil {
		return "", fmt.Errorf("researcher LLM call failed: %w", err)
	}
	return resp.Content, nil
}

// parseAuditResult парсит JSON ответ от Gemini
func (r *ResearcherAgent) parseAuditResult(url, content string) *VisualAuditResult {
	result := r.defaultAuditResult(url)

	// Сначала убираем <thinking>...</thinking> блоки (Claude 3.7)
	if start := strings.Index(content, "<thinking>"); start != -1 {
		if end := strings.Index(content, "</thinking>"); end != -1 {
			content = content[:start] + content[end+len("</thinking>"):]
		}
	}

	// Очищаем markdown-обёртки если есть
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Извлекаем JSON между первым { и последним } (защита от любого текста вокруг)
	if first := strings.Index(content, "{"); first != -1 {
		if last := strings.LastIndex(content, "}"); last != -1 && last > first {
			content = content[first : last+1]
		}
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

	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		log.Printf("⚠️ ResearcherAgent: не удалось распарсить JSON: %v", err)
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
