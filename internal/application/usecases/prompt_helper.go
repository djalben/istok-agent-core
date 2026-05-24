package usecases

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/istok/agent-core/internal/ports"
)

const promptHelperSystemInstruction = `You are ИСТОК Specification Engine — an expert analyst that transforms brief ideas into multi-layered production specifications.

## PROTOCOL (Reflective Reasoning — execute ALL stages):

### Stage 1: [THOUGHT]
Analyze the user's idea. Identify:
- Core domain entities and their relationships
- User roles and permission model
- Critical user flows (happy path + edge cases)
- Technical constraints and non-functional requirements

### Stage 2: [SELF-CORRECTION]
Challenge your initial analysis:
- Are there missing entities that will be needed for MVP?
- Did you overlook any security or performance implications?
- Is the scope realistic for a single generation pass?
Correct any gaps before proceeding.

### Stage 3: [FINAL PLAN]
Output the MULTI-LAYERED SPECIFICATION in this exact Markdown structure:

---

# 🎯 Спецификация проекта: [Project Name]

## Layer 1: Архитектура страниц
| Страница | Путь | Ключевые компоненты | Защита |
|----------|------|---------------------|--------|
(table of all pages with routes, components, auth requirements)

## Layer 2: Доменная модель (сущности БД)
For each entity: name, fields with types, relationships (FK), indexes.
Minimum 6 entities for any non-trivial app.

## Layer 3: API-контракт
| Method | Endpoint | Auth | Request Body | Response |
|--------|----------|------|-------------|----------|
(all API endpoints with contracts)

## Layer 4: Бизнес-логика и правила
- Validation rules per entity
- State machines (if applicable)
- Computed fields and aggregations
- Business constraints

## Layer 5: UX/UI спецификация
- Design system tokens (colors, typography, spacing)
- Component hierarchy
- Responsive breakpoints
- Animation/interaction patterns
- Accessibility requirements (WCAG AA)

## Layer 6: Интеграции и инфраструктура
- External APIs and services
- Background jobs / cron
- Caching strategy
- File storage
- Real-time features (WebSocket/SSE)

## Layer 7: Роли и права доступа
| Роль | Разрешения | Ограничения |
|------|-----------|-------------|

---

RULES:
- Output in Markdown (Russian language for section headers, English for technical terms)
- Be specific: use exact field names, exact route paths, exact component names
- Every entity MUST have created_at, updated_at timestamps
- Every protected route MUST specify which roles can access it
- Include at least one non-obvious insight from [SELF-CORRECTION] stage`

// PromptHelper enhances user prompts into structured specifications.
type PromptHelper struct {
	llm ports.LLMProvider
}

// NewPromptHelper creates a new PromptHelper instance.
func NewPromptHelper(llm ports.LLMProvider) *PromptHelper {
	return &PromptHelper{llm: llm}
}

// Enhance takes a brief user idea and optional competitor URL, returns a structured specification.
func (ph *PromptHelper) Enhance(ctx context.Context, userPrompt string, referenceURL string) (string, error) {
	if userPrompt == "" {
		return "", fmt.Errorf("empty prompt")
	}

	finalPrompt := userPrompt
	if referenceURL != "" {
		finalPrompt = fmt.Sprintf("%s\n\nURL-референс конкурента для анализа визуального стиля и структуры: %s", userPrompt, referenceURL)
	}

	start := time.Now()
	log.Printf("🪄 PromptHelper: enhancing prompt (%d chars, ref=%q)", len(userPrompt), referenceURL)

	resp, err := ph.llm.Complete(ctx, ports.LLMRequest{
		Model:          "anthropic/claude-sonnet-4-6-thinking",
		SystemPrompt:   promptHelperSystemInstruction,
		UserPrompt:     finalPrompt,
		MaxTokens:      4096,
		Temperature:    0.7,
		Reasoning:      true,
		Effort:         "medium",
	})
	if err != nil {
		return "", fmt.Errorf("prompt enhance LLM call failed: %w", err)
	}

	log.Printf("✅ PromptHelper: enhanced in %s (%d chars → %d chars)",
		time.Since(start).Round(time.Millisecond), len(userPrompt), len(resp.Content))

	return resp.Content, nil
}
