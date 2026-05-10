package usecases

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/istok/agent-core/internal/ports"
)

const promptHelperSystemInstruction = `Ты — эксперт-аналитик системы ИСТОК. Преврати краткую идею пользователя в структурированную спецификацию. Обязательно укажи: структуру страниц, необходимые сущности БД, типы графиков, роли пользователей и интеграции. Используй Markdown.`

// PromptHelper enhances user prompts into structured specifications.
type PromptHelper struct {
	llm ports.LLMProvider
}

// NewPromptHelper creates a new PromptHelper instance.
func NewPromptHelper(llm ports.LLMProvider) *PromptHelper {
	return &PromptHelper{llm: llm}
}

// Enhance takes a brief user idea and returns a structured specification.
func (ph *PromptHelper) Enhance(ctx context.Context, userPrompt string) (string, error) {
	if userPrompt == "" {
		return "", fmt.Errorf("empty prompt")
	}

	start := time.Now()
	log.Printf("🪄 PromptHelper: enhancing prompt (%d chars)", len(userPrompt))

	resp, err := ph.llm.Complete(ctx, ports.LLMRequest{
		Model:        "anthropic/claude-opus-4-7-thinking",
		SystemPrompt: promptHelperSystemInstruction,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
		Temperature:  0.7,
		Reasoning:    true,
		ThinkingBudget: 2048,
	})
	if err != nil {
		return "", fmt.Errorf("prompt enhance LLM call failed: %w", err)
	}

	log.Printf("✅ PromptHelper: enhanced in %s (%d chars → %d chars)",
		time.Since(start).Round(time.Millisecond), len(userPrompt), len(resp.Content))

	return resp.Content, nil
}
