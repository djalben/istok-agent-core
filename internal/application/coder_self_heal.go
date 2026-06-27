package application

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Inner Self-Healing Loop
//  Точечное авто-исправление критических ошибок
//  внутри chunked-генерации, до возврата файлов.
//  Отличие от внешнего phaseAgentVerification:
//    - только сломанные файлы (не полная регенерация)
//    - быстрее: 1 LLM-вызов на попытку вместо всего pipeline
//    - запускается сразу после finalizeAppShell
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// maxSelfHealingRetries — жёсткий лимит попыток inner self-healing.
// Управляет только точечными исправлениями внутри generateCodeChunked.
// Внешний autoFixMaxRetries (полная регенерация) остаётся независимым.
const maxSelfHealingRetries = 2

// selfHealFiles — inner healing loop. Запускается после finalizeAppShell.
// Прогоняет ValidateCode; при critical-ошибках посылает целевой LLM-запрос
// только для сломанных файлов. Мутирует переданную карту files на месте.
func (o *Orchestrator) selfHealFiles(ctx context.Context, specification string, manifest *SystemManifest, files map[string]string) {
	for attempt := 1; attempt <= maxSelfHealingRetries; attempt++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		o.sendStatus(ctx, RoleCoder, "running",
			fmt.Sprintf("🔍 Self-healing: проверка кода (попытка %d/%d)...", attempt, maxSelfHealingRetries),
			93+attempt)

		result := usecases.ValidateCode(files)
		applog(ctx).InfoContext(ctx, "self-heal validate",
			"attempt", attempt,
			"passed", result.Passed,
			"criticals", result.CriticalCount(),
			"issues", len(result.Issues),
		)

		if result.Passed {
			o.sendStatus(ctx, RoleCoder, "running", "✅ Self-healing: код прошёл проверку", 95)

			return
		}

		brokenFiles := collectBrokenFiles(result)
		if len(brokenFiles) == 0 {
			return
		}

		o.sendStatus(ctx, RoleCoder, "running",
			fmt.Sprintf("🔧 Self-healing: исправляю %d файлов... (попытка %d/%d)",
				len(brokenFiles), attempt, maxSelfHealingRetries),
			93+attempt)

		agent := o.agents[RoleCoder]
		fixCtx, cancel := context.WithTimeout(ctx, perGroupLLMTimeout)
		userPrompt := buildSelfHealPrompt(specification, manifest, files, result, brokenFiles)
		content, err := o.callLLMWithReasoning(fixCtx, agent.Model, chunkedCoderSystemPrompt, userPrompt, 16384)
		cancel()

		if err != nil {
			applog(ctx).WarnContext(ctx, "self-heal fix call failed",
				"attempt", attempt,
				"error", wrapper.Wrap(err),
			)

			break
		}

		fixed := o.parseCodeFiles(ctx, content)
		applied := 0
		for name, code := range fixed {
			if _, exists := files[name]; exists && strings.TrimSpace(code) != "" {
				files[name] = code
				o.busFromCtx(ctx).PublishFile(RoleCoder, name, code)
				applied++
			}
		}
		applog(ctx).InfoContext(ctx, "self-heal applied",
			"attempt", attempt,
			"offered", len(fixed),
			"applied", applied,
		)
	}
}

// collectBrokenFiles возвращает отсортированный список файлов с critical-ошибками.
func collectBrokenFiles(result *usecases.ValidationResult) []string {
	seen := make(map[string]struct{})
	for _, iss := range result.Issues {
		if iss.Severity == usecases.SeverityCritical && iss.File != "" {
			seen[iss.File] = struct{}{}
		}
	}
	files := make([]string, 0, len(seen))
	for f := range seen {
		files = append(files, f)
	}
	sort.Strings(files)

	return files
}

// buildSelfHealPrompt формирует целевой промпт: только сломанные файлы + список ошибок.
// Контекст спеки обрезается до 600 символов — фокус на исправлении, не на архитектуре.
func buildSelfHealPrompt(
	specification string,
	manifest *SystemManifest,
	files map[string]string,
	result *usecases.ValidationResult,
	brokenFiles []string,
) string {
	var b strings.Builder

	b.WriteString("SELF-HEALING PASS: Fix the following critical validation errors.\n")
	b.WriteString("Output ONLY the fixed files in <file path=\"...\">...</file> format. Do NOT rewrite unaffected files.\n\n")

	errCtx := result.ForCoderContext()
	if errCtx != "" {
		b.WriteString(errCtx)
		b.WriteString("\n")
	}

	spec := specification
	if len(spec) > 600 {
		spec = spec[:600] + "..."
	}
	fmt.Fprintf(&b, "PROJECT SPEC (reference):\n%s\n\n", spec)

	if manifest != nil && manifest.ProjectName != "" {
		fmt.Fprintf(&b, "PROJECT: %s\n\n", manifest.ProjectName)
	}

	b.WriteString("BROKEN FILES TO FIX (current content):\n")
	for _, name := range brokenFiles {
		content := files[name]
		if len(content) > 4000 {
			content = content[:4000] + "\n// ... (truncated)"
		}
		fmt.Fprintf(&b, "\n--- %s ---\n%s\n", name, content)
	}

	b.WriteString("\nFix ONLY the critical issues listed above. Return complete, corrected file contents.\n")

	return b.String()
}
