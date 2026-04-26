package usecases

import (
	"fmt"
	"regexp"
	"strings"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Verification Layer (Layer 3)
//  Quality Gate + Security Agent + Auto-Fix
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Severity уровень серьёзности найденной проблемы
type Severity string

const (
	SeverityCritical Severity = "critical" // блокирует деплой, требует auto-fix
	SeverityWarning  Severity = "warning"  // логируется, не блокирует
	SeverityInfo     Severity = "info"     // информационное
)

// ValidationIssue — одна найденная проблема в коде
type ValidationIssue struct {
	Severity Severity `json:"severity"`
	Category string   `json:"category"` // "quality" | "security" | "syntax"
	File     string   `json:"file"`
	Line     int      `json:"line,omitempty"`
	Message  string   `json:"message"`
	Snippet  string   `json:"snippet,omitempty"`
}

// ValidationResult — полный результат валидации всех файлов
type ValidationResult struct {
	Passed  bool              `json:"passed"`
	Issues  []ValidationIssue `json:"issues"`
	Summary string            `json:"summary"`
	FixHint string            `json:"fix_hint,omitempty"` // промпт для Coder при auto-fix
}

// CriticalCount возвращает количество critical issues
func (r *ValidationResult) CriticalCount() int {
	n := 0
	for _, iss := range r.Issues {
		if iss.Severity == SeverityCritical {
			n++
		}
	}
	return n
}

// ForCoderContext генерирует компактный лог ошибок для вставки в промпт Кодера при retry
func (r *ValidationResult) ForCoderContext() string {
	if r.Passed || len(r.Issues) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## VALIDATION ERRORS (fix ALL before returning code)\n\n")
	for i, iss := range r.Issues {
		b.WriteString(fmt.Sprintf("%d. [%s][%s] %s — %s\n",
			i+1, iss.Severity, iss.Category, iss.File, iss.Message))
		if iss.Snippet != "" {
			b.WriteString(fmt.Sprintf("   snippet: %s\n", iss.Snippet))
		}
	}
	if r.FixHint != "" {
		b.WriteString(fmt.Sprintf("\nFIX HINT: %s\n", r.FixHint))
	}
	return b.String()
}

// ────────────────────────────────────────────────────
//  Quality Gate — проверка качества кода
// ────────────────────────────────────────────────────

// QualityGate проверяет сгенерированный код на качество:
// - синтаксические ошибки (HTML structure, unclosed tags)
// - "Lorem Ipsum" заглушки
// - пустые функции / компоненты-заглушки
// - минимальный размер файлов
func QualityGate(files map[string]string) []ValidationIssue {
	var issues []ValidationIssue

	for filename, content := range files {
		lower := strings.ToLower(content)

		// ── Lorem Ipsum detection ──
		if strings.Contains(lower, "lorem ipsum") || strings.Contains(lower, "dolor sit amet") {
			line := findLineNumber(content, "lorem ipsum", "Lorem ipsum", "dolor sit amet")
			issues = append(issues, ValidationIssue{
				Severity: SeverityCritical,
				Category: "quality",
				File:     filename,
				Line:     line,
				Message:  "Contains Lorem Ipsum placeholder text — must use real content",
			})
		}

		// ── Empty functions / stub components ──
		emptyFuncPatterns := []struct {
			pattern *regexp.Regexp
			desc    string
		}{
			{regexp.MustCompile(`function\s+\w+\s*\([^)]*\)\s*\{\s*\}`), "empty function body"},
			{regexp.MustCompile(`=>\s*\{\s*\}`), "empty arrow function body"},
			{regexp.MustCompile(`(?i)//\s*TODO`), "TODO comment left in production code"},
			{regexp.MustCompile(`(?i)//\s*FIXME`), "FIXME comment left in production code"},
			{regexp.MustCompile(`(?i)//\s*HACK`), "HACK comment left in production code"},
			{regexp.MustCompile(`return\s+null\s*;?\s*//`), "component returns null with comment (stub)"},
		}
		for _, pat := range emptyFuncPatterns {
			if loc := pat.pattern.FindStringIndex(content); loc != nil {
				snippet := safeSnippet(content, loc[0], 60)
				issues = append(issues, ValidationIssue{
					Severity: SeverityWarning,
					Category: "quality",
					File:     filename,
					Message:  pat.desc,
					Snippet:  snippet,
				})
			}
		}

		// ── HTML structure checks (for .html files) ──
		if strings.HasSuffix(filename, ".html") || strings.HasSuffix(filename, ".htm") {
			issues = append(issues, checkHTMLStructure(filename, content)...)
		}

		// ── TypeScript/JSX checks ──
		if strings.HasSuffix(filename, ".tsx") || strings.HasSuffix(filename, ".ts") ||
			strings.HasSuffix(filename, ".jsx") || strings.HasSuffix(filename, ".js") {
			issues = append(issues, checkTSStructure(filename, content)...)
		}

		// ── Minimum file size ──
		if len(content) < 50 && len(content) > 0 {
			issues = append(issues, ValidationIssue{
				Severity: SeverityWarning,
				Category: "quality",
				File:     filename,
				Message:  fmt.Sprintf("File suspiciously small (%d bytes)", len(content)),
			})
		}
	}

	return issues
}

// checkHTMLStructure validates HTML file structure
func checkHTMLStructure(filename, content string) []ValidationIssue {
	var issues []ValidationIssue
	lower := strings.ToLower(content)

	if !strings.Contains(lower, "<!doctype") {
		issues = append(issues, ValidationIssue{
			Severity: SeverityCritical,
			Category: "syntax",
			File:     filename,
			Message:  "Missing <!DOCTYPE html> declaration",
		})
	}
	if !strings.Contains(lower, "<html") {
		issues = append(issues, ValidationIssue{
			Severity: SeverityCritical,
			Category: "syntax",
			File:     filename,
			Message:  "Missing <html> root element",
		})
	}
	if !strings.Contains(lower, "<head") {
		issues = append(issues, ValidationIssue{
			Severity: SeverityCritical,
			Category: "syntax",
			File:     filename,
			Message:  "Missing <head> section",
		})
	}
	if !strings.Contains(lower, "<body") {
		issues = append(issues, ValidationIssue{
			Severity: SeverityCritical,
			Category: "syntax",
			File:     filename,
			Message:  "Missing <body> section",
		})
	}
	if !strings.Contains(lower, "</html>") {
		issues = append(issues, ValidationIssue{
			Severity: SeverityCritical,
			Category: "syntax",
			File:     filename,
			Message:  "Missing closing </html> tag",
		})
	}

	// Unclosed script/style tags
	if strings.Count(lower, "<script") != strings.Count(lower, "</script>") {
		issues = append(issues, ValidationIssue{
			Severity: SeverityCritical,
			Category: "syntax",
			File:     filename,
			Message:  "Unclosed <script> tag",
		})
	}
	if strings.Count(lower, "<style") != strings.Count(lower, "</style>") {
		issues = append(issues, ValidationIssue{
			Severity: SeverityCritical,
			Category: "syntax",
			File:     filename,
			Message:  "Unclosed <style> tag",
		})
	}

	// JS runtime risk
	if strings.Contains(content, "document.getElementById") &&
		!strings.Contains(content, "DOMContentLoaded") &&
		!strings.Contains(content, "defer") {
		issues = append(issues, ValidationIssue{
			Severity: SeverityWarning,
			Category: "syntax",
			File:     filename,
			Message:  "JS uses getElementById without DOMContentLoaded or defer — runtime risk",
		})
	}

	return issues
}

// checkTSStructure validates TypeScript/JavaScript files
func checkTSStructure(filename, content string) []ValidationIssue {
	var issues []ValidationIssue

	// Check for relative imports (should use @/* aliases)
	relativeImportRe := regexp.MustCompile(`from\s+['"]\.\.?/`)
	if locs := relativeImportRe.FindAllStringIndex(content, -1); len(locs) > 3 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityWarning,
			Category: "quality",
			File:     filename,
			Message:  fmt.Sprintf("Uses %d relative imports — should use @/* aliases", len(locs)),
		})
	}

	// Check for 'any' type abuse
	anyTypeRe := regexp.MustCompile(`:\s*any\b`)
	if matches := anyTypeRe.FindAllString(content, -1); len(matches) > 3 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityWarning,
			Category: "quality",
			File:     filename,
			Message:  fmt.Sprintf("Excessive 'any' type usage (%d occurrences) — use proper types", len(matches)),
		})
	}

	// Check for console.log in production code
	if strings.Count(content, "console.log") > 5 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityInfo,
			Category: "quality",
			File:     filename,
			Message:  fmt.Sprintf("Excessive console.log (%d) — clean up for production", strings.Count(content, "console.log")),
		})
	}

	return issues
}

// ────────────────────────────────────────────────────
//  Security Agent — проверка безопасности
// ────────────────────────────────────────────────────

// SecurityAgent проверяет код на уязвимости:
// - eval() вызовы
// - dangerouslySetInnerHTML без санитизации
// - inline <script> без nonce
// - жёстко зашитые секреты/токены
func SecurityAgent(files map[string]string) []ValidationIssue {
	var issues []ValidationIssue

	for filename, content := range files {
		// ── eval() detection ──
		evalRe := regexp.MustCompile(`\beval\s*\(`)
		if locs := evalRe.FindAllStringIndex(content, -1); locs != nil {
			for _, loc := range locs {
				issues = append(issues, ValidationIssue{
					Severity: SeverityCritical,
					Category: "security",
					File:     filename,
					Line:     lineAt(content, loc[0]),
					Message:  "eval() call detected — XSS risk, use safer alternatives (JSON.parse, Function constructor with validation)",
					Snippet:  safeSnippet(content, loc[0], 60),
				})
			}
		}

		// ── new Function() detection ──
		newFuncRe := regexp.MustCompile(`new\s+Function\s*\(`)
		if locs := newFuncRe.FindAllStringIndex(content, -1); locs != nil {
			for _, loc := range locs {
				issues = append(issues, ValidationIssue{
					Severity: SeverityCritical,
					Category: "security",
					File:     filename,
					Line:     lineAt(content, loc[0]),
					Message:  "new Function() detected — equivalent to eval(), XSS risk",
					Snippet:  safeSnippet(content, loc[0], 60),
				})
			}
		}

		// ── dangerouslySetInnerHTML detection ──
		if strings.Contains(content, "dangerouslySetInnerHTML") {
			// Check if DOMPurify/sanitize is imported
			hasSanitizer := strings.Contains(content, "DOMPurify") ||
				strings.Contains(content, "sanitize") ||
				strings.Contains(content, "dompurify") ||
				strings.Contains(content, "xss")
			if !hasSanitizer {
				loc := strings.Index(content, "dangerouslySetInnerHTML")
				issues = append(issues, ValidationIssue{
					Severity: SeverityCritical,
					Category: "security",
					File:     filename,
					Line:     lineAt(content, loc),
					Message:  "dangerouslySetInnerHTML used without DOMPurify/sanitizer — XSS vulnerability",
					Snippet:  safeSnippet(content, loc, 80),
				})
			}
		}

		// ── Inline <script> without nonce ──
		if strings.HasSuffix(filename, ".html") || strings.HasSuffix(filename, ".htm") {
			scriptRe := regexp.MustCompile(`<script(?:\s[^>]*)?>`)
			for _, match := range scriptRe.FindAllStringSubmatchIndex(content, -1) {
				tag := content[match[0]:match[1]]
				// Skip external scripts (src=...) and scripts with nonce
				if strings.Contains(tag, "src=") {
					continue
				}
				if !strings.Contains(tag, "nonce=") {
					issues = append(issues, ValidationIssue{
						Severity: SeverityWarning,
						Category: "security",
						File:     filename,
						Line:     lineAt(content, match[0]),
						Message:  "Inline <script> without nonce attribute — CSP bypass risk",
						Snippet:  safeSnippet(content, match[0], 60),
					})
				}
			}
		}

		// ── Hardcoded secrets/tokens ──
		secretPatterns := []struct {
			re   *regexp.Regexp
			desc string
		}{
			{regexp.MustCompile(`(?i)(api[_-]?key|secret[_-]?key|password|token)\s*[:=]\s*['"][^'"]{8,}['"]`), "Hardcoded secret/API key"},
			{regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`), "OpenAI API key pattern"},
			{regexp.MustCompile(`r8_[a-zA-Z0-9]{20,}`), "Replicate API token pattern"},
			{regexp.MustCompile(`ghp_[a-zA-Z0-9]{20,}`), "GitHub personal access token"},
		}
		for _, pat := range secretPatterns {
			if loc := pat.re.FindStringIndex(content); loc != nil {
				issues = append(issues, ValidationIssue{
					Severity: SeverityCritical,
					Category: "security",
					File:     filename,
					Line:     lineAt(content, loc[0]),
					Message:  pat.desc + " detected — use environment variables",
					Snippet:  maskSecret(safeSnippet(content, loc[0], 40)),
				})
			}
		}

		// ── innerHTML without sanitization (non-React) ──
		if !strings.Contains(content, "dangerouslySetInnerHTML") {
			innerHTMLRe := regexp.MustCompile(`\.innerHTML\s*=`)
			if locs := innerHTMLRe.FindAllStringIndex(content, -1); locs != nil {
				hasSanitizer := strings.Contains(content, "DOMPurify") || strings.Contains(content, "sanitize")
				if !hasSanitizer {
					for _, loc := range locs {
						issues = append(issues, ValidationIssue{
							Severity: SeverityWarning,
							Category: "security",
							File:     filename,
							Line:     lineAt(content, loc[0]),
							Message:  ".innerHTML assignment without sanitizer — potential XSS",
							Snippet:  safeSnippet(content, loc[0], 60),
						})
					}
				}
			}
		}

		// ── CSP-specific checks (HTML files) ──
		if strings.HasSuffix(filename, ".html") || strings.HasSuffix(filename, ".htm") {
			lower := strings.ToLower(content)

			// Missing CSP meta tag
			hasCSPMeta := strings.Contains(lower, "content-security-policy")
			hasCSPHeader := strings.Contains(lower, `http-equiv="content-security-policy"`) ||
				strings.Contains(lower, `http-equiv='content-security-policy'`)
			if !hasCSPMeta && !hasCSPHeader {
				issues = append(issues, ValidationIssue{
					Severity: SeverityWarning,
					Category: "security",
					File:     filename,
					Message:  "No Content-Security-Policy meta tag found — add <meta http-equiv=\"Content-Security-Policy\" content=\"...\">",
				})
			}

			// CSP with unsafe-eval or unsafe-inline
			if strings.Contains(lower, "unsafe-eval") {
				loc := strings.Index(lower, "unsafe-eval")
				issues = append(issues, ValidationIssue{
					Severity: SeverityCritical,
					Category: "security",
					File:     filename,
					Line:     lineAt(content, loc),
					Message:  "CSP contains 'unsafe-eval' — defeats CSP protection, remove it",
					Snippet:  safeSnippet(content, loc, 60),
				})
			}
			if strings.Contains(lower, "unsafe-inline") {
				loc := strings.Index(lower, "unsafe-inline")
				issues = append(issues, ValidationIssue{
					Severity: SeverityWarning,
					Category: "security",
					File:     filename,
					Line:     lineAt(content, loc),
					Message:  "CSP contains 'unsafe-inline' — use nonce-based or hash-based CSP instead",
					Snippet:  safeSnippet(content, loc, 60),
				})
			}

			// Inline event handlers (onclick=, onerror=, onload=, etc.)
			inlineHandlerRe := regexp.MustCompile(`(?i)\bon(click|error|load|mouseover|mouseout|focus|blur|change|submit|input|keydown|keyup|keypress)\s*=\s*["']`)
			for _, loc := range inlineHandlerRe.FindAllStringIndex(content, -1) {
				issues = append(issues, ValidationIssue{
					Severity: SeverityCritical,
					Category: "security",
					File:     filename,
					Line:     lineAt(content, loc[0]),
					Message:  "Inline event handler (onclick/onerror/etc.) violates CSP — use addEventListener",
					Snippet:  safeSnippet(content, loc[0], 60),
				})
			}

			// javascript: URLs
			jsURLRe := regexp.MustCompile(`(?i)href\s*=\s*["']\s*javascript:`)
			for _, loc := range jsURLRe.FindAllStringIndex(content, -1) {
				issues = append(issues, ValidationIssue{
					Severity: SeverityCritical,
					Category: "security",
					File:     filename,
					Line:     lineAt(content, loc[0]),
					Message:  "javascript: URL in href — XSS vector, use addEventListener",
					Snippet:  safeSnippet(content, loc[0], 60),
				})
			}

			// target="_blank" without rel="noopener"
			blankRe := regexp.MustCompile(`(?i)target\s*=\s*["']_blank["']`)
			for _, loc := range blankRe.FindAllStringIndex(content, -1) {
				// Look for rel attribute within ±200 chars of the match
				start := loc[0] - 200
				if start < 0 {
					start = 0
				}
				end := loc[1] + 200
				if end > len(content) {
					end = len(content)
				}
				surrounding := strings.ToLower(content[start:end])
				if !strings.Contains(surrounding, "noopener") {
					issues = append(issues, ValidationIssue{
						Severity: SeverityWarning,
						Category: "security",
						File:     filename,
						Line:     lineAt(content, loc[0]),
						Message:  `target="_blank" without rel="noopener noreferrer" — tabnabbing risk`,
						Snippet:  safeSnippet(content, loc[0], 60),
					})
				}
			}

			// Mixed content: http:// in https-served context (warn on http-only resources)
			if mixedRe := regexp.MustCompile(`(?i)(src|href)\s*=\s*["']http://`); mixedRe != nil {
				if loc := mixedRe.FindStringIndex(content); loc != nil {
					issues = append(issues, ValidationIssue{
						Severity: SeverityWarning,
						Category: "security",
						File:     filename,
						Line:     lineAt(content, loc[0]),
						Message:  "Insecure http:// resource — use https:// to avoid mixed content",
						Snippet:  safeSnippet(content, loc[0], 60),
					})
				}
			}
		}
	}

	return issues
}

// ────────────────────────────────────────────────────
//  ValidateCode — полная проверка (Quality + Security)
// ────────────────────────────────────────────────────

// ValidateCode запускает QualityGate + SecurityAgent и формирует итоговый ValidationResult.
// Если есть critical issues — Passed=false и генерируется FixHint для auto-fix.
func ValidateCode(files map[string]string) *ValidationResult {
	var allIssues []ValidationIssue

	allIssues = append(allIssues, QualityGate(files)...)
	allIssues = append(allIssues, SecurityAgent(files)...)

	result := &ValidationResult{
		Issues: allIssues,
	}

	criticals := result.CriticalCount()
	warnings := 0
	for _, iss := range allIssues {
		if iss.Severity == SeverityWarning {
			warnings++
		}
	}

	if criticals > 0 {
		result.Passed = false
		result.Summary = fmt.Sprintf("FAILED: %d critical, %d warnings across %d files",
			criticals, warnings, len(files))

		// Build fix hint for Coder
		var criticalDescs []string
		for _, iss := range allIssues {
			if iss.Severity == SeverityCritical {
				criticalDescs = append(criticalDescs, fmt.Sprintf("[%s] %s: %s", iss.Category, iss.File, iss.Message))
			}
		}
		result.FixHint = fmt.Sprintf("Fix these %d critical issues:\n%s",
			criticals, strings.Join(criticalDescs, "\n"))
	} else {
		result.Passed = true
		result.Summary = fmt.Sprintf("PASSED: 0 critical, %d warnings across %d files",
			warnings, len(files))
	}

	return result
}

// ────────────────────────────────────────────────────
//  Helpers
// ────────────────────────────────────────────────────

// findLineNumber finds the 1-indexed line number of the first occurrence of any pattern
func findLineNumber(content string, patterns ...string) int {
	idx := -1
	for _, p := range patterns {
		if i := strings.Index(strings.ToLower(content), strings.ToLower(p)); i != -1 {
			if idx == -1 || i < idx {
				idx = i
			}
		}
	}
	if idx < 0 {
		return 0
	}
	return lineAt(content, idx)
}

// lineAt returns 1-indexed line number for byte offset
func lineAt(content string, offset int) int {
	if offset < 0 || offset >= len(content) {
		return 0
	}
	return strings.Count(content[:offset], "\n") + 1
}

// safeSnippet extracts a snippet around offset, capped at maxLen
func safeSnippet(content string, offset, maxLen int) string {
	if offset < 0 {
		offset = 0
	}
	end := offset + maxLen
	if end > len(content) {
		end = len(content)
	}
	snippet := content[offset:end]
	snippet = strings.ReplaceAll(snippet, "\n", " ")
	snippet = strings.TrimSpace(snippet)
	return snippet
}

// maskSecret replaces middle of a secret string with ***
func maskSecret(s string) string {
	if len(s) < 12 {
		return s
	}
	return s[:6] + "***" + s[len(s)-3:]
}
