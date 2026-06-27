package application

import (
	"strings"
	"testing"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
)

// ─────────────────────────────────────────────────────────
//  collectBrokenFiles
// ─────────────────────────────────────────────────────────

func TestCollectBrokenFiles_OnlyReturnsCriticals(t *testing.T) {
	t.Parallel()
	result := &usecases.ValidationResult{
		Issues: []usecases.ValidationIssue{
			{File: "src/App.tsx", Severity: usecases.SeverityCritical},
			{File: "src/utils.ts", Severity: usecases.SeverityWarning},
			{File: "src/main.tsx", Severity: usecases.SeverityCritical},
			{File: "", Severity: usecases.SeverityCritical}, // no file — skip
		},
	}
	got := collectBrokenFiles(result)
	if len(got) != 2 {
		t.Fatalf("want 2 broken files, got %d: %v", len(got), got)
	}
	if got[0] != "src/App.tsx" || got[1] != "src/main.tsx" {
		t.Errorf("unexpected order: %v", got)
	}
}

func TestCollectBrokenFiles_Dedup(t *testing.T) {
	t.Parallel()
	result := &usecases.ValidationResult{
		Issues: []usecases.ValidationIssue{
			{File: "src/App.tsx", Severity: usecases.SeverityCritical},
			{File: "src/App.tsx", Severity: usecases.SeverityCritical},
		},
	}
	got := collectBrokenFiles(result)
	if len(got) != 1 {
		t.Fatalf("want 1 deduped file, got %d", len(got))
	}
}

func TestCollectBrokenFiles_EmptyResult(t *testing.T) {
	t.Parallel()
	got := collectBrokenFiles(&usecases.ValidationResult{})
	if len(got) != 0 {
		t.Fatalf("want empty, got %v", got)
	}
}

// ─────────────────────────────────────────────────────────
//  extractImportingFiles
// ─────────────────────────────────────────────────────────

func TestExtractImportingFiles_ParsesFormat(t *testing.T) {
	t.Parallel()
	integrity := &usecases.IntegrityResult{
		MissingFiles: []string{
			"@/components/Button (in src/App.tsx)",
			"@/hooks/useAuth (in src/pages/Dashboard.tsx)",
			"@/components/Button (in src/pages/Dashboard.tsx)", // duplicate importer
			"malformed entry with no in-clause",
		},
	}
	got := extractImportingFiles(integrity)
	if len(got) != 2 {
		t.Fatalf("want 2 unique importers, got %d: %v", len(got), got)
	}
	if got[0] != "src/App.tsx" || got[1] != "src/pages/Dashboard.tsx" {
		t.Errorf("unexpected importers: %v", got)
	}
}

func TestExtractImportingFiles_Empty(t *testing.T) {
	t.Parallel()
	got := extractImportingFiles(&usecases.IntegrityResult{})
	if len(got) != 0 {
		t.Fatalf("want empty, got %v", got)
	}
}

// ─────────────────────────────────────────────────────────
//  mergeUniq
// ─────────────────────────────────────────────────────────

func TestMergeUniq_NoDuplicates(t *testing.T) {
	t.Parallel()
	a := []string{"a", "b"}
	b := []string{"b", "c"}
	got := mergeUniq(a, b)
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] want %q got %q", i, want[i], got[i])
		}
	}
}

func TestMergeUniq_EmptyInputs(t *testing.T) {
	t.Parallel()
	if got := mergeUniq(nil, nil); len(got) != 0 {
		t.Fatalf("want empty, got %v", got)
	}
	if got := mergeUniq([]string{"x"}, nil); len(got) != 1 || got[0] != "x" {
		t.Fatalf("unexpected: %v", got)
	}
}

// ─────────────────────────────────────────────────────────
//  buildSelfHealPrompt
// ─────────────────────────────────────────────────────────

func TestBuildSelfHealPrompt_IncludesIntegritySection(t *testing.T) {
	t.Parallel()
	result := &usecases.ValidationResult{Passed: true}
	integrity := &usecases.IntegrityResult{
		Valid:        false,
		MissingFiles: make([]string, minIntegrityMissing),
	}
	for i := range integrity.MissingFiles {
		integrity.MissingFiles[i] = "@/components/X (in src/App.tsx)"
	}

	prompt := buildSelfHealPrompt("spec", nil, map[string]string{}, result, integrity, nil)

	if !strings.Contains(prompt, "MISSING MODULE IMPORTS") {
		t.Error("expected MISSING MODULE IMPORTS section in prompt")
	}
	if !strings.Contains(prompt, "You MAY create NEW files") {
		t.Error("expected new-files permission in prompt when integrity is broken")
	}
}

func TestBuildSelfHealPrompt_NoIntegritySectionWhenClean(t *testing.T) {
	t.Parallel()
	result := &usecases.ValidationResult{Passed: false}
	integrity := &usecases.IntegrityResult{Valid: true, MissingFiles: []string{"one"}} // below threshold

	prompt := buildSelfHealPrompt("spec", nil, map[string]string{}, result, integrity, nil)

	if strings.Contains(prompt, "MISSING MODULE IMPORTS") {
		t.Error("should NOT include integrity section when missing count < minIntegrityMissing")
	}
}

func TestBuildSelfHealPrompt_TruncatesLongFiles(t *testing.T) {
	t.Parallel()
	bigContent := strings.Repeat("x", 5000)
	files := map[string]string{"src/Big.tsx": bigContent}
	result := &usecases.ValidationResult{}
	integrity := &usecases.IntegrityResult{}

	prompt := buildSelfHealPrompt("spec", nil, files, result, integrity, []string{"src/Big.tsx"})

	if strings.Contains(prompt, strings.Repeat("x", 4001)) {
		t.Error("file content should be truncated at 4000 chars")
	}
	if !strings.Contains(prompt, "truncated") {
		t.Error("expected truncation marker in prompt")
	}
}
