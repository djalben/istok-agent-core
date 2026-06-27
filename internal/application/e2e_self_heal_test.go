package application_test

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  E2E Self-Heal Pipeline Tests
//  Уровень 1 — Micro (counter app)  →  проверяет validation trigger
//  Уровень 2 — Macro (greenhouse dashboard) → проверяет integrity trigger
//  Уровень 3 — Clean pass  →  подтверждает отсутствие лишних LLM-вызовов
//
//  Архитектура мока:
//    - первые CODER-вызовы возвращают намеренно «битый» код (flaw-режим)
//    - вызовы "SELF-HEALING PASS" перехватываются и возвращают починенный код
//    - все вызовы пишутся в callLog для детального post-mortem
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/djalben/istok-agent-core/internal/application"
	"github.com/djalben/istok-agent-core/internal/ports"
)

// ─────────────────────────────────────────────────────────────
//  healFlaw — вид намеренного дефекта в генерируемом коде
// ─────────────────────────────────────────────────────────────

type healFlaw int

const (
	healFlawNone           healFlaw = iota
	healFlawLoremIpsum              // код содержит "Lorem ipsum" → валидация блокирует
	healFlawMissingImports          // код импортирует 3+ несуществующих @/ модулей
)

// ─────────────────────────────────────────────────────────────
//  healAwareMock — LLM-мок с полным логом вызовов
// ─────────────────────────────────────────────────────────────

type healAwareMock struct {
	mu      sync.Mutex
	callLog []healCall
	flaw    healFlaw
}

type healCall struct {
	N      int
	Kind   string // "coder" | "heal-validation" | "heal-integrity" | "other"
	Prompt string // первые 400 символов
	Files  int    // сколько файлов запрошено/возвращено
	At     time.Time
}

func newHealAwareMock(flaw healFlaw) *healAwareMock {
	return &healAwareMock{flaw: flaw}
}

func (m *healAwareMock) Complete(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	lower := strings.ToLower(req.UserPrompt + req.SystemPrompt)

	m.mu.Lock()
	n := len(m.callLog) + 1
	m.mu.Unlock()

	var kind, content string

	switch {
	// ── Self-heal calls ──────────────────────────────────────────
	case strings.Contains(req.UserPrompt, "SELF-HEALING PASS"):
		if strings.Contains(req.UserPrompt, "MISSING MODULE IMPORTS") {
			kind = "heal-integrity"
			content = m.healIntegrityResponse()
		} else {
			kind = "heal-validation"
			content = m.healValidationResponse(req.UserPrompt)
		}

	// ── App-shell compositor ─────────────────────────────────────
	case strings.Contains(lower, "app shell") || strings.Contains(lower, "ui wiring"):
		kind = "appshell"
		content = `<file path="src/App.tsx">
import React from 'react';
export default function App() { return <div className="p-4">App</div>; }
</file>`

	// ── Chunked coder calls ──────────────────────────────────────
	case strings.Contains(lower, "files to generate in this batch"):
		kind = "coder"
		files := extractFileListFromPrompt(req.UserPrompt)
		content = m.coderResponse(files)

	// ── Fallback ─────────────────────────────────────────────────
	default:
		kind = "other"
		content = `{"architecture":"SPA","steps":[],"components":[]}`
	}

	preview := req.UserPrompt
	if len(preview) > 2000 {
		preview = preview[:2000] + "..."
	}
	m.mu.Lock()
	m.callLog = append(m.callLog, healCall{
		N: n, Kind: kind, Prompt: preview, At: time.Now(),
	})
	m.mu.Unlock()

	return &ports.LLMResponse{Content: content, TokensUsed: 500, Model: req.Model}, nil
}

// coderResponse генерирует XML-артефакты с намеренным дефектом (зависит от m.flaw).
func (m *healAwareMock) coderResponse(files []string) string {
	var sb strings.Builder
	for _, f := range files {
		fmt.Fprintf(&sb, "<file path=\"%s\">\n", f)
		sb.WriteString(m.fileContent(f))
		sb.WriteString("\n</file>\n\n")
	}
	return sb.String()
}

func (m *healAwareMock) fileContent(path string) string {
	name := extractComponentName(path)
	switch m.flaw {
	case healFlawLoremIpsum:
		return fmt.Sprintf(`import React from 'react';
export const %s: React.FC = () => (
  <div className="p-4">
    <h1>%s</h1>
    <p>Lorem ipsum dolor sit amet consectetur adipiscing elit</p>
  </div>
);
export default %s;`, name, name, name)

	case healFlawMissingImports:
		return fmt.Sprintf(`import React from 'react';
import { GhostA } from '@/components/GhostA';
import { GhostB } from '@/components/GhostB';
import { GhostC } from '@/components/GhostC';
export const %s: React.FC = () => (
  <div className="p-4"><GhostA /><GhostB /><GhostC /></div>
);
export default %s;`, name, name)

	default:
		return fmt.Sprintf(`import React from 'react';
export const %s: React.FC = () => (
  <div className="p-4"><h1>%s</h1><p>Content for %s</p></div>
);
export default %s;`, name, name, name, name)
	}
}

// healValidationResponse парсит "--- path ---" блоки из промпта и возвращает чистый код.
func (m *healAwareMock) healValidationResponse(prompt string) string {
	var sb strings.Builder
	rest := prompt
	for {
		start := strings.Index(rest, "\n--- ")
		if start == -1 {
			break
		}
		rest = rest[start+5:]
		end := strings.Index(rest, " ---\n")
		if end == -1 {
			break
		}
		filename := strings.TrimSpace(rest[:end])
		rest = rest[end:]
		if !strings.HasSuffix(filename, ".tsx") && !strings.HasSuffix(filename, ".ts") {
			continue
		}
		name := extractComponentName(filename)
		fmt.Fprintf(&sb, "<file path=\"%s\">\nimport React from 'react';\n"+
			"export const %s: React.FC = () => <div className=\"p-4\">%s — clean</div>;\n"+
			"export default %s;\n</file>\n\n", filename, name, name, name)
	}
	if sb.Len() == 0 {
		return `<file path="src/App.tsx">
import React from 'react';
export default function App() { return <div className="p-4">Healed App</div>; }
</file>`
	}
	return sb.String()
}

// healIntegrityResponse создаёт заглушки для GhostA/B/C — трёх сломанных импортов.
func (m *healAwareMock) healIntegrityResponse() string {
	return `<file path="src/components/GhostA.tsx">
import React from 'react';
export const GhostA: React.FC = () => <div>GhostA</div>;
export default GhostA;
</file>
<file path="src/components/GhostB.tsx">
import React from 'react';
export const GhostB: React.FC = () => <div>GhostB</div>;
export default GhostB;
</file>
<file path="src/components/GhostC.tsx">
import React from 'react';
export const GhostC: React.FC = () => <div>GhostC</div>;
export default GhostC;
</file>`
}

// ─── helpers ─────────────────────────────────────────────────

func extractFileListFromPrompt(prompt string) []string {
	marker := "FILES TO GENERATE IN THIS BATCH:"
	_, after, ok := strings.Cut(prompt, marker)
	if !ok {
		return []string{"src/App.tsx"}
	}
	endIdx := strings.Index(after, "\nRULES:")
	if endIdx == -1 {
		endIdx = strings.Index(after, "\nCRITICAL")
	}
	if endIdx == -1 {
		endIdx = len(after)
	}
	block := strings.TrimSpace(after[:endIdx])
	var files []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

// healCallsOf — фильтрует callLog по kind.
func (m *healAwareMock) healCallsOf(kind string) []healCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []healCall
	for _, c := range m.callLog {
		if c.Kind == kind {
			out = append(out, c)
		}
	}
	return out
}

// ─────────────────────────────────────────────────────────────
//  Общий манифест для тестов
// ─────────────────────────────────────────────────────────────

func smallCounterManifest() *application.SystemManifest {
	return &application.SystemManifest{
		ProjectName: "SimpleCounter",
		Type:        "frontend",
		Frontend: application.FrontendManifest{
			Framework: "react",
			Styling:   "tailwindcss",
		},
		FileMap: []string{
			"src/types/counter.ts",
			"src/hooks/useCounter.ts",
			"src/components/Counter.tsx",
			"src/components/ui/Button.tsx",
			"src/App.tsx",
		},
	}
}

func greenhouseManifest() *application.SystemManifest {
	return &application.SystemManifest{
		ProjectName: "GreenhouseDashboard",
		Type:        "frontend",
		Frontend: application.FrontendManifest{
			Framework: "react",
			Styling:   "tailwindcss",
		},
		FileMap: []string{
			"src/types/boiler.ts",
			"src/types/climate.ts",
			"src/types/yield.ts",
			"src/services/boilerApi.ts",
			"src/services/climateApi.ts",
			"src/hooks/useBoiler.ts",
			"src/hooks/useClimate.ts",
			"src/components/BoilerControl.tsx",
			"src/components/YieldChart.tsx",
			"src/components/ClimatePanel.tsx",
			"src/components/HydroponicStatus.tsx",
			"src/pages/Dashboard.tsx",
			"src/App.tsx",
		},
	}
}

// ─────────────────────────────────────────────────────────────
//  helpers для запуска GenerateCodeChunkedForTest
// ─────────────────────────────────────────────────────────────

func runChunked(t *testing.T, mock ports.LLMProvider, spec string, manifest *application.SystemManifest) map[string]string {
	t.Helper()
	orch := application.NewOrchestratorForTest(mock)
	bus := application.SubscribeOrchestratorEventsForTest(orch)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	go func() {
		for range bus {
		}
	}()
	files, err := application.GenerateCodeChunkedForTest(ctx, orch, spec, manifest, nil, nil, nil, nil, application.MediaContext{})
	if err != nil {
		t.Logf("generateCodeChunked returned err (may be partial success): %v", err)
	}
	return files
}

// ─────────────────────────────────────────────────────────────
//  Level 1 — Micro: Simple React counter
//  Триггер: валидация — Lorem ipsum
// ─────────────────────────────────────────────────────────────

func TestE2E_Level1_CounterApp_LoremIpsumTrigger(t *testing.T) {
	t.Parallel()
	mock := newHealAwareMock(healFlawLoremIpsum)
	files := runChunked(t, mock,
		"Create a simple React counter application with one button and basic state.",
		smallCounterManifest())

	t.Logf("=== PIPELINE TRACE ===")
	for _, c := range mock.callLog {
		t.Logf("[%02d] %-20s %s", c.N, c.Kind, c.At.Format("15:04:05.000"))
	}
	t.Logf("total files delivered: %d", len(files))

	// Verify: self-heal fired at least once
	healCalls := mock.healCallsOf("heal-validation")
	if len(healCalls) == 0 {
		t.Error("FAIL: self-heal validation trigger never fired — Lorem ipsum was not caught")
	} else {
		t.Logf("PASS: heal-validation triggered %d time(s)", len(healCalls))
	}

	// Verify: final files must NOT contain Lorem ipsum
	loremFound := ""
	for path, content := range files {
		lower := strings.ToLower(content)
		if strings.Contains(lower, "lorem ipsum") {
			loremFound = path
			break
		}
	}
	if loremFound != "" {
		t.Errorf("FAIL: Lorem ipsum still present in final output: %s", loremFound)
	} else {
		t.Log("PASS: no Lorem ipsum in final delivered files")
	}

	// Verify: coder was called (files were generated)
	if len(mock.healCallsOf("coder")) == 0 {
		t.Error("FAIL: coder was never called — no files generated")
	}
}

// ─────────────────────────────────────────────────────────────
//  Level 2 — Macro: Greenhouse dashboard
//  Триггер: cross-file integrity — 3+ сломанных @/ импорта
// ─────────────────────────────────────────────────────────────

func TestE2E_Level2_GreenhouseDashboard_IntegrityTrigger(t *testing.T) {
	t.Parallel()
	mock := newHealAwareMock(healFlawMissingImports)
	files := runChunked(t, mock,
		"Создание дашборда тепличного комплекса: управление 2-мегаваттным котлом, "+
			"графики урожайности голландских роз, климат-контроль и гидропоника. "+
			"Строгий премиальный минимализм, стекло, темная тема.",
		greenhouseManifest())

	t.Logf("=== PIPELINE TRACE ===")
	for _, c := range mock.callLog {
		t.Logf("[%02d] %-22s %s  prompt_head=%q",
			c.N, c.Kind, c.At.Format("15:04:05.000"),
			func() string {
				p := c.Prompt
				if len(p) > 80 {
					return p[:80]
				}
				return p
			}())
	}
	t.Logf("total files delivered: %d", len(files))

	// Verify: integrity heal fired
	integrityCalls := mock.healCallsOf("heal-integrity")
	if len(integrityCalls) == 0 {
		t.Error("FAIL: integrity self-heal never triggered — CheckCrossFileIntegrity missed 3+ missing imports")
	} else {
		t.Logf("PASS: heal-integrity triggered %d time(s)", len(integrityCalls))
		// The heal prompt must contain MISSING MODULE IMPORTS section
		for _, c := range integrityCalls {
			if !strings.Contains(c.Prompt, "MISSING MODULE IMPORTS") {
				t.Errorf("FAIL: heal-integrity prompt missing MISSING MODULE IMPORTS section")
			}
		}
	}

	// Verify: stub files were created and published
	wantStubs := []string{
		"src/components/GhostA.tsx",
		"src/components/GhostB.tsx",
		"src/components/GhostC.tsx",
	}
	for _, stub := range wantStubs {
		if _, ok := files[stub]; !ok {
			t.Errorf("FAIL: stub file not created: %s", stub)
		} else {
			t.Logf("PASS: stub file created: %s", stub)
		}
	}

	// Verify: final code does NOT still import GhostA/B/C into files without those stubs
	// (all stubs should now exist → imports are resolved)
	stubsExist := map[string]bool{
		"@/components/GhostA": files["src/components/GhostA.tsx"] != "",
		"@/components/GhostB": files["src/components/GhostB.tsx"] != "",
		"@/components/GhostC": files["src/components/GhostC.tsx"] != "",
	}
	for imp, exists := range stubsExist {
		if !exists {
			t.Errorf("FAIL: missing import not resolved: %s", imp)
		}
	}
}

// ─────────────────────────────────────────────────────────────
//  Level 3 — Clean pass: no flaw → no self-heal LLM call
// ─────────────────────────────────────────────────────────────

func TestE2E_Level3_CleanCode_NoHeal(t *testing.T) {
	t.Parallel()
	mock := newHealAwareMock(healFlawNone)
	files := runChunked(t, mock,
		"Create a simple React counter application with one button and basic state.",
		smallCounterManifest())

	t.Logf("=== PIPELINE TRACE ===")
	for _, c := range mock.callLog {
		t.Logf("[%02d] %-20s %s", c.N, c.Kind, c.At.Format("15:04:05.000"))
	}

	// Verify: no self-heal LLM call was made (clean code → skip)
	healCalls := append(mock.healCallsOf("heal-validation"), mock.healCallsOf("heal-integrity")...)
	if len(healCalls) > 0 {
		t.Errorf("FAIL: self-heal LLM called %d time(s) on clean code — unnecessary overhead", len(healCalls))
	} else {
		t.Log("PASS: no self-heal LLM call on clean code")
	}

	// Verify: files were generated
	if len(files) == 0 {
		t.Error("FAIL: no files generated on clean pass")
	} else {
		t.Logf("PASS: %d files generated cleanly, zero self-heal overhead", len(files))
	}
}

// ─────────────────────────────────────────────────────────────
//  Level 4 — Self-heal prompt structure audit
//  Проверяет что buildSelfHealPrompt содержит правильные секции
// ─────────────────────────────────────────────────────────────

func TestE2E_Level4_SelfHealPrompt_ContainsRequiredSections(t *testing.T) {
	t.Parallel()
	mock := newHealAwareMock(healFlawLoremIpsum)
	_ = runChunked(t, mock,
		"Create a React counter app.",
		smallCounterManifest())

	healCalls := mock.healCallsOf("heal-validation")
	if len(healCalls) == 0 {
		t.Skip("no heal-validation calls captured — cannot audit prompt structure")
	}

	prompt := healCalls[0].Prompt
	t.Logf("heal-validation prompt head:\n%s", prompt)

	required := []string{
		"SELF-HEALING PASS",
		"VALIDATION ERRORS",
	}
	for _, section := range required {
		if !strings.Contains(prompt, section) {
			t.Errorf("FAIL: self-heal prompt missing required section %q", section)
		} else {
			t.Logf("PASS: section %q present", section)
		}
	}
}

// ─────────────────────────────────────────────────────────────
//  Level 5 — Integrity prompt structure audit
// ─────────────────────────────────────────────────────────────

func TestE2E_Level5_IntegrityHealPrompt_AllowsNewFiles(t *testing.T) {
	t.Parallel()
	mock := newHealAwareMock(healFlawMissingImports)
	_ = runChunked(t, mock,
		"Greenhouse dashboard.",
		greenhouseManifest())

	healCalls := mock.healCallsOf("heal-integrity")
	if len(healCalls) == 0 {
		t.Skip("no heal-integrity calls captured")
	}

	prompt := healCalls[0].Prompt
	t.Logf("heal-integrity prompt head:\n%s", prompt)

	required := []string{
		"SELF-HEALING PASS",
		"You MAY create NEW files",
		"MISSING MODULE IMPORTS",
		"GhostA",
		"GhostB",
		"GhostC",
	}
	for _, section := range required {
		if !strings.Contains(prompt, section) {
			t.Errorf("FAIL: integrity prompt missing %q", section)
		} else {
			t.Logf("PASS: %q present in prompt", section)
		}
	}
}
