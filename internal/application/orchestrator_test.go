package application

import (
	"context"
	"testing"
	"time"
)

// TestOrchestrator_GenerateTaxiGoEndToEnd tests the full chunked generation pipeline
// with a mock LLM. Validates:
//   - maxFilesPerChunk is enforced (max 2 files per LLM call)
//   - XML artifact parsing works correctly
//   - Partial success on truncation: closed <file> blocks are preserved
//   - No deadlocks: completes within timeout
func TestOrchestrator_GenerateTaxiGoEndToEnd(t *testing.T) {
	// ── Setup: Mock LLM with truncation on 2nd file ──
	mock := newMockLLM().withTruncation(1) // only 1 of 2 files per chunk will be complete
	orch := NewOrchestrator(mock)

	// Drain event bus so orchestrator doesn't block on publish
	go func() {
		ch := orch.events.Subscribe()
		for range ch {
			// discard
		}
	}()

	// ── Build a realistic manifest with 10 files across multiple categories ──
	manifest := &SystemManifest{
		ProjectName: "TaxiGo",
		Type:        "frontend",
		Frontend: FrontendManifest{
			Framework:       "react",
			Styling:         "tailwindcss",
			StateManagement: "zustand",
		},
		FileMap: []string{
			"src/types/booking.ts",
			"src/types/driver.ts",
			"src/lib/utils.ts",
			"src/lib/api-client.ts",
			"src/hooks/useBooking.ts",
			"src/components/ui/Button.tsx",
			"src/components/ui/Card.tsx",
			"src/components/BookingForm.tsx",
			"src/components/DriverCard.tsx",
			"src/pages/HomePage.tsx",
		},
	}

	plan := &MasterPlan{
		Architecture: "React SPA with TanStack Router",
		Components:   []string{"BookingForm", "DriverCard", "Map"},
		Steps:        []string{"Setup project", "Create components", "Add routing"},
	}

	spec := "TaxiGo — сервис заказа такси. Главная страница с формой бронирования, карточками водителей и картой."

	// ── Execute: run chunked generation with hard timeout ──
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	files, err := orch.generateCodeChunked(ctx, spec, manifest, plan, nil, nil, nil)

	// ── Assertions ──

	// 1. Must complete without deadlock (context not cancelled)
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("DEADLOCK: generateCodeChunked did not complete within 30s timeout")
	}

	// 2. Must not return error (partial success is OK)
	if err != nil {
		t.Fatalf("generateCodeChunked returned error: %v", err)
	}

	// 3. Must have generated SOME files (partial success acceptance)
	if len(files) == 0 {
		t.Fatal("generateCodeChunked returned 0 files — expected partial success from truncation")
	}

	t.Logf("✅ Generated %d files (out of %d in FileMap)", len(files), len(manifest.FileMap))
	for path := range files {
		t.Logf("   📄 %s", path)
	}

	// 4. Chunk size enforcement: no single LLM call should have > maxFilesPerGroup files
	mock.mu.Lock()
	maxSeen := mock.maxFilesPerReq
	mock.mu.Unlock()

	if maxSeen > maxFilesPerGroup {
		t.Errorf("CHUNK SIZE VIOLATION: LLM was asked for %d files in one call, max allowed is %d",
			maxSeen, maxFilesPerGroup)
	}
	t.Logf("✅ Max files per LLM call: %d (limit: %d)", maxSeen, maxFilesPerGroup)

	// 5. With truncation=1, each chunk of 2 yields 1 file.
	// 10 files / 2 per chunk = 5 chunks. Each yields 1 file → expect ~5 files.
	// Some groups may have 1 file (no truncation effect) → could be more.
	if len(files) < 3 {
		t.Errorf("Too few files: got %d, expected at least 3 from partial success", len(files))
	}

	// 6. Verify files have actual content (not empty)
	for path, content := range files {
		if len(content) < 10 {
			t.Errorf("File %s has suspiciously short content (%d bytes)", path, len(content))
		}
	}

	// 7. Total LLM calls should match number of chunks
	totalCalls := mock.callCount.Load()
	t.Logf("✅ Total LLM calls: %d", totalCalls)
	if totalCalls == 0 {
		t.Error("No LLM calls were made")
	}
}

// TestParseCodeFiles_XMLArtifacts tests the XML artifact parser directly.
func TestParseCodeFiles_XMLArtifacts(t *testing.T) {
	mock := newMockLLM()
	orch := NewOrchestrator(mock)

	tests := []struct {
		name     string
		input    string
		expected int // expected number of files
	}{
		{
			name: "clean XML artifacts",
			input: `<file path="src/App.tsx">
import React from 'react';
export default function App() { return <div>Hello</div>; }
</file>

<file path="src/index.tsx">
import React from 'react';
import ReactDOM from 'react-dom';
ReactDOM.render(<App />, document.getElementById('root'));
</file>`,
			expected: 2,
		},
		{
			name: "truncated XML — last file incomplete",
			input: `<file path="src/Button.tsx">
export const Button = () => <button>Click</button>;
</file>

<file path="src/Card.tsx">
export const Card = () => <div className="card">
  <p>This file was truncated by max_tok`,
			expected: 1, // only Button.tsx has a closing </file> tag
		},
		{
			name: "XML with thinking block",
			input: `<thinking>
Let me plan the component structure...
</thinking>

<file path="src/utils.ts">
export const formatDate = (d: Date) => d.toLocaleDateString();
</file>`,
			expected: 1,
		},
		{
			name:     "JSON fallback (legacy)",
			input:    `{"index.html":"<!DOCTYPE html><html><body>Hello</body></html>"}`,
			expected: 1,
		},
		{
			name:     "empty response",
			input:    "",
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			files := orch.parseCodeFiles(tc.input)
			if len(files) != tc.expected {
				t.Errorf("parseCodeFiles() returned %d files, expected %d", len(files), tc.expected)
				for k, v := range files {
					t.Logf("  got: %s (%d bytes)", k, len(v))
				}
			}
		})
	}
}

// TestInspectorProviderInjection verifies InspectorProvider is injected into React projects.
func TestInspectorProviderInjection(t *testing.T) {
	// React project (has .tsx files) — should inject
	reactFiles := map[string]string{
		"src/App.tsx":             "export default function App() {}",
		"src/components/Hero.tsx": "export const Hero = () => <div>Hero</div>",
		"src/lib/utils.ts":        "export const cn = () => '';",
		"src/types/index.ts":      "export interface User {}",
	}
	injectInspectorProvider(reactFiles)

	if _, ok := reactFiles[inspectorProviderPath]; !ok {
		t.Errorf("InspectorProvider was NOT injected into React project")
	}
	if len(reactFiles[inspectorProviderPath]) < 100 {
		t.Errorf("InspectorProvider content too short: %d bytes", len(reactFiles[inspectorProviderPath]))
	}
	t.Logf("✅ InspectorProvider injected (%d bytes)", len(reactFiles[inspectorProviderPath]))

	// Single-file HTML project — should NOT inject
	htmlFiles := map[string]string{
		"index.html": "<!DOCTYPE html><html><body>Hello</body></html>",
	}
	injectInspectorProvider(htmlFiles)
	if _, ok := htmlFiles[inspectorProviderPath]; ok {
		t.Errorf("InspectorProvider should NOT be injected into single-file HTML project")
	}
}

// TestChunkSizeEnforcement verifies groupFileMap respects maxFilesPerGroup.
func TestChunkSizeEnforcement(t *testing.T) {
	// Create a large FileMap with many components
	fileMap := []string{
		"src/components/BookingForm.tsx",
		"src/components/DriverCard.tsx",
		"src/components/MapView.tsx",
		"src/components/RideHistory.tsx",
		"src/components/PaymentForm.tsx",
		"src/components/ProfileEditor.tsx",
		"src/components/ReviewModal.tsx",
		"src/components/NotificationList.tsx",
		"src/components/SettingsPanel.tsx",
		"src/components/ChatWidget.tsx",
	}

	groups := groupFileMap(fileMap)

	for _, g := range groups {
		if len(g.Files) > maxFilesPerGroup {
			t.Errorf("Group %q has %d files, exceeds maxFilesPerGroup=%d",
				g.Name, len(g.Files), maxFilesPerGroup)
		}
		t.Logf("Group %q: %d files", g.Name, len(g.Files))
	}

	// With 10 files and max 2 per group, expect at least 5 groups
	if len(groups) < 5 {
		t.Errorf("Expected at least 5 groups for 10 files with max %d per group, got %d",
			maxFilesPerGroup, len(groups))
	}
}
