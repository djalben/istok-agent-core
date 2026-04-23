package application

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Auto-Healer — Railway Log Parser & Fix Generator
//  Разбирает логи ошибок и формулирует команды на исправление.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// BuildError — структурированная ошибка билда
type BuildError struct {
	Type    string // "compile", "runtime", "deploy", "timeout", "oom"
	File    string // файл с ошибкой
	Line    int    // строка
	Message string // текст ошибки
	Raw     string // сырой лог
}

// HealCommand — команда на исправление
type HealCommand struct {
	Action      string // "fix_code", "add_import", "increase_timeout", "fix_env"
	Target      string // файл или переменная
	Description string // что делать
	Priority    int    // 1=critical, 2=high, 3=medium
}

// ParseRailwayLogs разбирает лог Railway и извлекает структурированные ошибки.
func ParseRailwayLogs(rawLog string) []BuildError {
	var errors []BuildError
	lines := strings.Split(rawLog, "\n")

	for i, line := range lines {
		lower := strings.ToLower(line)

		// Go compile errors: "file.go:42:10: undefined: SomeFunc"
		if strings.Contains(line, ".go:") && (strings.Contains(lower, "undefined") ||
			strings.Contains(lower, "cannot") ||
			strings.Contains(lower, "syntax error") ||
			strings.Contains(lower, "imported and not used") ||
			strings.Contains(lower, "declared and not used")) {
			errors = append(errors, BuildError{
				Type:    "compile",
				Message: strings.TrimSpace(line),
				Raw:     contextLines(lines, i, 2),
			})
		}

		// Runtime panic
		if strings.Contains(lower, "panic:") || strings.Contains(lower, "goroutine") {
			errors = append(errors, BuildError{
				Type:    "runtime",
				Message: strings.TrimSpace(line),
				Raw:     contextLines(lines, i, 5),
			})
		}

		// Deploy/build failures
		if strings.Contains(lower, "error building") ||
			strings.Contains(lower, "build failed") ||
			strings.Contains(lower, "nixpacks") && strings.Contains(lower, "error") {
			errors = append(errors, BuildError{
				Type:    "deploy",
				Message: strings.TrimSpace(line),
				Raw:     contextLines(lines, i, 3),
			})
		}

		// Timeout
		if strings.Contains(lower, "context deadline exceeded") ||
			strings.Contains(lower, "timeout") {
			errors = append(errors, BuildError{
				Type:    "timeout",
				Message: strings.TrimSpace(line),
				Raw:     contextLines(lines, i, 2),
			})
		}

		// OOM
		if strings.Contains(lower, "out of memory") || strings.Contains(lower, "oom") {
			errors = append(errors, BuildError{
				Type:    "oom",
				Message: strings.TrimSpace(line),
				Raw:     contextLines(lines, i, 2),
			})
		}

		// Missing env vars
		if strings.Contains(lower, "not configured") ||
			(strings.Contains(lower, "env") && strings.Contains(lower, "empty")) {
			errors = append(errors, BuildError{
				Type:    "deploy",
				Message: strings.TrimSpace(line),
				Raw:     contextLines(lines, i, 1),
			})
		}
	}

	return dedup(errors)
}

// DiagnoseAndHeal принимает сырой лог, парсит ошибки, формулирует исправления.
// Если нужен LLM для сложных ошибок — вызывает ValidatorAgent.
func (o *Orchestrator) DiagnoseAndHeal(ctx context.Context, rawLog string) []HealCommand {
	errors := ParseRailwayLogs(rawLog)
	if len(errors) == 0 {
		log.Printf("✅ Auto-Healer: no errors found in logs")
		return nil
	}

	log.Printf("🩺 Auto-Healer: %d errors found, generating fix commands", len(errors))
	var commands []HealCommand

	for _, err := range errors {
		switch err.Type {
		case "compile":
			commands = append(commands, HealCommand{
				Action:      "fix_code",
				Target:      extractFile(err.Message),
				Description: fmt.Sprintf("Compile error: %s", err.Message),
				Priority:    1,
			})

		case "runtime":
			commands = append(commands, HealCommand{
				Action:      "fix_code",
				Target:      extractFile(err.Message),
				Description: fmt.Sprintf("Runtime panic: %s", err.Message),
				Priority:    1,
			})

		case "timeout":
			commands = append(commands, HealCommand{
				Action:      "increase_timeout",
				Target:      "orchestrator",
				Description: fmt.Sprintf("Timeout: %s — increase deadline or optimize LLM call", err.Message),
				Priority:    2,
			})

		case "oom":
			commands = append(commands, HealCommand{
				Action:      "fix_code",
				Target:      "memory",
				Description: "OOM — reduce max_tokens or batch processing size",
				Priority:    1,
			})

		case "deploy":
			commands = append(commands, HealCommand{
				Action:      "fix_env",
				Target:      "railway",
				Description: fmt.Sprintf("Deploy error: %s", err.Message),
				Priority:    1,
			})
		}
	}

	// For complex errors — ask LLM to diagnose
	if len(errors) > 0 && hasComplexError(errors) {
		llmDiagnosis := o.llmDiagnose(ctx, errors)
		if llmDiagnosis != "" {
			commands = append(commands, HealCommand{
				Action:      "fix_code",
				Target:      "auto",
				Description: llmDiagnosis,
				Priority:    1,
			})
		}
	}

	return commands
}

// llmDiagnose asks ValidatorAgent to analyze complex errors.
func (o *Orchestrator) llmDiagnose(ctx context.Context, errors []BuildError) string {
	var errSummary strings.Builder
	for _, e := range errors {
		errSummary.WriteString(fmt.Sprintf("[%s] %s\n%s\n---\n", e.Type, e.Message, e.Raw))
	}

	agent := o.agents[RoleValidator]
	if agent == nil {
		return ""
	}

	prompt := fmt.Sprintf(`Analyze these Railway build/runtime errors and provide a concise fix command.

ERRORS:
%s

Respond with:
1. Root cause (one sentence)
2. Fix command (specific file + change)
3. Prevention (one sentence)`, errSummary.String())

	result, err := o.callLLM(ctx, agent.Model,
		"You are a build error diagnostician. Be precise and actionable.",
		prompt, 1000)
	if err != nil {
		log.Printf("⚠️ Auto-Healer LLM diagnosis failed: %v", err)
		return ""
	}

	return strings.TrimSpace(result)
}

// ── Helpers ──────────────────────────────────────────────

func contextLines(lines []string, idx, radius int) string {
	start := idx - radius
	if start < 0 {
		start = 0
	}
	end := idx + radius + 1
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start:end], "\n")
}

func extractFile(msg string) string {
	// Extract "file.go" from "file.go:42:10: error..."
	parts := strings.SplitN(msg, ":", 2)
	if len(parts) > 0 {
		f := strings.TrimSpace(parts[0])
		if strings.Contains(f, ".go") || strings.Contains(f, ".ts") || strings.Contains(f, ".js") {
			return f
		}
	}
	return "unknown"
}

func hasComplexError(errors []BuildError) bool {
	for _, e := range errors {
		if e.Type == "runtime" || e.Type == "oom" {
			return true
		}
	}
	return false
}

func dedup(errors []BuildError) []BuildError {
	seen := make(map[string]bool)
	var result []BuildError
	for _, e := range errors {
		key := e.Type + "|" + e.Message
		if !seen[key] {
			seen[key] = true
			result = append(result, e)
		}
	}
	return result
}
