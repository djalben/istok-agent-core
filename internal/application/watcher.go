package application

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Watcher V1 — Self-Healing Error Monitor
//  Принимает webhook-сигналы об ошибках, диагностирует
//  и формирует отчёт / автоисправление.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ErrorWebhookPayload — входящий сигнал об ошибке
type ErrorWebhookPayload struct {
	StatusCode int    `json:"status_code"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Message    string `json:"message"`
	Timestamp  string `json:"timestamp"`
	Source     string `json:"source"` // "railway", "vercel", "manual"
	LogTail    string `json:"log_tail,omitempty"`
}

// RepairReport — отчёт о диагностике и попытке исправления
type RepairReport struct {
	ID          string    `json:"id"`
	ReceivedAt  time.Time `json:"received_at"`
	Error       ErrorWebhookPayload `json:"error"`
	Diagnosis   string    `json:"diagnosis"`
	Action      string    `json:"action"`   // "self_test", "log_analysis", "llm_diagnosis", "notify_only"
	Result      string    `json:"result"`   // "fixed", "identified", "budget_exceeded", "unknown"
	Details     string    `json:"details"`
	CreditsUsed int       `json:"credits_used"`
}

// Watcher — сервис мониторинга и самовосстановления
type Watcher struct {
	orchestrator *Orchestrator
	baseURL      string // собственный URL для self-test (e.g. http://localhost:8080)

	mu             sync.Mutex
	dailyCredits   int
	maxCredits     int
	lastResetDay   int
	autoHealEnabled bool
	reports        []RepairReport
	logBuffer      []string // последние строки логов
	logMu          sync.Mutex
}

// NewWatcher creates Watcher with env-based config.
func NewWatcher(orchestrator *Orchestrator, selfBaseURL string) *Watcher {
	maxCredits := 10 // default
	if v := os.Getenv("MAX_AUTO_REPAIR_CREDITS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxCredits = n
		}
	}

	autoHeal := os.Getenv("AUTO_HEAL_ENABLED") == "true"

	w := &Watcher{
		orchestrator:    orchestrator,
		baseURL:         selfBaseURL,
		maxCredits:      maxCredits,
		autoHealEnabled: autoHeal,
		reports:         make([]RepairReport, 0),
		logBuffer:       make([]string, 0, 100),
	}

	log.Printf("🔭 Watcher V1 initialized | max_credits=%d auto_heal=%v base=%s", maxCredits, autoHeal, selfBaseURL)
	return w
}

// AppendLog добавляет строку в кольцевой буфер логов (последние 50 строк).
func (w *Watcher) AppendLog(line string) {
	w.logMu.Lock()
	defer w.logMu.Unlock()
	w.logBuffer = append(w.logBuffer, line)
	if len(w.logBuffer) > 50 {
		w.logBuffer = w.logBuffer[len(w.logBuffer)-50:]
	}
}

// GetReports returns all repair reports.
func (w *Watcher) GetReports() []RepairReport {
	w.mu.Lock()
	defer w.mu.Unlock()
	cp := make([]RepairReport, len(w.reports))
	copy(cp, w.reports)
	return cp
}

// HandleError — основная точка входа: принимает ошибку, сортирует, диагностирует.
func (w *Watcher) HandleError(ctx context.Context, payload ErrorWebhookPayload) RepairReport {
	w.resetDailyBudgetIfNeeded()

	report := RepairReport{
		ID:         fmt.Sprintf("wr-%d", time.Now().UnixMilli()),
		ReceivedAt: time.Now(),
		Error:      payload,
	}

	log.Printf("🔭 Watcher: received error signal | %d %s %s | source=%s",
		payload.StatusCode, payload.Method, payload.Path, payload.Source)

	// ── TOKEN GUARD ──
	if !w.canSpendCredits(1) {
		report.Action = "notify_only"
		report.Result = "budget_exceeded"
		report.Diagnosis = fmt.Sprintf("Daily repair budget exhausted (%d/%d). Switching to notify-only mode.", w.dailyCredits, w.maxCredits)
		report.Details = "Set MAX_AUTO_REPAIR_CREDITS higher or wait until tomorrow."
		log.Printf("⚠️ Watcher: budget exceeded (%d/%d), notify only", w.dailyCredits, w.maxCredits)
		w.saveReport(report)
		return report
	}

	// ── TRIAGE ──
	switch {
	case payload.StatusCode == 404:
		report = w.triage404(ctx, payload, report)
	case payload.StatusCode >= 500:
		report = w.triage5xx(ctx, payload, report)
	default:
		report.Action = "notify_only"
		report.Result = "identified"
		report.Diagnosis = fmt.Sprintf("Non-critical error %d on %s %s", payload.StatusCode, payload.Method, payload.Path)
		report.Details = payload.Message
	}

	w.saveReport(report)
	return report
}

// ── TRIAGE: 404 — self-test to find working route ──

func (w *Watcher) triage404(ctx context.Context, payload ErrorWebhookPayload, report RepairReport) RepairReport {
	report.Action = "self_test"
	log.Printf("🔍 Watcher: 404 triage — self-testing routes for %s %s", payload.Method, payload.Path)

	// Test the reported path
	testPaths := []string{
		payload.Path,
		"/api/v1/generate/stream",
		"/api/v1/generate",
		"/api/v1/health",
	}

	var results []string
	var workingPath string

	client := &http.Client{Timeout: 10 * time.Second}

	for _, path := range testPaths {
		url := w.baseURL + path
		method := payload.Method
		if method == "" {
			method = "GET"
		}

		var resp *http.Response
		var err error

		if method == "POST" {
			resp, err = client.Post(url, "application/json", strings.NewReader(`{"specification":"watcher self-test","mode":"code"}`))
		} else {
			resp, err = client.Get(url)
		}

		if err != nil {
			results = append(results, fmt.Sprintf("  %s %s → ERROR: %v", method, path, err))
			continue
		}
		resp.Body.Close()

		status := resp.StatusCode
		results = append(results, fmt.Sprintf("  %s %s → %d", method, path, status))

		if status >= 200 && status < 400 && workingPath == "" && path != "/api/v1/health" {
			workingPath = path
		}
	}

	report.Details = strings.Join(results, "\n")

	if workingPath != "" && workingPath != payload.Path {
		report.Diagnosis = fmt.Sprintf("Route mismatch: %s returns 404, but %s returns 200. Frontend should use %s.", payload.Path, workingPath, workingPath)
		report.Result = "identified"
		w.spendCredits(1)
	} else if workingPath == payload.Path {
		report.Diagnosis = fmt.Sprintf("Route %s actually works (self-test returned 200). The 404 may be transient or from a different deployment.", payload.Path)
		report.Result = "identified"
	} else {
		report.Diagnosis = "No working route found via self-test. Possible full outage or route not registered."
		report.Result = "unknown"
		w.spendCredits(1)
	}

	log.Printf("🔍 Watcher 404 result: %s", report.Diagnosis)
	return report
}

// ── TRIAGE: 5xx — read last 50 log lines + LLM diagnosis ──

func (w *Watcher) triage5xx(ctx context.Context, payload ErrorWebhookPayload, report RepairReport) RepairReport {
	report.Action = "log_analysis"
	log.Printf("🩺 Watcher: 5xx triage — analyzing logs for %s %s (status=%d)", payload.Method, payload.Path, payload.StatusCode)

	// Get last 50 log lines
	logTail := w.getLogTail()
	if payload.LogTail != "" {
		logTail = payload.LogTail // use provided log tail if available
	}

	if logTail == "" {
		report.Diagnosis = fmt.Sprintf("Server error %d on %s %s. No logs available for analysis.", payload.StatusCode, payload.Method, payload.Path)
		report.Result = "unknown"
		report.Details = "Log buffer is empty. Error may have occurred before Watcher started capturing logs."
		return report
	}

	// Parse logs with Auto-Healer
	errors := ParseRailwayLogs(logTail)

	if len(errors) == 0 {
		report.Diagnosis = fmt.Sprintf("Server error %d on %s %s. Logs present but no recognizable error patterns found.", payload.StatusCode, payload.Method, payload.Path)
		report.Result = "identified"
		report.Details = "Last 50 log lines:\n" + truncate(logTail, 2000)
		return report
	}

	// Format error summary
	var errLines []string
	for _, e := range errors {
		errLines = append(errLines, fmt.Sprintf("[%s] %s", e.Type, e.Message))
	}
	report.Details = strings.Join(errLines, "\n")

	// If we have budget and it's a panic — try LLM diagnosis
	if hasComplexError(errors) && w.canSpendCredits(1) {
		report.Action = "llm_diagnosis"
		w.spendCredits(1)

		commands := w.orchestrator.DiagnoseAndHeal(ctx, logTail)
		if len(commands) > 0 {
			var cmdLines []string
			for _, cmd := range commands {
				cmdLines = append(cmdLines, fmt.Sprintf("[P%d] %s → %s: %s", cmd.Priority, cmd.Action, cmd.Target, cmd.Description))
			}
			report.Diagnosis = fmt.Sprintf("Found %d errors, LLM generated %d fix commands.", len(errors), len(commands))
			report.Details += "\n\nFix commands:\n" + strings.Join(cmdLines, "\n")

			if w.autoHealEnabled {
				report.Result = "identified"
				report.Details += "\n\n⚠️ AUTO_HEAL_ENABLED=true but auto-push to main is NOT implemented yet (safety)."
			} else {
				report.Result = "identified"
				report.Details += "\n\nAUTO_HEAL_ENABLED=false — manual intervention required."
			}
		} else {
			report.Diagnosis = fmt.Sprintf("Found %d errors but LLM couldn't generate fix commands.", len(errors))
			report.Result = "identified"
		}
	} else {
		report.Diagnosis = fmt.Sprintf("Found %d errors: %s", len(errors), errLines[0])
		report.Result = "identified"
	}

	log.Printf("🩺 Watcher 5xx result: %s", report.Diagnosis)
	return report
}

// ── Token Guard helpers ──

func (w *Watcher) resetDailyBudgetIfNeeded() {
	w.mu.Lock()
	defer w.mu.Unlock()
	today := time.Now().YearDay()
	if w.lastResetDay != today {
		w.dailyCredits = 0
		w.lastResetDay = today
		log.Printf("🔄 Watcher: daily credit budget reset (max=%d)", w.maxCredits)
	}
}

func (w *Watcher) canSpendCredits(n int) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.dailyCredits+n <= w.maxCredits
}

func (w *Watcher) spendCredits(n int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.dailyCredits += n
	log.Printf("💰 Watcher: spent %d credit(s), used=%d/%d", n, w.dailyCredits, w.maxCredits)
}

func (w *Watcher) saveReport(r RepairReport) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.reports = append(w.reports, r)
	// Keep last 50 reports
	if len(w.reports) > 50 {
		w.reports = w.reports[len(w.reports)-50:]
	}
}

func (w *Watcher) getLogTail() string {
	w.logMu.Lock()
	defer w.logMu.Unlock()
	return strings.Join(w.logBuffer, "\n")
}

// ── Log Writer adapter — captures log output into Watcher buffer ──

// WatcherLogWriter is an io.Writer that tees log output to both
// the original writer and the Watcher's ring buffer.
type WatcherLogWriter struct {
	Original io.Writer
	Watcher  *Watcher
}

func (wlw *WatcherLogWriter) Write(p []byte) (n int, err error) {
	line := strings.TrimRight(string(p), "\n")
	if line != "" {
		wlw.Watcher.AppendLog(line)
	}
	return wlw.Original.Write(p)
}
