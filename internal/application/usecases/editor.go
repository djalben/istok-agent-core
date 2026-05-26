package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/istok/agent-core/internal/ports"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Editor Agent (Chat-to-Modify)
//  Принимает файлы + сообщение → возвращает JSON-патчи
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const editorSystemPrompt = `Ты — AI-редактор кода. Тебе передано текущее состояние файлов проекта и требование пользователя.
Твоя задача — вернуть СТРОГО валидный JSON массив объектов с правками, без markdown-огней и лишнего текста.
Формат объекта: {"filePath": "путь_к_файлу", "search": "точный фрагмент старого кода", "replace": "новый код"}.
Заменяй только то, что просит пользователь. Не трогай файлы, которые не затронуты запросом.

КРИТИЧЕСКИЕ ПРАВИЛА:
1. Ответ — ТОЛЬКО JSON массив. Без markdown, без пояснений, без обёрток.
2. Поле "search" должно содержать ТОЧНЫЙ фрагмент из текущего файла (с точностью до пробелов и переносов строк).
3. Если нужно добавить новый код, используй пустую строку в "search" и укажи "insertAfter" с якорной строкой.
4. Если нужно создать новый файл, используй пустой "search" и полное содержимое в "replace".
5. Минимизируй объём правок — точечные замены, а не перезапись файлов целиком.`

// FilePatch — одна точечная правка файла.
type FilePatch struct {
	FilePath string `json:"filePath"`
	Search   string `json:"search"`
	Replace  string `json:"replace"`
}

// Editor — агент для интерактивного редактирования кода через чат.
type Editor struct {
	llm ports.LLMProvider
}

// NewEditor создаёт экземпляр агента-редактора.
func NewEditor(llm ports.LLMProvider) *Editor {
	return &Editor{llm: llm}
}

// Edit принимает текущие файлы и сообщение пользователя, возвращает массив патчей.
func (e *Editor) Edit(ctx context.Context, message string, files map[string]string) ([]FilePatch, error) {
	if message == "" {
		return nil, fmt.Errorf("empty editor message")
	}

	// Собираем контекст файлов для промпта
	var filesCtx strings.Builder
	filesCtx.WriteString("## Текущие файлы проекта:\n\n")
	for path, content := range files {
		// Ограничиваем размер файла в промпте (макс. 8000 символов на файл)
		truncated := content
		if len(truncated) > 8000 {
			truncated = truncated[:8000] + "\n... (truncated)"
		}
		filesCtx.WriteString(fmt.Sprintf("### %s\n```\n%s\n```\n\n", path, truncated))
	}

	userPrompt := fmt.Sprintf("%s\n## Запрос пользователя:\n%s", filesCtx.String(), message)

	log.Printf("🖊️ Editor: processing request, files=%d, message=%q", len(files), message)

	resp, err := e.llm.Complete(ctx, ports.LLMRequest{
		Model:        "anthropic/claude-sonnet-4-6",
		SystemPrompt: editorSystemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
		Temperature:  0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("editor LLM call failed: %w", err)
	}

	// Парсим JSON из ответа (strip markdown fences if LLM wraps anyway)
	raw := strings.TrimSpace(resp.Content)
	raw = stripJSONFences(raw)

	var patches []FilePatch
	if err := json.Unmarshal([]byte(raw), &patches); err != nil {
		log.Printf("⚠️ Editor: failed to parse JSON patches: %v\nRaw response: %s", err, raw[:min(len(raw), 500)])
		return nil, fmt.Errorf("failed to parse editor response as JSON: %w", err)
	}

	log.Printf("✅ Editor: generated %d patches", len(patches))
	return patches, nil
}

// stripJSONFences removes markdown code fences if LLM wraps the JSON.
func stripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	// Remove ```json ... ``` or ``` ... ```
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if strings.HasSuffix(s, "```") {
			s = s[:len(s)-3]
		}
	}
	return strings.TrimSpace(s)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
