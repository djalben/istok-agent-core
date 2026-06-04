package usecases

import (
	"context"
	"fmt"

	"log/slog"
	"regexp"
	"strings"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Verification Gate (Layer 3 aggregator)
//  Объединяет Security + Tester + UI/UX Reviewer.
//  StateCompleted разрешён только если все 3 → Approved.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// AgentApproval — статус одобрения от конкретного агента.
type AgentApproval struct {
	Agent    string `json:"agent"` // "security" | "tester" | "ui_reviewer"
	Approved bool   `json:"approved"`
	Summary  string `json:"summary"`
	FixHint  string `json:"fix_hint,omitempty"`
}

// VerificationReport — итог Layer 3.
// Approved=true ТОЛЬКО если все три AgentApproval.Approved=true.
type VerificationReport struct {
	Approved      bool            `json:"approved"`
	Approvals     []AgentApproval `json:"approvals"`
	BlockingAgent string          `json:"blocking_agent,omitempty"` // первый агент, который заблокировал
	Summary       string          `json:"summary"`
	FixHint       string          `json:"fix_hint,omitempty"`
	TestsSkipped  bool            `json:"tests_skipped"` // true если Tester не запускался (RunTests=false) — верификация неполная

	// Детальные суб-отчёты (для дебага и UI)
	Validation *ValidationResult `json:"validation,omitempty"`
	Tests      *TesterReport     `json:"tests,omitempty"`
	UIUX       *UIReviewReport   `json:"uiux,omitempty"`
	Integrity  *IntegrityResult  `json:"integrity,omitempty"`
}

// ForCoderContext агрегирует ошибки от всех агентов для retry-промпта Кодера.
func (r *VerificationReport) ForCoderContext() string {
	if r.Approved {
		return ""
	}
	var b strings.Builder
	b.WriteString("# VERIFICATION GATE FAILED — fix all issues below\n\n")

	if r.Validation != nil && !r.Validation.Passed {
		b.WriteString(r.Validation.ForCoderContext())
		b.WriteString("\n")
	}
	if r.Tests != nil && !r.Tests.Approved {
		b.WriteString(r.Tests.ForCoderContext())
		b.WriteString("\n")
	}
	if r.UIUX != nil && !r.UIUX.Approved {
		b.WriteString(r.UIUX.ForCoderContext())
		b.WriteString("\n")
	}

	return b.String()
}

// ────────────────────────────────────────────────────
//  VerificationGate
// ────────────────────────────────────────────────────

// VerificationGate координирует Security + Tester + UI/UX Reviewer.
// Все три агента запускаются ВСЕГДА; результат блокирует переход в Completed,
// если хотя бы один не одобрил.
type VerificationGate struct {
	Tester     *TesterAgent
	RunTests   bool // если false, Tester помечается как Skipped и считается Approved
	StrictMode bool // если true, warnings от Security/UIUX тоже блокируют
}

// NewVerificationGate создаёт gate с разумными дефолтами.
// RunTests=true (запускаем тесты), StrictMode=false (только critical блокирует).
func NewVerificationGate() *VerificationGate {
	return &VerificationGate{
		Tester:     NewTesterAgent(),
		RunTests:   true,
		StrictMode: false,
	}
}

// Verify прогоняет все три агента и возвращает агрегированный отчёт.
// ВАЖНО: вызывает все три, не короткозамыкая на первой ошибке —
// Кодер должен получить полный список проблем для retry за один раз.
func (g *VerificationGate) Verify(ctx context.Context, files map[string]string) *VerificationReport {
	if len(files) == 0 {
		return &VerificationReport{
			Approved: false,
			Summary:  "no files to verify",
		}
	}

	report := &VerificationReport{}

	// ── 1. Security Agent (Quality Gate + Security checks) ──
	validation := ValidateCode(files)
	report.Validation = validation
	secApproval := AgentApproval{
		Agent:    "security",
		Approved: validation.Passed,
		Summary:  validation.Summary,
		FixHint:  validation.FixHint,
	}
	report.Approvals = append(report.Approvals, secApproval)
	slog.Info(fmt.Sprintf("🛡️ VerificationGate[security]: approved=%v %s", secApproval.Approved, secApproval.Summary))

	// ── 2. Tester Agent ──
	if g.RunTests && g.Tester != nil {
		testReport := g.Tester.RunTests(ctx, files)
		report.Tests = testReport
		report.Approvals = append(report.Approvals, AgentApproval{
			Agent:    "tester",
			Approved: testReport.Approved,
			Summary:  testReport.Summary,
			FixHint:  testReport.FixHint,
		})
		slog.Info(fmt.Sprintf("🧪 VerificationGate[tester]: approved=%v %s", testReport.Approved, testReport.Summary))
	} else {
		report.TestsSkipped = true
		report.Approvals = append(report.Approvals, AgentApproval{
			Agent:    "tester",
			Approved: true,
			Summary:  "tests skipped (RunTests=false)",
		})
		slog.Info(fmt.Sprintf("🧪 VerificationGate[tester]: SKIPPED"))
	}

	// ── 3. UI/UX Reviewer ──
	uiuxReport := ReviewUIUX(files)
	report.UIUX = uiuxReport
	report.Approvals = append(report.Approvals, AgentApproval{
		Agent:    "ui_reviewer",
		Approved: uiuxReport.Approved,
		Summary:  uiuxReport.Summary,
		FixHint:  uiuxReport.FixHint,
	})
	slog.Info(fmt.Sprintf("🎨 VerificationGate[ui_reviewer]: approved=%v %s", uiuxReport.Approved, uiuxReport.Summary))

	// ── 4. Cross-File Integrity (informational, non-blocking) ──
	integrity := CheckCrossFileIntegrity(files)
	report.Integrity = integrity
	if integrity.TotalImports > 0 {
		slog.Info(fmt.Sprintf("🔗 VerificationGate[integrity]: %d/%d imports resolved, %d missing",
			integrity.ResolvedCount, integrity.TotalImports, len(integrity.MissingFiles)))
	}

	// ── Aggregate: ВСЕ три должны быть Approved ──
	allApproved := true
	for _, a := range report.Approvals {
		if !a.Approved {
			allApproved = false
			if report.BlockingAgent == "" {
				report.BlockingAgent = a.Agent
			}
		}
	}

	report.Approved = allApproved
	if allApproved {
		report.Summary = fmt.Sprintf("✅ VerificationGate APPROVED: %d/%d agents passed", len(report.Approvals), len(report.Approvals))
	} else {
		passed := 0
		for _, a := range report.Approvals {
			if a.Approved {
				passed++
			}
		}
		report.Summary = fmt.Sprintf("❌ VerificationGate BLOCKED by [%s]: %d/%d agents passed",
			report.BlockingAgent, passed, len(report.Approvals))
		report.FixHint = fmt.Sprintf("Layer 3 rejected: blocking agent=%s. See per-agent FixHints in approvals.",
			report.BlockingAgent)
	}

	return report
}

// CanTransitionToCompleted — guard для FSM.
// Возвращает nil только если Verify дал Approved=true.
// Используется оркестратором перед переводом в StateCompleted.
func (g *VerificationGate) CanTransitionToCompleted(report *VerificationReport) error {
	if report == nil {
		return ErrVerificationGateNoReport
	}
	if !report.Approved {
		return fmt.Errorf("%w: [%s] %s",
			ErrVerificationGateBlocked, report.BlockingAgent, report.Summary)
	}

	return nil
}

// ────────────────────────────────────────────────────
//  Cross-File Integrity Check
// ────────────────────────────────────────────────────

// IntegrityResult — результат проверки целостности импортов.
type IntegrityResult struct {
	Valid         bool     `json:"valid"`
	MissingFiles  []string `json:"missing_files,omitempty"`
	BrokenImports []string `json:"broken_imports,omitempty"`
	TotalImports  int      `json:"total_imports"`
	ResolvedCount int      `json:"resolved_count"`
}

// CheckCrossFileIntegrity validates that import paths referenced in generated files
// actually exist in the file set. Checks @/ alias imports and relative imports.
// Returns informational result (non-blocking — missing files may be in node_modules).
func CheckCrossFileIntegrity(files map[string]string) *IntegrityResult {
	result := &IntegrityResult{Valid: true}

	// Build index of known file paths (normalized)
	knownFiles := make(map[string]bool, len(files))
	for name := range files {
		knownFiles[name] = true
		// Also index without extension for TS/TSX resolution
		for _, ext := range []string{".ts", ".tsx", ".js", ".jsx"} {
			if before, ok := strings.CutSuffix(name, ext); ok {
				knownFiles[before] = true
			}
		}
		// Index directory (for index.ts resolution)
		if idx := strings.LastIndex(name, "/"); idx > 0 {
			dir := name[:idx]
			knownFiles[dir] = true
		}
	}

	// Regex for import/require statements with @/ alias or relative paths
	importRe := regexp.MustCompile(`(?:import|from|require\()\s*['"](@/[^'"]+|\.\.?/[^'"]+)['"]`)

	missingSet := make(map[string]bool)

	for filename, content := range files {
		if !isSourceFile(filename) {
			continue
		}

		matches := importRe.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			importPath := m[1]
			result.TotalImports++

			resolved := resolveImportPath(filename, importPath)
			if resolved == "" {
				continue
			}

			if countImportResolved(knownFiles, resolved, importPath, filename, missingSet) {
				result.ResolvedCount++
			}
		}
	}

	for missing := range missingSet {
		result.MissingFiles = append(result.MissingFiles, missing)
	}

	if len(result.MissingFiles) > 5 {
		result.Valid = false
	}

	return result
}

func countImportResolved(knownFiles map[string]bool, resolved, importPath, filename string, missingSet map[string]bool) bool {
	if knownFiles[resolved] || knownFiles[resolved+"/index"] {
		return true
	}
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".css"} {
		if knownFiles[resolved+ext] {
			return true
		}
	}
	if isNodeModulePath(importPath) {
		return true
	}
	missingSet[importPath+" (in "+filename+")"] = true

	return false
}

// resolveImportPath converts an import path to a file path relative to project root.
func resolveImportPath(fromFile, importPath string) string {
	// @/ alias → src/
	if strings.HasPrefix(importPath, "@/") {
		return "src/" + importPath[2:]
	}

	// Relative path — resolve from importing file's directory
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		dir := fromFile
		if idx := strings.LastIndex(dir, "/"); idx >= 0 {
			dir = dir[:idx]
		} else {
			dir = ""
		}

		parts := strings.Split(importPath, "/")
		dirParts := strings.Split(dir, "/")

		for _, p := range parts {
			switch p {
			case ".":
				// stay
			case "..":
				if len(dirParts) > 0 {
					dirParts = dirParts[:len(dirParts)-1]
				}
			default:
				dirParts = append(dirParts, p)
			}
		}

		resolved := strings.Join(dirParts, "/")
		if resolved == "" {
			return importPath
		}

		return resolved
	}

	return ""
}

func isSourceFile(name string) bool {
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".vue", ".svelte"} {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}

	return false
}

func isNodeModulePath(path string) bool {
	// @/ paths that don't resolve are likely typos, not node_modules
	if strings.HasPrefix(path, "@/") {
		return false
	}
	// Relative paths are not node_modules
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		return false
	}

	return true
}

// importClauseRe captures default (1), namespace (2), named (3) specifiers and path (4).
var importClauseRe = regexp.MustCompile(`import\s+(?:([\w$]+)\s*,?\s*)?(?:\*\s*as\s+([\w$]+)\s*)?(?:\{([^}]*)\})?\s*from\s*['"](@/[^'"]+|\.\.?/[^'"]+)['"]`)

var sourceExts = []string{".ts", ".tsx", ".js", ".jsx"}

// BackfillMissingImports creates minimal stub files for local (@/ or relative) imports
// that don't resolve to any generated file. A missing module is a FATAL bundler error
// (white screen), so when a coder chunk fails this guarantees the project still renders.
// Stubs export the exact default/named symbols the importers reference (esbuild also
// errors on missing named exports). Mutates files in place; returns created stub paths.
type importStubSpec struct {
	def   string
	names map[string]bool
	star  bool
}

func BackfillMissingImports(files map[string]string) []string {
	known := buildKnownFilesIndex(files)
	needed := collectImportStubSpecs(files, known)

	return materializeImportStubs(files, needed)
}

func buildKnownFilesIndex(files map[string]string) map[string]bool {
	known := make(map[string]bool, len(files))
	for name := range files {
		known[name] = true
		for _, ext := range sourceExts {
			if before, ok := strings.CutSuffix(name, ext); ok {
				known[before] = true
			}
		}
		if idx := strings.LastIndex(name, "/"); idx > 0 {
			known[name[:idx]] = true
		}
	}

	return known
}

func importPathExists(known map[string]bool, resolved string) bool {
	if known[resolved] || known[resolved+"/index"] {
		return true
	}
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".css"} {
		if known[resolved+ext] {
			return true
		}
	}

	return false
}

func collectImportStubSpecs(files map[string]string, known map[string]bool) map[string]*importStubSpec {
	needed := map[string]*importStubSpec{}
	for filename, content := range files {
		if !isSourceFile(filename) {
			continue
		}
		recordMissingImports(filename, content, known, needed)
	}

	return needed
}

func recordMissingImports(filename, content string, known map[string]bool, needed map[string]*importStubSpec) {
	for _, m := range importClauseRe.FindAllStringSubmatch(content, -1) {
		full, defName, starName, named, importPath := m[0], m[1], m[2], m[3], m[4]
		if strings.HasPrefix(strings.TrimSpace(full), "import type") {
			continue
		}
		if isNodeModulePath(importPath) || strings.HasSuffix(importPath, ".css") {
			continue
		}
		resolved := resolveImportPath(filename, importPath)
		if resolved == "" || importPathExists(known, resolved) {
			continue
		}
		s := needed[resolved]
		if s == nil {
			s = &importStubSpec{names: map[string]bool{}}
			needed[resolved] = s
		}
		if defName != "" && s.def == "" {
			s.def = defName
		}
		if starName != "" {
			s.star = true
		}
		for n := range strings.SplitSeq(named, ",") {
			n = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(n), "type "))
			if i := strings.Index(n, " as "); i >= 0 {
				n = strings.TrimSpace(n[:i])
			}
			if n != "" {
				s.names[n] = true
			}
		}
	}
}

func materializeImportStubs(files map[string]string, needed map[string]*importStubSpec) []string {
	var created []string
	for resolved, s := range needed {
		path := resolved
		if !hasSourceExt(path) {
			path += ".tsx"
		}
		if _, ok := files[path]; ok {
			continue
		}
		files[path] = buildImportStub(s.names, s.def != "" || s.star || len(s.names) == 0)
		created = append(created, path)
	}

	return created
}

func hasSourceExt(p string) bool {
	for _, ext := range sourceExts {
		if strings.HasSuffix(p, ext) {
			return true
		}
	}

	return false
}

// buildImportStub renders a no-op module exporting the requested named symbols
// (and a default when needed). `any` typing keeps it valid as both component and value.
func buildImportStub(names map[string]bool, withDefault bool) string {
	var b strings.Builder
	b.WriteString("// AUTO-GENERATED STUB — original module was not produced (coder chunk failed).\n")
	b.WriteString("// Prevents a fatal bundler 'module not found' error so the app still renders.\n")
	b.WriteString("const Noop: any = () => null;\n")
	for n := range names {
		fmt.Fprintf(&b, "export const %s: any = Noop;\n", n)
	}
	if withDefault {
		b.WriteString("export default Noop;\n")
	}

	return b.String()
}
