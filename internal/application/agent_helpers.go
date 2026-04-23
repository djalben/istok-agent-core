package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Agent Helpers — shared LLM helpers for Director + Coder
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// callLLM sends a chat-completion request to OpenRouter and returns the text response.
// Shared by Director (createMasterPlan) and Coder (generateCode).
func (o *Orchestrator) callLLM(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	return o.callLLMInternal(ctx, model, systemPrompt, userPrompt, maxTokens, false, 0)
}

// callLLMWithReasoning sends a request with extended reasoning/thinking enabled.
// Used for agents that need deep architectural reasoning (Gemini 3 Pro via Replicate).
func (o *Orchestrator) callLLMWithReasoning(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens, thinkingBudget int) (string, error) {
	return o.callLLMInternal(ctx, model, systemPrompt, userPrompt, maxTokens, true, thinkingBudget)
}

// callLLMInternal is the shared implementation for all LLM calls.
// Dual routing: Anthropic models → Replicate, everything else → OpenRouter.
func (o *Orchestrator) callLLMInternal(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int, reasoning bool, thinkingBudget int) (string, error) {
	// ── Проверка: если клиент уже отключился — не тратим кредиты ──
	select {
	case <-ctx.Done():
		log.Printf("⛔ ОТМЕНА: клиент отключился до вызова LLM model=%s", model)
		return "", fmt.Errorf("cancelled before LLM call: %w", ctx.Err())
	default:
	}

	// ── DUAL ROUTING: Anthropic+Google → Replicate, остальные → OpenRouter ──
	if isReplicateModel(model) {
		log.Printf("🔀 Routing %s → Replicate", model)
		temp := 0.7
		if reasoning {
			temp = 1.0
		}
		return callReplicate(ctx, model, systemPrompt, userPrompt, maxTokens, temp)
	}

	// ── OpenRouter path (DeepSeek, Gemini, etc.) ──
	if o.apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY not configured")
	}

	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	var messages []msg
	if systemPrompt != "" {
		messages = append(messages, msg{Role: "system", Content: systemPrompt})
	}
	messages = append(messages, msg{Role: "user", Content: userPrompt})

	payload := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": 0.7,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal failed: %w", err)
	}

	openRouterURL := os.Getenv("OPENROUTER_PROXY_URL")
	if openRouterURL == "" {
		openRouterURL = "https://openrouter.ai/api/v1"
	}
	body, status, err := httpPost(ctx, openRouterURL+"/chat/completions", o.apiKey, payloadBytes)
	if err != nil {
		return "", fmt.Errorf("LLM request failed: %w", err)
	}
	if status != 200 {
		maxLog := len(body)
		if maxLog > 400 {
			maxLog = 400
		}
		log.Printf("🚨 LLM error | model=%s status=%d | %s", model, status, string(body[:maxLog]))
		return "", fmt.Errorf("LLM API error (HTTP %d): %s", status, string(body[:maxLog]))
	}

	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse response failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from model")
	}
	return resp.Choices[0].Message.Content, nil
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
func (o *Orchestrator) parseMasterPlan(content, spec string, audit *ReverseEngineeringResult) *MasterPlan {
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

	var parsed struct {
		Architecture string   `json:"architecture"`
		Components   []string `json:"components"`
		Technologies []string `json:"technologies"`
		Timeline     string   `json:"timeline"`
		Steps        []string `json:"steps"`
	}

	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		log.Printf("⚠️ parseMasterPlan JSON error: %v", err)
		return o.defaultMasterPlan(spec, audit)
	}

	plan := &MasterPlan{
		Architecture: parsed.Architecture,
		Components:   parsed.Components,
		Technologies: parsed.Technologies,
		Timeline:     parsed.Timeline,
		Steps:        parsed.Steps,
	}
	if plan.Architecture == "" {
		plan.Architecture = spec
	}
	if len(plan.Steps) == 0 {
		plan.Steps = []string{spec}
	}
	return plan
}

// synthesizeStrategy asks Gemini Brain to produce a concise strategic brief
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

// validateAndHeal checks generated HTML files for common issues and auto-fixes via LLM if needed.
// Returns the (possibly fixed) files map. Max 1 heal attempt to avoid loops.
func (o *Orchestrator) validateAndHeal(ctx context.Context, files map[string]string, spec string) map[string]string {
	html, ok := files["index.html"]
	if !ok || len(html) < 50 {
		return files
	}

	var issues []string
	if !strings.Contains(html, "<!DOCTYPE") && !strings.Contains(html, "<!doctype") {
		issues = append(issues, "missing <!DOCTYPE html>")
	}
	if !strings.Contains(html, "<body") {
		issues = append(issues, "missing <body> tag")
	}
	if !strings.Contains(strings.ToLower(html), "tailwind") {
		issues = append(issues, "missing TailwindCSS CDN — add <script src=\"https://cdn.tailwindcss.com\"></script>")
	}
	if strings.Contains(html, "Lorem ipsum") || strings.Contains(html, "lorem ipsum") {
		issues = append(issues, "contains Lorem Ipsum placeholder text — use real content")
	}
	if len(html) < 500 {
		issues = append(issues, "generated HTML is suspiciously short (< 500 chars)")
	}

	if len(issues) == 0 {
		log.Printf("✅ validateAndHeal: no issues found")
		return files
	}

	log.Printf("🩺 validateAndHeal: %d issues: %v — auto-fixing", len(issues), issues)
	o.sendStatus(RoleCoder, "running", fmt.Sprintf("🩺 Auto-healing %d проблем в коде...", len(issues)), 85)

	healPrompt := fmt.Sprintf(`Fix these issues in the HTML code below:

ISSUES: %s

SPECIFICATION: %s

CODE TO FIX:
%s

Return ONLY the fixed complete HTML file. No JSON wrapper, no markdown fences. Start with <!DOCTYPE html>.`,
		strings.Join(issues, "; "), spec, html)

	fixed, err := o.callLLM(ctx, "qwen/qwen-2.5-72b-instruct",
		"You are a frontend code fixer. Return only valid, complete HTML. No explanations.",
		healPrompt, 16000)

	if err != nil {
		log.Printf("⚠️ validateAndHeal: fix failed: %v", err)
		return files
	}

	fixed = strings.TrimSpace(fixed)
	fixed = strings.TrimPrefix(fixed, "```html")
	fixed = strings.TrimPrefix(fixed, "```")
	fixed = strings.TrimSuffix(fixed, "```")
	fixed = strings.TrimSpace(fixed)

	if strings.Contains(fixed, "<!DOCTYPE") || strings.Contains(fixed, "<html") {
		files["index.html"] = fixed
		log.Printf("✅ validateAndHeal: code auto-fixed successfully (%d chars)", len(fixed))
		o.sendStatus(RoleCoder, "running", "✅ Код автоматически исправлен", 90)
	}

	return files
}

// validatorCheck runs the ValidatorAgent — programmatic syntax & runtime check.
// Detects unclosed tags, missing elements, JS errors, and auto-fixes via LLM.
func (o *Orchestrator) validatorCheck(ctx context.Context, files map[string]string, spec string) map[string]string {
	html, ok := files["index.html"]
	if !ok || len(html) < 100 {
		return files
	}

	var errors []string

	// Syntax checks
	openTags := 0
	for _, ch := range html {
		if ch == '<' {
			openTags++
		}
		if ch == '>' {
			openTags--
		}
	}
	if openTags != 0 {
		errors = append(errors, fmt.Sprintf("unbalanced HTML tags (delta=%d)", openTags))
	}

	// Must-have elements
	lowerHTML := strings.ToLower(html)
	if !strings.Contains(lowerHTML, "<html") {
		errors = append(errors, "missing <html> root element")
	}
	if !strings.Contains(lowerHTML, "<head") {
		errors = append(errors, "missing <head> section")
	}
	if !strings.Contains(lowerHTML, "<body") {
		errors = append(errors, "missing <body> section")
	}
	if !strings.Contains(lowerHTML, "</html>") {
		errors = append(errors, "missing closing </html> tag")
	}

	// JS runtime risk checks
	if strings.Contains(html, "document.getElementById") && !strings.Contains(html, "DOMContentLoaded") && !strings.Contains(html, "defer") {
		errors = append(errors, "JS uses getElementById without DOMContentLoaded or defer — runtime risk")
	}
	if strings.Count(html, "<script") != strings.Count(html, "</script>") {
		errors = append(errors, "unclosed <script> tag")
	}
	if strings.Count(html, "<style") != strings.Count(html, "</style>") {
		errors = append(errors, "unclosed <style> tag")
	}

	// Encoding safety
	if strings.Contains(html, "\\u00") || strings.Contains(html, "\\x") {
		errors = append(errors, "suspicious escape sequences that may cause rendering issues")
	}

	if len(errors) == 0 {
		log.Printf("✅ ValidatorAgent: all checks passed")
		return files
	}

	log.Printf("🛡️ ValidatorAgent: %d issues found: %v", len(errors), errors)
	o.sendStatus(RoleValidator, "running", fmt.Sprintf("🛡️ Найдено %d ошибок, авто-исправление...", len(errors)), 95)

	agent := o.agents[RoleValidator]
	fixPrompt := fmt.Sprintf(`You are a code validator. Fix ALL these issues in the HTML below.

ISSUES FOUND:
%s

SPECIFICATION: %s

CODE:
%s

Return ONLY the complete fixed HTML. No markdown, no explanation. Start with <!DOCTYPE html>.`,
		strings.Join(errors, "\n"), spec, html)

	fixed, err := o.callLLM(ctx, agent.Model,
		"You are an expert code validator and fixer. Fix all issues. Return complete HTML only.",
		fixPrompt, 32000)

	if err != nil {
		log.Printf("⚠️ ValidatorAgent fix failed: %v", err)
		return files
	}

	fixed = strings.TrimSpace(fixed)
	fixed = strings.TrimPrefix(fixed, "```html")
	fixed = strings.TrimPrefix(fixed, "```")
	fixed = strings.TrimSuffix(fixed, "```")
	fixed = strings.TrimSpace(fixed)

	if strings.Contains(fixed, "<!DOCTYPE") || strings.Contains(fixed, "<html") {
		files["index.html"] = fixed
		log.Printf("✅ ValidatorAgent: code fixed successfully (%d chars)", len(fixed))
	}

	return files
}

// defaultMasterPlan returns a sensible fallback plan when Director API fails.
func (o *Orchestrator) defaultMasterPlan(spec string, audit *ReverseEngineeringResult) *MasterPlan {
	plan := &MasterPlan{
		Architecture: spec,
		Components:   []string{"Hero Section", "Navigation", "Feature Cards", "CTA", "Footer"},
		Technologies: []string{"HTML5", "CSS3", "JavaScript"},
		Timeline:     "immediate",
		Steps:        []string{spec},
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
