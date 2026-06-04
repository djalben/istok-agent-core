package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper"
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

// llmCallTimeout — hard per-call timeout for any single LLM request.
// Prevents infinite hangs when LLM is unresponsive or stuck.
// Heavy Sonnet/Opus thinking generations (architect, coder) can exceed 4min,
// so we allow 6min (still below the 8min HTTP client ceiling).
const llmCallTimeout = 6 * time.Minute

// callLLM sends a chat-completion request via the LLM port and returns the text response.
// Shared by Director (createMasterPlan) and Coder (generateCode).
// Effort defaults to "medium" for token economy.
// If the provider returns ErrInsufficientFunds, the call pauses (SSE event + WaitForFunds)
// and retries once after the user resumes.
func (o *Orchestrator) callLLM(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	for range 2 {
		callCtx, cancel := context.WithTimeout(ctx, llmCallTimeout)
		resp, err := o.llm.Complete(callCtx, ports.LLMRequest{
			Model:        model,
			SystemPrompt: withStrictRule(systemPrompt),
			UserPrompt:   userPrompt,
			MaxTokens:    maxTokens,
		})
		cancel()

		if err == nil {
			return resp.Content, nil
		}

		if errors.Is(err, ports.ErrInsufficientFunds) {
			waitErr := o.pauseForFunds(ctx)
			if waitErr != nil {
				return "", waitErr
			}

			continue // retry after resume
		}

		if callCtx.Err() != nil {
			applog(ctx).ErrorContext(ctx, "LLM call timed out", "timeout", llmCallTimeout, "model", model)
		}

		return "", wrapper.Wrap(err)
	}

	return "", ports.ErrInsufficientFunds
}

// callLLMWithReasoning sends a request with extended reasoning/thinking enabled.
// Adaptive Thinking API — effort "high" for complex agents. No budget_tokens needed.
// Pauses on ErrInsufficientFunds (same as callLLM).
func (o *Orchestrator) callLLMWithReasoning(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	for range 2 {
		callCtx, cancel := context.WithTimeout(ctx, llmCallTimeout)
		resp, err := o.llm.Complete(callCtx, ports.LLMRequest{
			Model:        model,
			SystemPrompt: withStrictRule(systemPrompt),
			UserPrompt:   userPrompt,
			MaxTokens:    maxTokens,
			Reasoning:    true,
		})
		cancel()

		if err == nil {
			return resp.Content, nil
		}

		if errors.Is(err, ports.ErrInsufficientFunds) {
			waitErr := o.pauseForFunds(ctx)
			if waitErr != nil {
				return "", waitErr
			}

			continue
		}

		if callCtx.Err() != nil {
			applog(ctx).ErrorContext(ctx, "LLM reasoning call timed out", "timeout", llmCallTimeout, "model", model)
		}

		return "", wrapper.Wrap(err)
	}

	return "", ports.ErrInsufficientFunds
}

// pauseForFunds publishes an SSE event and blocks until the user resumes (tops up balance).
func (o *Orchestrator) pauseForFunds(ctx context.Context) error {
	sessionID, _ := ctx.Value(sessionIDKey{}).(string)
	if sessionID == "" {
		return ports.ErrInsufficientFunds // can't pause without session
	}
	applog(ctx).WarnContext(ctx, "insufficient funds, pausing session", "session_id", sessionID)

	o.fundsRegistry.Register(sessionID)
	o.events.Publish(domain.AgentEvent{
		Kind:      domain.EventInsufficientFunds,
		Agent:     "system",
		Message:   "Недостаточно средств для продолжения генерации. Пополните кошелек.",
		SessionID: sessionID,
		Timestamp: time.Now(),
	})

	return o.fundsRegistry.WaitForFunds(ctx, sessionID)
}

// xmlFileRegex matches <file path="...">...</file> blocks produced by LLM in the XML artifact protocol.
// (?s) enables dot-matches-newline so multi-line code is captured correctly.
var xmlFileRegex = regexp.MustCompile(`(?s)<file\s+path="([^"]+)"\s*>\s*(.*?)\s*</file>`)

// parseCodeFiles extracts a filename→content map from raw LLM output.
// Strategy priority:
//  1. XML artifact protocol (<file path="...">...</file>) — primary, most resilient
//  2. JSON parse (legacy fallback for older prompts / single-file mode)
//  3. Truncated JSON recovery
//  4. Raw HTML extraction
func (o *Orchestrator) parseCodeFiles(ctx context.Context, content string) map[string]string {
	original := content
	content = stripThinkingBlocks(content)

	if files := parseCodeFilesFromXML(ctx, content); len(files) > 0 {
		return files
	}
	if files := parseCodeFilesFromJSON(ctx, content); len(files) > 0 {
		return files
	}
	if files := extractIndexHTMLFromJSON(content); len(files) > 0 {
		return files
	}
	if files := extractRawHTMLFiles(original); len(files) > 0 {
		return files
	}
	applog(ctx).WarnContext(ctx, "parseCodeFiles: all strategies failed", "len", len(content), "head", content[:min(100, len(content))])

	return nil
}

func stripThinkingBlocks(content string) string {
	for strings.Contains(content, "<thinking>") {
		start := strings.Index(content, "<thinking>")
		end := strings.Index(content, "</thinking>")
		if end == -1 {
			break
		}
		content = content[:start] + content[end+len("</thinking>"):]
	}

	return strings.TrimSpace(content)
}

func parseCodeFilesFromXML(ctx context.Context, content string) map[string]string {
	matches := xmlFileRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	files := make(map[string]string, len(matches))
	for _, m := range matches {
		path := strings.TrimSpace(m[1])
		code := m[2]
		if path != "" && len(code) > 0 {
			files[path] = code
		}
	}
	if len(files) == 0 {
		return nil
	}
	slog.Default().InfoContext(ctx, "parseCodeFiles strategy", "strategy", "xml", "files", len(files))

	return files
}

func parseCodeFilesFromJSON(ctx context.Context, content string) map[string]string {
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	first := strings.Index(content, "{")
	if first == -1 {
		return nil
	}
	last := strings.LastIndex(content, "}")
	if last > first {
		var files map[string]string
		if json.Unmarshal([]byte(content[first:last+1]), &files) == nil && len(files) > 0 {
			slog.Default().DebugContext(ctx, "parseCodeFiles strategy", "strategy", "json", "files", len(files))

			return files
		}
	}
	if last > first {
		fixed := strings.ReplaceAll(content[first:last+1], "\t", "\\t")
		if !strings.HasSuffix(strings.TrimSpace(fixed), "}") {
			fixed = strings.TrimSpace(fixed) + "\"}"
		}
		var files map[string]string
		if json.Unmarshal([]byte(fixed), &files) == nil && len(files) > 0 {
			slog.Default().DebugContext(ctx, "parseCodeFiles strategy", "strategy", "json_fixed", "files", len(files))

			return files
		}
	}
	if recovered := recoverTruncatedJSON(content[first:]); len(recovered) > 0 {
		slog.Default().DebugContext(ctx, "parseCodeFiles strategy", "strategy", "json_truncated", "files", len(recovered))

		return recovered
	}

	return nil
}

func extractRawHTMLFiles(src string) map[string]string {
	if htmlIdx := strings.Index(src, "<!DOCTYPE"); htmlIdx != -1 {
		htmlEnd := strings.LastIndex(src, "</html>")
		if htmlEnd != -1 {
			html := src[htmlIdx : htmlEnd+len("</html>")]
			slog.Info(fmt.Sprintf("✅ parseCodeFiles: strategy 6 (raw HTML) — %d chars", len(html)))

			return map[string]string{"index.html": html}
		}
		slog.Info(fmt.Sprintf("✅ parseCodeFiles: strategy 6 (raw HTML, no closing tag) — %d chars", len(src[htmlIdx:])))

		return map[string]string{"index.html": src[htmlIdx:]}
	}
	if htmlIdx := strings.Index(src, "<html"); htmlIdx != -1 {
		slog.Info(fmt.Sprintf("✅ parseCodeFiles: strategy 6 (raw <html>) — %d chars", len(src[htmlIdx:])))

		return map[string]string{"index.html": src[htmlIdx:]}
	}

	return nil
}

func extractIndexHTMLFromJSON(content string) map[string]string {
	_, after, ok := strings.Cut(content, `"index.html"`)
	if !ok {
		return nil
	}
	rest := after
	colonIdx := strings.Index(rest, ":")
	if colonIdx == -1 {
		return nil
	}
	rest = strings.TrimSpace(rest[colonIdx+1:])
	if len(rest) == 0 || rest[0] != '"' {
		return nil
	}
	html := extractJSONStringValue(rest)
	if len(html) <= 50 {
		return nil
	}
	slog.Info(fmt.Sprintf("✅ parseCodeFiles: strategy 5 (manual extract) — %d chars", len(html)))

	return map[string]string{"index.html": html}
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

// recoverTruncatedJSON extracts complete "key": "value" pairs from truncated JSON.
// Used when LLM hits max_tokens and the JSON is cut mid-string-value.
// Returns all files that were fully written before the truncation point.
func recoverTruncatedJSON(s string) map[string]string {
	if len(s) < 5 || s[0] != '{' {
		return nil
	}

	files := make(map[string]string)
	i := 1 // skip opening {

	for i < len(s) {
		key, value, next, ok := parseTruncatedJSONEntry(s, i)
		if !ok {
			break
		}
		i = next
		if len(key) > 0 && len(value) > 20 {
			files[key] = value
		}
	}

	if len(files) < 1 {
		return nil
	}

	return files
}

func parseTruncatedJSONEntry(s string, i int) (key, value string, next int, ok bool) {
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r' || s[i] == ',') {
		i++
	}
	if i >= len(s) || s[i] == '}' {
		return "", "", i, false
	}
	if s[i] != '"' {
		return "", "", i, false
	}
	key = extractJSONStringAt(s, i)
	if key == "" {
		return "", "", i, false
	}
	i += jsonStringLen(s, i)

	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	if i >= len(s) || s[i] != ':' {
		return "", "", i, false
	}
	i++

	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	if i >= len(s) || s[i] != '"' {
		return "", "", i, false
	}
	value = extractJSONStringAt(s, i)
	valLen := jsonStringLen(s, i)
	if valLen == 0 {
		return "", "", i, false
	}

	return key, value, i + valLen, true
}

// extractJSONStringAt extracts the decoded string value at position i (must start with ").
// Returns "" if the string is truncated (no closing quote found).
func extractJSONStringAt(s string, pos int) string {
	if pos >= len(s) || s[pos] != '"' {
		return ""
	}
	var b strings.Builder
	i := pos + 1 // skip opening quote
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
			case 'u':
				// Unicode escape \uXXXX — write as-is for simplicity
				if i+5 < len(s) {
					b.WriteString(s[i : i+6])
					i += 6
				} else {
					i += 2
				}

				continue
			default:
				b.WriteByte('\\')
				b.WriteByte(next)
			}
			i += 2

			continue
		}
		if ch == '"' {
			return b.String() // complete string
		}
		b.WriteByte(ch)
		i++
	}

	return "" // truncated — no closing quote
}

// jsonStringLen returns the raw byte length of a JSON string at pos (including quotes).
// Returns 0 if truncated.
func jsonStringLen(s string, pos int) int {
	if pos >= len(s) || s[pos] != '"' {
		return 0
	}
	i := pos + 1
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			i += 2

			continue
		}
		if s[i] == '"' {
			return i - pos + 1 // include both quotes
		}
		i++
	}

	return 0 // truncated
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
		slog.Info(fmt.Sprintf("🚨 parseMasterPlan: NO JSON object found | total_len=%d | head=%q", origLen, head))

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

	err := json.Unmarshal([]byte(jsonBlock), &parsed)
	if err != nil {
		head := jsonBlock
		if len(head) > 500 {
			head = head[:500]
		}
		slog.Info(fmt.Sprintf("🚨 parseMasterPlan: JSON unmarshal error: %v | total_len=%d block_len=%d | head=%q",
			err, origLen, len(jsonBlock), head))

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
		slog.Info(fmt.Sprintf("⚠️ parseMasterPlan: degenerate plan (%d tasks), substituting default DAG", len(plan.DAG)))
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
	slog.Info(fmt.Sprintf("✅ parseMasterPlan: %d steps, %d DAG tasks", len(plan.Steps), len(plan.DAG)))

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
	slog.Info(fmt.Sprintf("✅ Brain: strategy synthesized (%d chars)", len(result)))

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
