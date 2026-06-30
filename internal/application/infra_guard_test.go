package application

import (
	"strings"
	"testing"
)

// TestHasExport проверяет все паттерны определения экспорта.
func TestHasExport(t *testing.T) {
	cases := []struct {
		name    string
		content string
		export  string
		want    bool
	}{
		{"export const", `export const queryClient = new QueryClient()`, "queryClient", true},
		{"export let", `export let queryClient = null`, "queryClient", true},
		{"export function", `export function apiclient() {}`, "apiclient", true},
		{"export { name }", `export { queryClient, other }`, "queryClient", true},
		{"export default name", `export default apiclient`, "apiclient", true},
		{"named export without match", `export const otherThing = 1`, "queryClient", false},
		{"no export at all", `const queryClient = new QueryClient()`, "queryClient", false},
		{"partial name should not match", `export const queryClientExtra = 1`, "queryClient", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasExport(tc.content, tc.export); got != tc.want {
				t.Errorf("hasExport(%q, %q) = %v, want %v", tc.content[:min(len(tc.content), 40)], tc.export, got, tc.want)
			}
		})
	}
}

// TestEnsureApiClientExports_FileAbsent — если файла нет, инъекции нет.
func TestEnsureApiClientExports_FileAbsent(t *testing.T) {
	files := map[string]string{}
	injected := ensureApiClientExports(files)
	if injected {
		t.Fatal("expected no injection when file absent")
	}
}

// TestEnsureApiClientExports_BothPresent — если оба экспорта есть, инъекции нет.
func TestEnsureApiClientExports_BothPresent(t *testing.T) {
	files := map[string]string{
		"src/lib/api-client.ts": `
import { QueryClient } from "@tanstack/react-query";
export const queryClient = new QueryClient();
export const apiclient = { get: async () => {} };
`,
	}
	if injected := ensureApiClientExports(files); injected {
		t.Fatal("expected no injection when both exports present")
	}
}

// TestEnsureApiClientExports_MissingBoth — оба отсутствуют, оба инжектируются.
func TestEnsureApiClientExports_MissingBoth(t *testing.T) {
	const original = `// empty api-client\nconst baseURL = "/api";\n`
	files := map[string]string{"src/lib/api-client.ts": original}
	injected := ensureApiClientExports(files)
	if !injected {
		t.Fatal("expected injection when both exports missing")
	}
	result := files["src/lib/api-client.ts"]
	if !hasExport(result, "queryClient") {
		t.Error("queryClient not injected")
	}
	if !hasExport(result, "apiclient") {
		t.Error("apiclient not injected")
	}
}

// TestEnsureApiClientExports_MissingQueryClientOnly — только queryClient отсутствует.
func TestEnsureApiClientExports_MissingQueryClientOnly(t *testing.T) {
	files := map[string]string{
		"src/lib/api-client.ts": `export const apiclient = { get: async () => {} };`,
	}
	injected := ensureApiClientExports(files)
	if !injected {
		t.Fatal("expected injection for missing queryClient")
	}
	if !hasExport(files["src/lib/api-client.ts"], "queryClient") {
		t.Error("queryClient not injected")
	}
	if !hasExport(files["src/lib/api-client.ts"], "apiclient") {
		t.Error("apiclient should still be present")
	}
}

// TestEnsureApiClientExports_MissingApiClientOnly — только apiclient отсутствует.
func TestEnsureApiClientExports_MissingApiClientOnly(t *testing.T) {
	files := map[string]string{
		"src/lib/api-client.ts": `
import { QueryClient } from "@tanstack/react-query";
export const queryClient = new QueryClient();
`,
	}
	injected := ensureApiClientExports(files)
	if !injected {
		t.Fatal("expected injection for missing apiclient")
	}
	if !hasExport(files["src/lib/api-client.ts"], "queryClient") {
		t.Error("queryClient should still be present")
	}
	if !hasExport(files["src/lib/api-client.ts"], "apiclient") {
		t.Error("apiclient not injected")
	}
}

// TestGroupFileMap_InfraClassification — api-client.ts идёт в группу infra (Tier 0).
func TestGroupFileMap_InfraClassification(t *testing.T) {
	fileMap := []string{
		"src/lib/api-client.ts",
		"src/lib/query-client.ts",
		"src/lib/apiclient.ts",
		"src/lib/http-client.ts",
		"src/lib/utils.ts",
		"src/components/Dashboard.tsx",
		"src/types/index.ts",
	}
	groups := groupFileMap(fileMap)

	infraFiles := map[string]bool{}
	libFiles := map[string]bool{}
	for _, g := range groups {
		for _, f := range g.Files {
			// Auto-split produces "infra_1", "infra_2", etc. — use prefix match.
			if strings.HasPrefix(g.Name, "infra") {
				infraFiles[f] = true
			} else if strings.HasPrefix(g.Name, "lib") {
				libFiles[f] = true
			}
		}
	}

	mustBeInfra := []string{
		"src/lib/api-client.ts",
		"src/lib/query-client.ts",
		"src/lib/apiclient.ts",
		"src/lib/http-client.ts",
	}
	for _, f := range mustBeInfra {
		if !infraFiles[f] {
			t.Errorf("%q should be in infra group, got lib=%v", f, libFiles[f])
		}
	}

	if infraFiles["src/lib/utils.ts"] {
		t.Error("utils.ts should NOT be in infra group")
	}
	if !libFiles["src/lib/utils.ts"] {
		t.Error("utils.ts should be in lib group")
	}
}

// TestFixTemplateLiteralOverClose проверяет исправление двойной ')' в ${...}%.
func TestFixTemplateLiteralOverClose(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
		fixed bool
	}{
		{
			name:  "exact reported error",
			input: "width: `${Math.max(0, Math.min(100, (sensorData?.temperature || 0) / 30) * 100))}%` }}",
			want:  "width: `${Math.max(0, Math.min(100, (sensorData?.temperature || 0) / 30) * 100)}%` }}",
			fixed: true,
		},
		{
			name:  "percent with unit",
			input: "`${value * 100))}%`",
			want:  "`${value * 100)}%`",
			fixed: true,
		},
		{
			name:  "already correct — no double paren",
			input: "width: `${Math.min(100, val)}%`",
			want:  "width: `${Math.min(100, val)}%`",
			fixed: false,
		},
		{
			name:  "double paren not before template close — should not change",
			input: "fn(a, b))",
			want:  "fn(a, b))",
			fixed: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{"src/components/Test.tsx": tc.input}
			n := fixTemplateLiteralOverClose(files)
			got := files["src/components/Test.tsx"]
			if got != tc.want {
				t.Errorf("got:  %q\nwant: %q", got, tc.want)
			}
			wasFixed := n > 0
			if wasFixed != tc.fixed {
				t.Errorf("fixed=%v want %v", wasFixed, tc.fixed)
			}
		})
	}
}

// TestGroupFileMap_InfraTier0 — infra группа должна иметь Tier 0.
func TestGroupFileMap_InfraTier0(t *testing.T) {
	groups := groupFileMap([]string{"src/lib/api-client.ts", "vite.config.ts", "src/types/index.ts"})
	tiers := buildGenerationTiers(groups)
	for _, tier := range tiers {
		for _, g := range tier.Groups {
			if g.Name == "infra" && tier.Level != 0 {
				t.Errorf("infra group at tier %d, want 0", tier.Level)
			}
		}
	}
}
