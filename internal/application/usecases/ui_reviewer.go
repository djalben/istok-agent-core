package usecases

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — UI/UX Reviewer Agent (Layer 3)
//  Premium Minimal style audit
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// UIReviewReport — итоговый отчёт UI/UX-ревью.
type UIReviewReport struct {
	Approved bool              `json:"approved"`
	Issues   []ValidationIssue `json:"issues"`
	Summary  string            `json:"summary"`
	FixHint  string            `json:"fix_hint,omitempty"`
}

// CriticalCount возвращает кол-во critical issues.
func (r *UIReviewReport) CriticalCount() int {
	n := 0
	for _, iss := range r.Issues {
		if iss.Severity == SeverityCritical {
			n++
		}
	}
	return n
}

// ForCoderContext генерирует лог для контекста Кодера при retry.
func (r *UIReviewReport) ForCoderContext() string {
	if r.Approved || len(r.Issues) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## UI/UX REVIEW FAILURES (Premium Minimal style)\n\n")
	for i, iss := range r.Issues {
		b.WriteString(fmt.Sprintf("%d. [%s][%s] %s — %s\n", i+1, iss.Severity, iss.Category, iss.File, iss.Message))
		if iss.Snippet != "" {
			b.WriteString("   snippet: " + iss.Snippet + "\n")
		}
	}
	if r.FixHint != "" {
		b.WriteString("\nFIX HINT: " + r.FixHint + "\n")
	}
	return b.String()
}

// ────────────────────────────────────────────────────
//  ReviewUIUX — entry point
// ────────────────────────────────────────────────────

// ReviewUIUX проверяет код на соответствие Premium Minimal стилю:
// - излишек border-* классов (Premium Minimal избегает рамок)
// - низкий контраст цветовых пар
// - корректное использование shadcn/ui (импорты + структура)
// - hardcoded hex без CSS-переменных
func ReviewUIUX(files map[string]string) *UIReviewReport {
	report := &UIReviewReport{}

	for filename, content := range files {
		// Только UI-файлы: tsx, jsx, html, css
		if !isUIFile(filename) {
			continue
		}

		report.Issues = append(report.Issues, checkExcessiveBorders(filename, content)...)
		report.Issues = append(report.Issues, checkContrast(filename, content)...)
		report.Issues = append(report.Issues, checkShadcnUsage(filename, content)...)
		report.Issues = append(report.Issues, checkHardcodedColors(filename, content)...)
		report.Issues = append(report.Issues, checkPremiumMinimalAntipatterns(filename, content)...)
	}

	criticals := report.CriticalCount()
	warnings := 0
	for _, iss := range report.Issues {
		if iss.Severity == SeverityWarning {
			warnings++
		}
	}

	if criticals > 0 {
		report.Approved = false
		report.Summary = fmt.Sprintf("REJECTED (Premium Minimal): %d critical, %d warnings", criticals, warnings)
		var hints []string
		for _, iss := range report.Issues {
			if iss.Severity == SeverityCritical {
				hints = append(hints, iss.File+": "+iss.Message)
			}
		}
		report.FixHint = "Fix Premium Minimal violations: " + strings.Join(hints, "; ")
	} else {
		report.Approved = true
		report.Summary = fmt.Sprintf("APPROVED (Premium Minimal): 0 critical, %d warnings", warnings)
	}

	return report
}

// ────────────────────────────────────────────────────
//  Checks
// ────────────────────────────────────────────────────

// checkExcessiveBorders — Premium Minimal избегает явных рамок.
// Допустимо: border-b на разделителях, border на input/card.
// Превышение порога → critical.
func checkExcessiveBorders(filename, content string) []ValidationIssue {
	var issues []ValidationIssue

	borderRe := regexp.MustCompile(`\bborder(?:-(?:t|r|b|l|x|y|2|4|8|solid|dashed|dotted|\[\d+px\]))?\b`)
	matches := borderRe.FindAllStringIndex(content, -1)
	count := len(matches)

	// Subtract acceptable usages (border-input, border-border which are shadcn tokens)
	acceptable := strings.Count(content, "border-input") +
		strings.Count(content, "border-border") +
		strings.Count(content, "border-transparent") +
		strings.Count(content, "border-collapse") +
		strings.Count(content, "border-spacing") +
		strings.Count(content, "border-none")

	excess := count - acceptable

	switch {
	case excess > 15:
		issues = append(issues, ValidationIssue{
			Severity: SeverityCritical,
			Category: "ui_ux",
			File:     filename,
			Message:  fmt.Sprintf("Excessive borders (%d in file) — Premium Minimal prefers whitespace + subtle shadows over borders", excess),
		})
	case excess > 8:
		issues = append(issues, ValidationIssue{
			Severity: SeverityWarning,
			Category: "ui_ux",
			File:     filename,
			Message:  fmt.Sprintf("High border usage (%d) — review for Premium Minimal compliance", excess),
		})
	}

	// Heavy borders (>2px) anywhere
	heavyBorderRe := regexp.MustCompile(`\bborder-(?:4|8|\[(\d+)px\])`)
	for _, m := range heavyBorderRe.FindAllStringSubmatchIndex(content, -1) {
		full := content[m[0]:m[1]]
		// Extract pixel value if [Npx]
		px := 0
		if m[2] != -1 && m[3] != -1 {
			px, _ = strconv.Atoi(content[m[2]:m[3]])
		}
		if strings.Contains(full, "border-8") || px >= 4 {
			issues = append(issues, ValidationIssue{
				Severity: SeverityWarning,
				Category: "ui_ux",
				File:     filename,
				Line:     lineAt(content, m[0]),
				Message:  "Heavy border (>2px) — Premium Minimal uses 1px or none",
				Snippet:  safeSnippet(content, m[0], 40),
			})
		}
	}

	return issues
}

// checkContrast — выявляет потенциально проблемные пары fg/bg.
// Heuristic: ищем text-{color}-{shade} рядом с bg-{color}-{shade} в той же строке.
func checkContrast(filename, content string) []ValidationIssue {
	var issues []ValidationIssue

	// Грубая проверка: text-gray-400 на bg-white = низкий контраст
	lowContrastRe := regexp.MustCompile(`(?:text-(?:gray|slate|zinc|neutral|stone)-(?:300|400)).*?bg-(?:white|gray-50|slate-50)`)
	for _, loc := range lowContrastRe.FindAllStringIndex(content, -1) {
		issues = append(issues, ValidationIssue{
			Severity: SeverityWarning,
			Category: "ui_ux",
			File:     filename,
			Line:     lineAt(content, loc[0]),
			Message:  "Potential low contrast: light gray text on white background — WCAG AA requires 4.5:1 for body text",
			Snippet:  safeSnippet(content, loc[0], 80),
		})
	}

	// Проверка hex-пар на низкий контраст (если оба указаны рядом)
	hexRe := regexp.MustCompile(`#[0-9a-fA-F]{6}\b`)
	hexes := hexRe.FindAllStringIndex(content, -1)
	if len(hexes) >= 2 {
		// Берём последовательные пары и считаем контраст
		for i := 0; i < len(hexes)-1; i++ {
			h1 := content[hexes[i][0]:hexes[i][1]]
			h2 := content[hexes[i+1][0]:hexes[i+1][1]]
			// Только если на одной строке
			if lineAt(content, hexes[i][0]) != lineAt(content, hexes[i+1][0]) {
				continue
			}
			ratio := contrastRatio(h1, h2)
			if ratio > 0 && ratio < 3.0 {
				issues = append(issues, ValidationIssue{
					Severity: SeverityWarning,
					Category: "ui_ux",
					File:     filename,
					Line:     lineAt(content, hexes[i][0]),
					Message:  fmt.Sprintf("Low color contrast %s vs %s (ratio %.2f:1, need ≥4.5:1 for body text)", h1, h2, ratio),
				})
			}
		}
	}

	return issues
}

// checkShadcnUsage — проверяет правильность использования shadcn/ui.
func checkShadcnUsage(filename, content string) []ValidationIssue {
	var issues []ValidationIssue

	// Только TSX/JSX
	if !strings.HasSuffix(filename, ".tsx") && !strings.HasSuffix(filename, ".jsx") {
		return issues
	}

	// Импорты shadcn должны идти из @/components/ui/*
	wrongShadcnRe := regexp.MustCompile(`from\s+["']\.\./.+?/components/ui/(\w+)["']`)
	for _, m := range wrongShadcnRe.FindAllStringIndex(content, -1) {
		issues = append(issues, ValidationIssue{
			Severity: SeverityCritical,
			Category: "ui_ux",
			File:     filename,
			Line:     lineAt(content, m[0]),
			Message:  "shadcn import uses relative path — must use @/components/ui/* alias",
			Snippet:  safeSnippet(content, m[0], 80),
		})
	}

	// Native HTML elements where shadcn alternatives exist
	nativeAntipatterns := []struct {
		pattern *regexp.Regexp
		shadcn  string
	}{
		{regexp.MustCompile(`<button\s+(?:[^>]*\s)?className=["'][^"']*(?:bg-|px-|py-|rounded)`), "Button"},
		{regexp.MustCompile(`<input\s+(?:[^>]*\s)?className=["'][^"']*(?:border|rounded|px-)`), "Input"},
		{regexp.MustCompile(`<select\s+(?:[^>]*\s)?className=["']`), "Select"},
		{regexp.MustCompile(`<dialog\s`), "Dialog"},
	}
	for _, ap := range nativeAntipatterns {
		if loc := ap.pattern.FindStringIndex(content); loc != nil {
			// Skip if shadcn equivalent is imported
			importCheck := fmt.Sprintf(`@/components/ui/%s`, strings.ToLower(ap.shadcn))
			if strings.Contains(content, importCheck) {
				continue
			}
			issues = append(issues, ValidationIssue{
				Severity: SeverityWarning,
				Category: "ui_ux",
				File:     filename,
				Line:     lineAt(content, loc[0]),
				Message:  fmt.Sprintf("Native HTML element styled inline — use shadcn <%s> from @/components/ui/%s instead", ap.shadcn, strings.ToLower(ap.shadcn)),
				Snippet:  safeSnippet(content, loc[0], 60),
			})
		}
	}

	// cn() utility from @/lib/utils should be used for class merging
	if strings.Contains(content, "className={`") || strings.Contains(content, `className={"`) {
		if !strings.Contains(content, "import { cn }") && !strings.Contains(content, "from \"@/lib/utils\"") {
			issues = append(issues, ValidationIssue{
				Severity: SeverityInfo,
				Category: "ui_ux",
				File:     filename,
				Message:  "Template-literal className without cn() utility — import { cn } from \"@/lib/utils\" for proper class merging",
			})
		}
	}

	return issues
}

// checkHardcodedColors — Premium Minimal требует CSS-переменные / Tailwind tokens.
func checkHardcodedColors(filename, content string) []ValidationIssue {
	var issues []ValidationIssue

	// Только TSX/CSS файлы
	if !strings.HasSuffix(filename, ".tsx") &&
		!strings.HasSuffix(filename, ".jsx") &&
		!strings.HasSuffix(filename, ".css") {
		return issues
	}

	// Hardcoded hex в TSX className
	hardcodedInClassRe := regexp.MustCompile(`className=["'][^"']*\[#[0-9a-fA-F]{6}\][^"']*["']`)
	matches := hardcodedInClassRe.FindAllStringIndex(content, -1)
	if len(matches) > 3 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityWarning,
			Category: "ui_ux",
			File:     filename,
			Message:  fmt.Sprintf("Excessive arbitrary hex values (%d) in className — use CSS variables / theme tokens", len(matches)),
		})
	}

	// rgba(...) inline в className (антипаттерн)
	rgbaInClassRe := regexp.MustCompile(`className=["'][^"']*rgba?\(`)
	if loc := rgbaInClassRe.FindStringIndex(content); loc != nil {
		issues = append(issues, ValidationIssue{
			Severity: SeverityWarning,
			Category: "ui_ux",
			File:     filename,
			Line:     lineAt(content, loc[0]),
			Message:  "Inline rgba() in className — use Tailwind opacity modifiers (bg-foreground/50)",
			Snippet:  safeSnippet(content, loc[0], 80),
		})
	}

	return issues
}

// checkPremiumMinimalAntipatterns — глобальные антипаттерны стиля.
func checkPremiumMinimalAntipatterns(filename, content string) []ValidationIssue {
	var issues []ValidationIssue

	// Premium Minimal: subtle shadows allowed, gradient backgrounds rare
	gradientRe := regexp.MustCompile(`bg-gradient-to-(?:t|r|b|l|tr|tl|br|bl)`)
	gradientCount := len(gradientRe.FindAllStringIndex(content, -1))
	if gradientCount > 3 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityWarning,
			Category: "ui_ux",
			File:     filename,
			Message:  fmt.Sprintf("Excessive gradients (%d) — Premium Minimal prefers solid colors with subtle accents", gradientCount),
		})
	}

	// Heavy shadows (shadow-2xl) overuse
	heavyShadowRe := regexp.MustCompile(`shadow-2xl`)
	if c := len(heavyShadowRe.FindAllStringIndex(content, -1)); c > 3 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityWarning,
			Category: "ui_ux",
			File:     filename,
			Message:  fmt.Sprintf("Heavy shadow-2xl used %d times — Premium Minimal uses shadow-sm/shadow-md", c),
		})
	}

	// Text size escalation: too many text-{size} variations in one file
	textSizes := map[string]bool{}
	textSizeRe := regexp.MustCompile(`\btext-(?:xs|sm|base|lg|xl|2xl|3xl|4xl|5xl|6xl|7xl|8xl|9xl)\b`)
	for _, m := range textSizeRe.FindAllString(content, -1) {
		textSizes[m] = true
	}
	if len(textSizes) > 6 {
		issues = append(issues, ValidationIssue{
			Severity: SeverityWarning,
			Category: "ui_ux",
			File:     filename,
			Message:  fmt.Sprintf("Too many text-size variations (%d) — Premium Minimal uses tight typography scale (3-5 sizes max)", len(textSizes)),
		})
	}

	return issues
}

// ────────────────────────────────────────────────────
//  Helpers
// ────────────────────────────────────────────────────

func isUIFile(filename string) bool {
	for _, ext := range []string{".tsx", ".jsx", ".html", ".htm", ".css", ".vue", ".svelte"} {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// contrastRatio — WCAG contrast ratio between two #RRGGBB hex strings.
// Returns 0 if either is invalid.
func contrastRatio(hex1, hex2 string) float64 {
	l1, ok1 := relativeLuminance(hex1)
	l2, ok2 := relativeLuminance(hex2)
	if !ok1 || !ok2 {
		return 0
	}
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// relativeLuminance computes WCAG relative luminance for a #RRGGBB hex.
func relativeLuminance(hex string) (float64, bool) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, false
	}
	rInt, err := strconv.ParseInt(hex[0:2], 16, 32)
	if err != nil {
		return 0, false
	}
	gInt, err := strconv.ParseInt(hex[2:4], 16, 32)
	if err != nil {
		return 0, false
	}
	bInt, err := strconv.ParseInt(hex[4:6], 16, 32)
	if err != nil {
		return 0, false
	}
	r := channelToLinear(float64(rInt) / 255.0)
	g := channelToLinear(float64(gInt) / 255.0)
	b := channelToLinear(float64(bInt) / 255.0)
	return 0.2126*r + 0.7152*g + 0.0722*b, true
}

func channelToLinear(c float64) float64 {
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}
