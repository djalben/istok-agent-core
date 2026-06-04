package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/djalben/istok-agent-core/internal/ports"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — ComponentEditor (Surgical Edit)
//  Принимает один файл + промпт → возвращает обновлённый код.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const componentEditorSystemPrompt = `Ты — AI-эксперт по React/HTML/CSS. Тебе передан текущий код одного файла и требование пользователя.
Твоя задача — вернуть ПОЛНЫЙ обновлённый код этого файла в формате XML-артифакта:

<file path="ПУТЬ_К_ФАЙЛУ">
...полный код файла...
</file>

КРИТИЧЕСКИЕ ПРАВИЛА:
1. Верни ТОЛЬКО один блок <file path="...">...</file>. Без пояснений, без markdown.
2. Код должен быть ПОЛНЫМ — не пропускай строки, не используй "// остальное без изменений".
3. Выполняй ТОЛЬКО то, что просит пользователь. Не меняй ничего лишнего.
4. Сохраняй стиль, импорты и структуру оригинала.`

// componentFileRegex matches <file path="...">...</file> (same pattern as agent_helpers.go).
var componentFileRegex = regexp.MustCompile(`(?s)<file\s+path="([^"]+)"\s*>\s*(.*?)\s*</file>`)

// ComponentEditor — точечное редактирование одного компонента/файла через LLM.
type ComponentEditor struct {
	llm ports.LLMProvider
}

// NewComponentEditor создаёт экземпляр.
func NewComponentEditor(llm ports.LLMProvider) *ComponentEditor {
	return &ComponentEditor{llm: llm}
}

// ComponentEditRequest — входные данные для точечной правки.
type ComponentEditRequest struct {
	FilePath    string `json:"file_path"`
	CurrentCode string `json:"current_code"`
	Prompt      string `json:"prompt"`
}

// ComponentEditResponse — результат правки.
type ComponentEditResponse struct {
	FilePath string `json:"file_path"`
	NewCode  string `json:"new_code"`
}

// Edit выполняет точечную правку одного файла.
func (ce *ComponentEditor) Edit(ctx context.Context, req ComponentEditRequest) (*ComponentEditResponse, error) {
	if req.Prompt == "" {
		return nil, ErrEmptyEditPrompt
	}
	if req.CurrentCode == "" {
		return nil, ErrEmptyCurrentCode
	}

	userPrompt := fmt.Sprintf("## Файл: %s\n```\n%s\n```\n\n## Требование пользователя:\n%s", req.FilePath, req.CurrentCode, req.Prompt)
	slog.Info(fmt.Sprintf("🔧 ComponentEditor: file=%s, prompt=%q", req.FilePath, req.Prompt))

	resp, err := ce.llm.Complete(ctx, ports.LLMRequest{
		Model:        "anthropic/claude-sonnet-4-6",
		SystemPrompt: componentEditorSystemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    8192,
		Temperature:  0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("component edit LLM call failed: %w", err)
	}

	// Parse XML artifact
	content := strings.TrimSpace(resp.Content)
	matches := componentFileRegex.FindStringSubmatch(content)
	if matches == nil {
		slog.
			// Fallback: if LLM returned raw code without XML wrapper, use as-is
			Info(fmt.Sprintf("⚠️ ComponentEditor: no XML artifact found, using raw response (%d chars)", len(content)))
		// Strip markdown fences
		content = stripJSONFences(content)

		return &ComponentEditResponse{
			FilePath: req.FilePath,
			NewCode:  content,
		}, nil
	}

	filePath := strings.TrimSpace(matches[1])
	newCode := matches[2]
	slog.Info(fmt.Sprintf("✅ ComponentEditor: edited %s (%d → %d chars)", filePath, len(req.CurrentCode), len(newCode)))

	return &ComponentEditResponse{
		FilePath: filePath,
		NewCode:  newCode,
	}, nil
}
