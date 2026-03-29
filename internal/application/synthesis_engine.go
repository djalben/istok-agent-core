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
//  ИСТОК АГЕНТ — Synthesis Engine
//  Глубокий анализ конкурентов → Задачи для кодинга
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// CompetitorFeature — фича конкурента, извлечённая из анализа
type CompetitorFeature struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`    // "payments", "auth", "dashboard", "notifications"
	Priority    string   `json:"priority"`    // "critical", "high", "medium", "low"
	Complexity  string   `json:"complexity"`  // "simple", "moderate", "complex"
	UXPatterns  []string `json:"ux_patterns"` // ["modal-confirmation", "swipe-to-pay", "pull-to-refresh"]
	Endpoints   []string `json:"endpoints"`   // ["/api/transfers", "/api/cards/limits"]
	UIElements  []string `json:"ui_elements"` // ["TransferForm", "CardLimitSlider"]
}

// SynthesisResult — полный результат глубокого синтеза
type SynthesisResult struct {
	CompetitorURL  string              `json:"competitor_url"`
	CompetitorName string              `json:"competitor_name"`
	Features       []CompetitorFeature `json:"features"`
	TechStack      []string            `json:"tech_stack"`
	DesignPatterns []string            `json:"design_patterns"`
	CodingTasks    []CodingTask        `json:"coding_tasks"`
	AnalyzedAt     time.Time           `json:"analyzed_at"`
}

// CodingTask — задача для кодинга, сгенерированная из фичи конкурента
type CodingTask struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Feature     string   `json:"feature"`     // ссылка на CompetitorFeature.Name
	Type        string   `json:"type"`        // "frontend", "backend", "database", "api"
	Files       []string `json:"files"`       // файлы которые нужно создать
	Priority    int      `json:"priority"`    // 1-5
	EstMinutes  int      `json:"est_minutes"` // оценка времени
}

// deepSynthesis выполняет глубокий анализ конкурента и генерирует задачи для кодинга
func (o *Orchestrator) deepSynthesis(ctx context.Context, url, spec string) (*SynthesisResult, error) {
	agent := o.agents[RoleResearcher]
	ctx, cancel := context.WithTimeout(ctx, agent.Timeout)
	defer cancel()

	o.sendStatus(RoleResearcher, "running", "🔍 Глубокий синтез: анализ функций конкурента...", 5)

	prompt := fmt.Sprintf(`You are an expert competitive analyst and product engineer.
Deeply analyze the product/service at URL: %s

The user wants to build: %s

Your task:
1. Enumerate ALL features of the competitor product (even hidden/non-obvious ones)
2. Classify each feature by category, priority, complexity
3. Identify UX patterns used for each feature
4. Map each feature to API endpoints and UI components needed
5. Convert everything into actionable coding tasks

CRITICAL: Output ONLY valid JSON. No markdown, no explanation. Start with {.

{
  "competitor_url": "%s",
  "competitor_name": "CompetitorName",
  "features": [
    {
      "name": "P2P Transfers",
      "description": "Send money to other users by phone number or username",
      "category": "payments",
      "priority": "critical",
      "complexity": "complex",
      "ux_patterns": ["modal-confirmation", "contact-picker", "amount-keypad"],
      "endpoints": ["/api/transfers/send", "/api/transfers/history", "/api/contacts/search"],
      "ui_elements": ["TransferForm", "ContactPicker", "AmountInput", "ConfirmModal", "TransferHistory"]
    },
    {
      "name": "Card Limits Management",
      "description": "Set daily/monthly spending limits, ATM withdrawal limits",
      "category": "cards",
      "priority": "high",
      "complexity": "moderate",
      "ux_patterns": ["slider-input", "toggle-switch", "instant-apply"],
      "endpoints": ["/api/cards/{id}/limits", "/api/cards/{id}/settings"],
      "ui_elements": ["LimitSlider", "CardSettings", "LimitConfirmation"]
    }
  ],
  "tech_stack": ["React", "Node.js", "PostgreSQL", "Redis", "Stripe API"],
  "design_patterns": ["dark-mode-first", "bottom-navigation", "card-based-layout", "skeleton-loading"],
  "coding_tasks": [
    {
      "id": "task-1",
      "title": "Implement P2P Transfer Flow",
      "description": "Build complete money transfer flow with contact search, amount input, and confirmation",
      "feature": "P2P Transfers",
      "type": "fullstack",
      "files": ["frontend/TransferPage.tsx", "frontend/ContactPicker.tsx", "backend/handlers/transfers.go", "backend/db/transfers.sql"],
      "priority": 1,
      "est_minutes": 120
    }
  ]
}

Be EXHAUSTIVE. List 10-30 features. Generate 15-40 coding tasks. Think like a PM doing competitive analysis for a startup.`, url, spec, url)

	log.Printf("🔍 SynthesisEngine: deep analysis of %s via %s", url, agent.Model)

	result, err := o.callLLM(ctx, agent.Model,
		"You are an expert competitive analyst. Enumerate ALL features of the target product. Be exhaustive. Output pure JSON only.",
		prompt, 16384)

	if err != nil {
		log.Printf("⚠️ SynthesisEngine: LLM error: %v", err)
		o.sendStatus(RoleResearcher, "error", fmt.Sprintf("⚠️ Ошибка синтеза: %v", err), 0)
		return o.defaultSynthesisResult(url, spec), nil
	}

	synthesis := o.parseSynthesisResult(result, url)

	o.sendStatus(RoleResearcher, "completed",
		fmt.Sprintf("✅ Глубокий синтез: %d фич, %d задач для кодинга",
			len(synthesis.Features), len(synthesis.CodingTasks)), 100)

	log.Printf("✅ SynthesisEngine: %d features, %d tasks from %s",
		len(synthesis.Features), len(synthesis.CodingTasks), url)

	return synthesis, nil
}

// parseSynthesisResult парсит JSON от DeepSeek
func (o *Orchestrator) parseSynthesisResult(content, url string) *SynthesisResult {
	// Strip thinking blocks
	for strings.Contains(content, "<thinking>") {
		start := strings.Index(content, "<thinking>")
		end := strings.Index(content, "</thinking>")
		if end == -1 {
			break
		}
		content = content[:start] + content[end+len("</thinking>"):]
	}

	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	if first := strings.Index(content, "{"); first != -1 {
		if last := strings.LastIndex(content, "}"); last > first {
			content = content[first : last+1]
		}
	}

	var result SynthesisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		log.Printf("⚠️ parseSynthesisResult JSON error: %v", err)
		return o.defaultSynthesisResult(url, "")
	}

	result.AnalyzedAt = time.Now()
	return &result
}

// defaultSynthesisResult возвращает базовый результат при ошибке
func (o *Orchestrator) defaultSynthesisResult(url, spec string) *SynthesisResult {
	return &SynthesisResult{
		CompetitorURL:  url,
		CompetitorName: "Unknown",
		Features: []CompetitorFeature{
			{Name: "User Authentication", Description: "Login/Register/OAuth", Category: "auth", Priority: "critical", Complexity: "moderate"},
			{Name: "Dashboard", Description: "Main user dashboard with key metrics", Category: "dashboard", Priority: "critical", Complexity: "moderate"},
			{Name: "Settings", Description: "User profile and app settings", Category: "settings", Priority: "high", Complexity: "simple"},
		},
		TechStack:      []string{"React", "Go", "PostgreSQL", "TailwindCSS"},
		DesignPatterns: []string{"dark-mode", "card-layout", "sidebar-navigation"},
		CodingTasks: []CodingTask{
			{ID: "task-1", Title: "Auth System", Description: "JWT auth with login/register", Feature: "User Authentication", Type: "fullstack", Priority: 1},
			{ID: "task-2", Title: "Dashboard UI", Description: "Main dashboard with metrics cards", Feature: "Dashboard", Type: "frontend", Priority: 1},
			{ID: "task-3", Title: "Settings Page", Description: "User settings and profile", Feature: "Settings", Type: "frontend", Priority: 2},
		},
		AnalyzedAt: time.Now(),
	}
}

// featuresToContext превращает результат синтеза в текстовый контекст для Coder
func featuresToContext(synthesis *SynthesisResult) string {
	if synthesis == nil || len(synthesis.Features) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\nCOMPETITOR ANALYSIS (Deep Synthesis):\n")
	sb.WriteString(fmt.Sprintf("Competitor: %s (%s)\n", synthesis.CompetitorName, synthesis.CompetitorURL))
	sb.WriteString(fmt.Sprintf("Tech Stack: %s\n", strings.Join(synthesis.TechStack, ", ")))
	sb.WriteString(fmt.Sprintf("Design Patterns: %s\n\n", strings.Join(synthesis.DesignPatterns, ", ")))

	sb.WriteString("FEATURES TO IMPLEMENT:\n")
	for i, f := range synthesis.Features {
		sb.WriteString(fmt.Sprintf("%d. [%s/%s] %s — %s\n", i+1, f.Priority, f.Complexity, f.Name, f.Description))
		if len(f.UXPatterns) > 0 {
			sb.WriteString(fmt.Sprintf("   UX: %s\n", strings.Join(f.UXPatterns, ", ")))
		}
		if len(f.Endpoints) > 0 {
			sb.WriteString(fmt.Sprintf("   API: %s\n", strings.Join(f.Endpoints, ", ")))
		}
	}

	if len(synthesis.CodingTasks) > 0 {
		sb.WriteString("\nCODING TASKS (ordered by priority):\n")
		for _, t := range synthesis.CodingTasks {
			sb.WriteString(fmt.Sprintf("- [P%d] %s: %s\n", t.Priority, t.Title, t.Description))
			if len(t.Files) > 0 {
				sb.WriteString(fmt.Sprintf("  Files: %s\n", strings.Join(t.Files, ", ")))
			}
		}
	}

	return sb.String()
}
