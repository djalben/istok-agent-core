package usecases

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Tester Agent (Layer 3)
//  Запускает go test и vitest/npm test после генерации кода.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// TestStatus — итоговый статус прогона тестов.
type TestStatus string

const (
	TestStatusPassed  TestStatus = "passed"
	TestStatusFailed  TestStatus = "failed"
	TestStatusSkipped TestStatus = "skipped" // toolchain недоступен или нет тестовых файлов
	TestStatusError   TestStatus = "error"   // setup/runtime error
)

// TestResult — результат одного прогона (go test или vitest).
type TestResult struct {
	Runner   string        `json:"runner"` // "go" | "vitest" | "npm"
	Status   TestStatus    `json:"status"`
	Output   string        `json:"output,omitempty"`
	Duration time.Duration `json:"duration"`
	ExitCode int           `json:"exit_code"`
}

// TesterReport — агрегированный отчёт от TesterAgent.
type TesterReport struct {
	Approved bool         `json:"approved"`
	Results  []TestResult `json:"results"`
	Summary  string       `json:"summary"`
	FixHint  string       `json:"fix_hint,omitempty"` // лог ошибок для контекста Кодера при retry
}

// CriticalFailures возвращает количество runners со статусом Failed.
func (r *TesterReport) CriticalFailures() int {
	n := 0
	for _, res := range r.Results {
		if res.Status == TestStatusFailed || res.Status == TestStatusError {
			n++
		}
	}
	return n
}

// ForCoderContext генерирует компактный лог ошибок для вставки в промпт Кодера при retry.
func (r *TesterReport) ForCoderContext() string {
	if r.Approved || len(r.Results) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## TEST FAILURES (fix ALL before returning code)\n\n")
	for _, res := range r.Results {
		if res.Status == TestStatusFailed || res.Status == TestStatusError {
			b.WriteString(fmt.Sprintf("### %s [%s] (exit=%d, %v)\n", res.Runner, res.Status, res.ExitCode, res.Duration))
			out := res.Output
			if len(out) > 1500 {
				out = out[:1500] + "\n...(truncated)"
			}
			b.WriteString("```\n" + out + "\n```\n\n")
		}
	}
	if r.FixHint != "" {
		b.WriteString("\nFIX HINT: " + r.FixHint + "\n")
	}
	return b.String()
}

// ────────────────────────────────────────────────────
//  TesterAgent
// ────────────────────────────────────────────────────

// TesterAgent выполняет тесты на сгенерированном коде.
// Файлы пишутся во временную директорию, далее запускаются go test и/или vitest.
type TesterAgent struct {
	GoTimeout      time.Duration
	NodeTimeout    time.Duration
	NPMTestCommand string // переопределяемая команда (default: "vitest run")
}

// NewTesterAgent создаёт агента с разумными дефолтами по таймаутам.
func NewTesterAgent() *TesterAgent {
	return &TesterAgent{
		GoTimeout:      90 * time.Second,
		NodeTimeout:    120 * time.Second,
		NPMTestCommand: "vitest run",
	}
}

// RunTests прогоняет все доступные тесты на сгенерированных файлах.
// files — карта путь → содержимое. Записываются во временную директорию.
// Если тулчейн недоступен, runner помечается Skipped (не Failed).
func (t *TesterAgent) RunTests(ctx context.Context, files map[string]string) *TesterReport {
	if len(files) == 0 {
		return &TesterReport{
			Approved: false,
			Summary:  "no files to test",
		}
	}

	// Materialize files в temp dir
	workDir, err := os.MkdirTemp("", "istok-tester-*")
	if err != nil {
		return &TesterReport{
			Approved: false,
			Summary:  fmt.Sprintf("tester setup failed: %v", err),
		}
	}
	defer os.RemoveAll(workDir)

	if err := materializeFiles(workDir, files); err != nil {
		return &TesterReport{
			Approved: false,
			Summary:  fmt.Sprintf("file materialize failed: %v", err),
		}
	}

	report := &TesterReport{}
	hasGoFiles := false
	hasJSFiles := false
	for name := range files {
		switch {
		case strings.HasSuffix(name, ".go"):
			hasGoFiles = true
		case strings.HasSuffix(name, ".ts"), strings.HasSuffix(name, ".tsx"),
			strings.HasSuffix(name, ".js"), strings.HasSuffix(name, ".jsx"):
			hasJSFiles = true
		}
	}

	if hasGoFiles {
		report.Results = append(report.Results, t.runGoTest(ctx, workDir))
	}
	if hasJSFiles {
		report.Results = append(report.Results, t.runJSTest(ctx, workDir, files))
	}

	if len(report.Results) == 0 {
		report.Approved = true
		report.Summary = "no testable files (no .go/.ts/.tsx/.js)"
		return report
	}

	// Approved только если ни один runner не Failed/Error
	failed := report.CriticalFailures()
	passed := 0
	skipped := 0
	for _, r := range report.Results {
		switch r.Status {
		case TestStatusPassed:
			passed++
		case TestStatusSkipped:
			skipped++
		}
	}

	if failed > 0 {
		report.Approved = false
		report.Summary = fmt.Sprintf("FAILED: %d runners failed, %d passed, %d skipped", failed, passed, skipped)
		var hints []string
		for _, r := range report.Results {
			if r.Status == TestStatusFailed || r.Status == TestStatusError {
				hints = append(hints, fmt.Sprintf("%s exit=%d", r.Runner, r.ExitCode))
			}
		}
		report.FixHint = "Fix failing tests: " + strings.Join(hints, ", ")
	} else {
		report.Approved = true
		report.Summary = fmt.Sprintf("PASSED: %d passed, %d skipped", passed, skipped)
	}

	return report
}

// runGoTest исполняет `go test ./...` в workDir.
func (t *TesterAgent) runGoTest(ctx context.Context, workDir string) TestResult {
	res := TestResult{Runner: "go"}
	start := time.Now()

	if _, err := exec.LookPath("go"); err != nil {
		res.Status = TestStatusSkipped
		res.Output = "go toolchain not in PATH"
		res.Duration = time.Since(start)
		return res
	}

	// Если нет go.mod — go test не запустится корректно. Проверяем.
	if _, err := os.Stat(filepath.Join(workDir, "go.mod")); err != nil {
		res.Status = TestStatusSkipped
		res.Output = "no go.mod, skipping go test"
		res.Duration = time.Since(start)
		return res
	}

	cctx, cancel := context.WithTimeout(ctx, t.GoTimeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, "go", "test", "./...")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	res.Output = string(out)
	res.Duration = time.Since(start)

	if cctx.Err() == context.DeadlineExceeded {
		res.Status = TestStatusError
		res.Output = "go test timed out: " + res.Output
		res.ExitCode = -1
		return res
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		res.ExitCode = exitErr.ExitCode()
		res.Status = TestStatusFailed
		log.Printf("🧪 go test failed (exit=%d, %v)", res.ExitCode, res.Duration)
	} else if err != nil {
		res.Status = TestStatusError
		res.Output = err.Error() + "\n" + res.Output
	} else {
		res.Status = TestStatusPassed
		log.Printf("🧪 go test PASSED (%v)", res.Duration)
	}
	return res
}

// runJSTest выполняет vitest или npm test, в зависимости от наличия package.json и tooling.
func (t *TesterAgent) runJSTest(ctx context.Context, workDir string, files map[string]string) TestResult {
	res := TestResult{Runner: "vitest"}
	start := time.Now()

	// Need package.json
	if _, err := os.Stat(filepath.Join(workDir, "package.json")); err != nil {
		res.Status = TestStatusSkipped
		res.Output = "no package.json, skipping JS tests"
		res.Duration = time.Since(start)
		return res
	}

	// Detect tool: bun > npx vitest > npm test
	var cmdName string
	var cmdArgs []string

	switch {
	case lookPath("bun"):
		cmdName = "bun"
		cmdArgs = []string{"test"}
		res.Runner = "bun"
	case lookPath("npx"):
		cmdName = "npx"
		cmdArgs = []string{"--yes", "vitest", "run", "--reporter=basic"}
		res.Runner = "vitest"
	case lookPath("npm"):
		cmdName = "npm"
		cmdArgs = []string{"test", "--silent"}
		res.Runner = "npm"
	default:
		res.Status = TestStatusSkipped
		res.Output = "no JS toolchain (bun/npx/npm) in PATH"
		res.Duration = time.Since(start)
		return res
	}

	// Need at least one *.test.ts(x) or *.spec.ts(x) file
	hasTestFile := false
	for name := range files {
		lower := strings.ToLower(name)
		if strings.Contains(lower, ".test.") || strings.Contains(lower, ".spec.") {
			hasTestFile = true
			break
		}
	}
	if !hasTestFile {
		res.Status = TestStatusSkipped
		res.Output = "no .test.* or .spec.* files, skipping"
		res.Duration = time.Since(start)
		return res
	}

	cctx, cancel := context.WithTimeout(ctx, t.NodeTimeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, cmdName, cmdArgs...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "CI=true", "NODE_ENV=test")
	out, err := cmd.CombinedOutput()
	res.Output = string(out)
	res.Duration = time.Since(start)

	if cctx.Err() == context.DeadlineExceeded {
		res.Status = TestStatusError
		res.Output = "JS test timed out: " + res.Output
		res.ExitCode = -1
		return res
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		res.ExitCode = exitErr.ExitCode()
		res.Status = TestStatusFailed
		log.Printf("🧪 %s failed (exit=%d, %v)", res.Runner, res.ExitCode, res.Duration)
	} else if err != nil {
		res.Status = TestStatusError
		res.Output = err.Error() + "\n" + res.Output
	} else {
		res.Status = TestStatusPassed
		log.Printf("🧪 %s PASSED (%v)", res.Runner, res.Duration)
	}
	return res
}

// ────────────────────────────────────────────────────
//  Helpers
// ────────────────────────────────────────────────────

// materializeFiles записывает map "path → content" в директорию.
// Создаёт промежуточные директории. Игнорирует абсолютные пути и .. -сегменты.
func materializeFiles(root string, files map[string]string) error {
	for relPath, content := range files {
		clean := filepath.Clean(relPath)
		// Защита от path traversal
		if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
			continue
		}
		full := filepath.Join(root, clean)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func lookPath(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}
