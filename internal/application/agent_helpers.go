package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/istok/agent-core/internal/application/usecases"
	"github.com/istok/agent-core/internal/ports"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Agent Helpers — shared LLM helpers for Director + Coder
//  All calls go through ports.LLMProvider (no direct HTTP).
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// IstokStrictRule — обязательный префикс для всех системных промптов агентов Истока.
// Жесткая экономия токенов + White Label brand discipline.
const IstokStrictRule = "Strict Rule: Minimise reasoning. No conversational fillers. Be concise. Use White Label (Istok Core only).\n\n"

// withStrictRule добавляет Кодекс Истока в начало системного промпта.
func withStrictRule(systemPrompt string) string {
	if strings.HasPrefix(systemPrompt, "Strict Rule:") {
		return systemPrompt
	}
	return IstokStrictRule + systemPrompt
}

// callLLM sends a chat-completion request via the LLM port and returns the text response.
// Shared by Director (createMasterPlan) and Coder (generateCode).
// Effort defaults to "medium" for token economy.
func (o *Orchestrator) callLLM(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	resp, err := o.llm.Complete(ctx, ports.LLMRequest{
		Model:        model,
		SystemPrompt: withStrictRule(systemPrompt),
		UserPrompt:   userPrompt,
		MaxTokens:    maxTokens,
		Effort:       "medium",
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// callLLMWithReasoning sends a request with extended reasoning/thinking enabled.
// Adaptive thinking + effort control. thinkingBudget kept for port compatibility but
// adapter now uses adaptive mode (no explicit budget). Effort "high" for complex agents.
func (o *Orchestrator) callLLMWithReasoning(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens, thinkingBudget int) (string, error) {
	resp, err := o.llm.Complete(ctx, ports.LLMRequest{
		Model:        model,
		SystemPrompt: withStrictRule(systemPrompt),
		UserPrompt:   userPrompt,
		MaxTokens:    maxTokens,
		Reasoning:    true,
		Effort:       "high",
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// parseCodeFiles extracts a filename→content map from raw LLM output.
// Handles <thinking> blocks, markdown fences, broken JSON, and raw HTML extraction.
func (o *Orchestrator) parseCodeFiles(content string) map[string]string {
	original := content

	// Strip <thinking>...</thinking> blocks
	for strings.Contains(content, "<thinking>") {
		start := strings.Index(content, "<thinking>")
		end := strings.Index(content, "</thinking>")
		if end == -1 {
			break
		}
		content = content[:start] + content[end+len("</thinking>"):]
	}

	content = strings.TrimSpace(content)
	// Strip markdown fences
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// ── Strategy 1: Standard JSON parse ──
	first := strings.Index(content, "{")
	last := strings.LastIndex(content, "}")
	if first != -1 && last > first {
		jsonStr := content[first : last+1]
		var files map[string]string
		if err := json.Unmarshal([]byte(jsonStr), &files); err == nil && len(files) > 0 {
			log.Printf("✅ parseCodeFiles: strategy 1 (clean JSON) — %d files", len(files))
			return files
		}
	}

	// ── Strategy 2: Fix common JSON corruption then parse ──
	if first != -1 && last > first {
		fixed := content[first : last+1]
		// Replace literal control characters that break JSON
		fixed = strings.ReplaceAll(fixed, "\t", "\\t")
		// Fix truncated JSON: if it doesn't end with }, try to close it
		if !strings.HasSuffix(strings.TrimSpace(fixed), "}") {
			fixed = strings.TrimSpace(fixed) + "\"}"
		}
		var files map[string]string
		if err := json.Unmarshal([]byte(fixed), &files); err == nil && len(files) > 0 {
			log.Printf("✅ parseCodeFiles: strategy 2 (fixed JSON) — %d files", len(files))
			return files
		}
	}

	// ── Strategy 3: Extract "index.html" value manually ──
	// Find "index.html" key and extract the string value after it
	if idx := strings.Index(content, `"index.html"`); idx != -1 {
		rest := content[idx+len(`"index.html"`):] // skip key
		// Find the colon, then the opening quote
		colonIdx := strings.Index(rest, ":")
		if colonIdx != -1 {
			rest = rest[colonIdx+1:]
			rest = strings.TrimSpace(rest)
			if len(rest) > 0 && rest[0] == '"' {
				// Walk forward finding the matching unescaped closing quote
				html := extractJSONStringValue(rest)
				if len(html) > 50 {
					log.Printf("✅ parseCodeFiles: strategy 3 (manual extract) — %d chars", len(html))
					return map[string]string{"index.html": html}
				}
			}
		}
	}

	// ── Strategy 4: Raw HTML extraction ──
	src := original // use original (before thinking strip) as last resort
	if htmlIdx := strings.Index(src, "<!DOCTYPE"); htmlIdx != -1 {
		htmlEnd := strings.LastIndex(src, "</html>")
		if htmlEnd != -1 {
			html := src[htmlIdx : htmlEnd+len("</html>")]
			log.Printf("✅ parseCodeFiles: strategy 4 (raw HTML) — %d chars", len(html))
			return map[string]string{"index.html": html}
		}
		log.Printf("✅ parseCodeFiles: strategy 4 (raw HTML, no closing tag) — %d chars", len(src[htmlIdx:]))
		return map[string]string{"index.html": src[htmlIdx:]}
	}
	if htmlIdx := strings.Index(src, "<html"); htmlIdx != -1 {
		log.Printf("✅ parseCodeFiles: strategy 4 (raw <html>) — %d chars", len(src[htmlIdx:]))
		return map[string]string{"index.html": src[htmlIdx:]}
	}

	log.Printf("⚠️ parseCodeFiles: all strategies failed | len=%d | first100=%s", len(content), content[:min(100, len(content))])
	return nil
}

// extractJSONStringValue extracts the unescaped string value from a JSON string starting with ".
// It handles \" escape sequences and returns the decoded content.
func extractJSONStringValue(s string) string {
	if len(s) < 2 || s[0] != '"' {
		return ""
	}
	var b strings.Builder
	i := 1 // skip opening quote
	for i < len(s) {
		ch := s[i]
		if ch == '\\' && i+1 < len(s) {
			next := s[i+1]
			switch next {
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '/':
				b.WriteByte('/')
			default:
				b.WriteByte('\\')
				b.WriteByte(next)
			}
			i += 2
			continue
		}
		if ch == '"' {
			break // closing quote
		}
		b.WriteByte(ch)
		i++
	}
	return b.String()
}

// parseMasterPlan parses Director JSON output into a MasterPlan struct.
// Использует ExtractFirstJSONObject (bracket-counting) для устойчивости к длинным
// ответам Opus 4.7, где модель может добавлять prose ДО и ПОСЛЕ JSON-блока.
func (o *Orchestrator) parseMasterPlan(content, spec string, audit *ReverseEngineeringResult) *MasterPlan {
	origLen := len(content)
	jsonBlock, ok := usecases.ExtractFirstJSONObject(content)
	if !ok {
		head := content
		if len(head) > 500 {
			head = head[:500]
		}
		log.Printf("🚨 parseMasterPlan: NO JSON object found | total_len=%d | head=%q", origLen, head)
		return o.defaultMasterPlan(spec, audit)
	}

	var parsed struct {
		Architecture string    `json:"architecture"`
		Components   []string  `json:"components"`
		Technologies []string  `json:"technologies"`
		Timeline     string    `json:"timeline"`
		Steps        []string  `json:"steps"`
		DAG          []DAGTask `json:"dag"`
	}

	if err := json.Unmarshal([]byte(jsonBlock), &parsed); err != nil {
		head := jsonBlock
		if len(head) > 500 {
			head = head[:500]
		}
		log.Printf("🚨 parseMasterPlan: JSON unmarshal error: %v | total_len=%d block_len=%d | head=%q",
			err, origLen, len(jsonBlock), head)
		return o.defaultMasterPlan(spec, audit)
	}

	plan := &MasterPlan{
		Architecture: parsed.Architecture,
		Components:   parsed.Components,
		Technologies: parsed.Technologies,
		Timeline:     parsed.Timeline,
		Steps:        parsed.Steps,
		DAG:          parsed.DAG,
	}
	if plan.Architecture == "" {
		plan.Architecture = spec
	}
	// Если DAG пуст но steps есть — синтезируем DAG из steps для обратной совместимости
	if len(plan.DAG) == 0 && len(plan.Steps) > 0 {
		for i, step := range plan.Steps {
			var deps []string
			if i > 0 {
				deps = []string{fmt.Sprintf("T%d", i)}
			}
			plan.DAG = append(plan.DAG, DAGTask{
				ID:          fmt.Sprintf("T%d", i+1),
				Title:       step,
				Description: step,
				DependsOn:   deps,
			})
		}
	}

	// Hard floor: если всё ещё пусто — используем default 4-task DAG вместо одной
	// фиктивной задачи "= spec". Даёт осмысленный progress вместо пустого SSE.
	if len(plan.DAG) < 3 {
		log.Printf("⚠️ parseMasterPlan: degenerate plan (%d tasks), substituting default DAG", len(plan.DAG))
		default4 := o.defaultMasterPlan(spec, audit)
		plan.DAG = default4.DAG
		if len(plan.Steps) == 0 {
			plan.Steps = default4.Steps
		}
		if len(plan.Components) == 0 {
			plan.Components = default4.Components
		}
		if len(plan.Technologies) == 0 {
			plan.Technologies = default4.Technologies
		}
	}

	log.Printf("✅ parseMasterPlan: %d steps, %d DAG tasks", len(plan.Steps), len(plan.DAG))
	return plan
}

// synthesizeStrategy asks the Brain (ядро Истока) to produce a concise strategic brief
// from the Researcher audit data, enriching context for the Director.
func (o *Orchestrator) synthesizeStrategy(ctx context.Context, spec string, audit *ReverseEngineeringResult) (string, error) {
	if audit == nil {
		return "", nil
	}
	agent := o.agents[RoleBrain]
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	prompt := fmt.Sprintf(`Based on this research audit, write a 3-5 sentence strategic brief for the development team.
Focus on: key differentiators, UX priorities, visual identity, and must-have components.

SPECIFICATION: %s

RESEARCH AUDIT:
- Colors: %v
- Components: %v
- Layout: %s
- Technologies: %v
- Details: %s

Output ONLY the strategic brief text. No JSON, no markdown fences.`, spec, audit.Colors, audit.Components, audit.Layout, audit.Technologies, audit.Audit)

	result, err := o.callLLM(ctx, agent.Model,
		"You are a senior product strategist. Be concise and actionable. 3-5 sentences max.",
		prompt, 500)
	if err != nil {
		return "", err
	}
	log.Printf("✅ Brain: strategy synthesized (%d chars)", len(result))
	return strings.TrimSpace(result), nil
}

// defaultMasterPlan returns a sensible 4-task fallback plan when Director API fails
// or returns a degenerate response. Гарантирует осмысленный progress, а не пустой SSE.
func (o *Orchestrator) defaultMasterPlan(spec string, audit *ReverseEngineeringResult) *MasterPlan {
	plan := &MasterPlan{
		Architecture: spec,
		Components:   []string{"AppLayout", "Sidebar", "Header", "FeaturePages", "DataHooks"},
		Technologies: []string{"vite", "react", "typescript", "tailwindcss", "shadcn/ui", "@tanstack/react-query"},
		Timeline:     "immediate",
		Steps: []string{
			"Project scaffold",
			"UI shell (layout + navigation)",
			"Data layer (hooks + services)",
			"Feature pages",
		},
		DAG: []DAGTask{
			{ID: "T1", Title: "Project scaffold", Description: "Initialize Vite + React 18 + TypeScript with TailwindCSS, shadcn/ui and @/* aliases", DependsOn: nil, ImpactedFiles: []string{"package.json", "vite.config.ts", "tsconfig.json"}, RequiredDependencies: []string{"vite", "react", "react-dom", "tailwindcss"}},
			{ID: "T2", Title: "UI shell", Description: "Build AppLayout with Sidebar, Header and route container using shadcn primitives", DependsOn: []string{"T1"}, ImpactedFiles: []string{"src/components/layout/AppLayout.tsx", "src/components/layout/Sidebar.tsx", "src/components/layout/Header.tsx"}, RequiredDependencies: []string{"@radix-ui/react-slot", "lucide-react"}},
			{ID: "T3", Title: "Data layer", Description: "Create TanStack Query hooks and API service functions", DependsOn: []string{"T1"}, ImpactedFiles: []string{"src/hooks/useApi.ts", "src/services/api.ts"}, RequiredDependencies: []string{"@tanstack/react-query"}},
			{ID: "T4", Title: "Feature pages", Description: "Build all route pages consuming hooks from T3 and layout from T2", DependsOn: []string{"T2", "T3"}, ImpactedFiles: []string{"src/pages/Home.tsx"}, RequiredDependencies: nil},
		},
	}
	if audit != nil {
		if len(audit.Technologies) > 0 {
			plan.Technologies = audit.Technologies
		}
		if len(audit.Components) > 0 {
			plan.Components = audit.Components
		}
	}
	return plan
}
