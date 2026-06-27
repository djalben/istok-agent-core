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
//
//  Два триггера (оба проверяются на каждой итерации):
//    1. ValidateCode — Security + Quality критические ошибки
//    2. CheckCrossFileIntegrity — сломанные @/* импорты (причина #1 белого экрана)
//
//  Integrity-triggered healing разрешает создание НОВЫХ файлов:
//  LLM может сгенерировать недостающий модуль, а не только починить существующий.
//
//  Отличие от внешнего phaseAgentVerification:
//    - только сломанные файлы (не полная регенерация)
//    - 1 LLM-вызов на попытку вместо всего pipeline
//    - запускается сразу после finalizeAppShell
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// maxSelfHealingRetries — жёсткий лимит попыток inner self-healing.
// Управляет только точечными исправлениями внутри generateCodeChunked.
// Внешний autoFixMaxRetries (полная регенерация) остаётся независимым.
const maxSelfHealingRetries = 2

// minIntegrityMissing — порог недостающих импортов для активации LLM-исправления.
// Меньше этого порога — BackfillMissingImports (stubs) справляется сам.
const minIntegrityMissing = 3

// selfHealFiles — inner healing loop. Запускается после finalizeAppShell.
// Прогоняет ValidateCode + CheckCrossFileIntegrity; при проблемах посылает
// целевой LLM-запрос для сломанных файлов. Мутирует переданную карту files на месте.
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
		integrity := usecases.CheckCrossFileIntegrity(files)

		integrityBroken := len(integrity.MissingFiles) >= minIntegrityMissing
		applog(ctx).InfoContext(ctx, "self-heal validate",
			"attempt", attempt,
			"validationPassed", result.Passed,
			"criticals", result.CriticalCount(),
			"integrityValid", integrity.Valid,
			"missingImports", len(integrity.MissingFiles),
			"integrityTrigger", integrityBroken,
		)

		if result.Passed && !integrityBroken {
			o.sendStatus(ctx, RoleCoder, "running", "✅ Self-healing: код прошёл проверку", 95)

			return
		}

		brokenFiles := collectBrokenFiles(result)
		allowNewFiles := false
		if integrityBroken {
			allowNewFiles = true
			brokenFiles = mergeUniq(brokenFiles, extractImportingFiles(integrity))
		}

		if len(brokenFiles) == 0 {
			return
		}

		o.sendStatus(ctx, RoleCoder, "running",
			fmt.Sprintf("🔧 Self-healing: исправляю %d файлов... (попытка %d/%d)",
				len(brokenFiles), attempt, maxSelfHealingRetries),
			93+attempt)

		agent := o.agents[RoleCoder]
		fixCtx, cancel := context.WithTimeout(ctx, perGroupLLMTimeout)
		userPrompt := buildSelfHealPrompt(specification, manifest, files, result, integrity, brokenFiles)
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
			isExisting := files[name] != ""
			if (isExisting || allowNewFiles) && strings.TrimSpace(code) != "" {
				files[name] = code
				o.busFromCtx(ctx).PublishFile(RoleCoder, name, code)
				applied++
			}
		}
		applog(ctx).InfoContext(ctx, "self-heal applied",
			"attempt", attempt,
			"offered", len(fixed),
			"applied", applied,
			"newFilesAllowed", allowNewFiles,
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

// extractImportingFiles извлекает пути файлов, которые содержат сломанные импорты.
// IntegrityResult.MissingFiles содержит строки вида "@/components/Foo (in src/App.tsx)".
func extractImportingFiles(integrity *usecases.IntegrityResult) []string {
	seen := make(map[string]struct{})
	for _, entry := range integrity.MissingFiles {
		// Формат: "@/components/Foo (in src/pages/Bar.tsx)"
		if i := strings.Index(entry, " (in "); i >= 0 {
			importer := strings.TrimSuffix(entry[i+5:], ")")
			importer = strings.TrimSpace(importer)
			if importer != "" {
				seen[importer] = struct{}{}
			}
		}
	}
	files := make([]string, 0, len(seen))
	for f := range seen {
		files = append(files, f)
	}
	sort.Strings(files)

	return files
}

// mergeUniq объединяет два среза строк без дублей, сохраняя порядок.
func mergeUniq(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	result := make([]string, 0, len(a)+len(b))
	for _, s := range a {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	for _, s := range b {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}

	return result
}

// buildSelfHealPrompt формирует целевой промпт: ошибки валидации + integrity + содержимое файлов.
// При наличии integrity-проблем — явно разрешает LLM создать недостающие модули.
func buildSelfHealPrompt(
	specification string,
	manifest *SystemManifest,
	files map[string]string,
	result *usecases.ValidationResult,
	integrity *usecases.IntegrityResult,
	brokenFiles []string,
) string {
	var b strings.Builder

	integrityBroken := len(integrity.MissingFiles) >= minIntegrityMissing
	b.WriteString("SELF-HEALING PASS: Fix the following issues.\n")
	if integrityBroken {
		b.WriteString("You MAY create NEW files to resolve missing module imports — output them as <file path=\"...\">...</file> blocks.\n")
	}
	b.WriteString("Output ONLY fixed/new files in <file path=\"...\">...</file> format. Do NOT rewrite unaffected files.\n\n")

	errCtx := result.ForCoderContext()
	if errCtx != "" {
		b.WriteString(errCtx)
		b.WriteString("\n")
	}

	if integrityBroken {
		b.WriteString("## MISSING MODULE IMPORTS (create these files OR fix the broken imports)\n\n")
		for _, m := range integrity.MissingFiles {
			fmt.Fprintf(&b, "- %s\n", m)
		}
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

	if len(brokenFiles) > 0 {
		b.WriteString("FILES TO FIX (current content):\n")
		for _, name := range brokenFiles {
			content := files[name]
			if content == "" {
				continue
			}
			if len(content) > 4000 {
				content = content[:4000] + "\n// ... (truncated)"
			}
			fmt.Fprintf(&b, "\n--- %s ---\n%s\n", name, content)
		}
	}

	b.WriteString("\nFix ONLY the issues listed above. Return complete, corrected file contents.\n")

	return b.String()
}
