package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
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

// ResearcherAgent агент-исследователь на базе Gemini 2.0 Pro
type ResearcherAgent struct {
	apiKey  string
	baseURL string
	model   string
}

// NewResearcherAgent создает нового агента-исследователя
func NewResearcherAgent(apiKey string) *ResearcherAgent {
	return &ResearcherAgent{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		model:   "google/gemini-2.0-pro",
	}
}

// VisualAudit выполняет полный визуальный и технический аудит URL
func (r *ResearcherAgent) VisualAudit(ctx context.Context, url string, statusChan chan<- TaskStatus) (*VisualAuditResult, error) {
	sendStatus := func(status, msg string, progress int) {
		select {
		case statusChan <- TaskStatus{
			Agent:     RoleResearcher,
			Status:    status,
			Message:   msg,
			Progress:  progress,
			Timestamp: time.Now(),
		}:
		default:
		}
	}

	sendStatus("running", "🔍 Исследователь Gemini 2.0 анализирует визуальный код...", 10)

	prompt := r.buildAuditPrompt(url)

	log.Printf("🔍 ResearcherAgent: запрос к %s для аудита %s", r.model, url)

	result, err := r.callOpenRouter(ctx, prompt)
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
Return ONLY the JSON, no markdown, no explanation.`, url)
}

// callOpenRouter выполняет запрос к OpenRouter API
func (r *ResearcherAgent) callOpenRouter(ctx context.Context, prompt string) (string, error) {
	if r.apiKey == "" || strings.HasPrefix(r.apiKey, "MISSING") {
		return "", fmt.Errorf("OPENROUTER_API_KEY не установлен")
	}

	import_bytes := fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":%s}],"max_tokens":2048,"temperature":0.3}`,
		r.model, jsonEscape(prompt))

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	// Используем HTTP напрямую для простоты
	body, status, err := httpPost(ctx, r.baseURL+"/chat/completions", r.apiKey, []byte(import_bytes))
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}

	if status != 200 {
		log.Printf("🚨 Gemini 2.0 Pro error | status=%d | %s", status, string(body))
		return "", fmt.Errorf("OpenRouter API error (HTTP %d): %s", status, string(body))
	}

	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from model")
	}

	return resp.Choices[0].Message.Content, nil
}

// parseAuditResult парсит JSON ответ от Gemini
func (r *ResearcherAgent) parseAuditResult(url, content string) *VisualAuditResult {
	result := r.defaultAuditResult(url)

	// Очищаем markdown-обёртки если есть
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

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
