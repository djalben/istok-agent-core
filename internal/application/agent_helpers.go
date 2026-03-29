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
//  Agent Helpers — shared LLM helpers for Director + Coder
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// callLLM sends a chat-completion request to OpenRouter and returns the text response.
// Shared by Director (createMasterPlan) and Coder (generateCode).
func (o *Orchestrator) callLLM(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int) (string, error) {
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

	const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"
	body, status, err := httpPost(ctx, openRouterURL, o.apiKey, payloadBytes)
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
// Handles <thinking> blocks, markdown fences, and JSON extraction.
func (o *Orchestrator) parseCodeFiles(content string) map[string]string {
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

	// Extract JSON between first { and last }
	first := strings.Index(content, "{")
	last := strings.LastIndex(content, "}")
	if first == -1 || last <= first {
		return nil
	}
	content = content[first : last+1]

	var files map[string]string
	if err := json.Unmarshal([]byte(content), &files); err != nil {
		log.Printf("⚠️ parseCodeFiles JSON error: %v | len=%d", err, len(content))
		return nil
	}
	return files
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

// synthesizeStrategy asks Claude Brain to produce a concise strategic brief
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
