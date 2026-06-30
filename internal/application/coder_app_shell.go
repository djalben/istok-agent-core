package application

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"gitlab.com/libs-artifex/wrapper/v2"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — App Shell Composer (Wiring Gate)
//  Детерминированный финальный шаг: гарантирует, что src/App.tsx подключает
//  реальный UI. Без него генерация могла отдать слой данных без презентации,
//  оставив инертную scaffold-заглушку → белый экран.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// isScaffoldAppTsx — true, если App.tsx пуст или дословно равен scaffold-плейсхолдеру
// (Кодер так и не переписал его под реальное приложение).
func isScaffoldAppTsx(code string) bool {
	c := strings.TrimSpace(code)

	return c == "" || c == strings.TrimSpace(scaffoldFiles["src/App.tsx"])
}

// hasPresentationLayer — есть ли в проекте хоть одна реальная страница/маршрут/лейаут.
// Голый слой данных (types/services/hooks) без presentation = нечего рендерить.
func hasPresentationLayer(files map[string]string) bool {
	for name, code := range files {
		if strings.TrimSpace(code) == "" {
			continue
		}
		nl := strings.ToLower(name)
		if strings.Contains(nl, "/routes/") || strings.Contains(nl, "/pages/") ||
			strings.Contains(nl, "/layout/") {
			return true
		}
	}

	return false
}

// appShellNeeded — срабатывает гейт: App.tsx остался заглушкой ИЛИ нет ни одной
// страницы/лейаута. В обоих случаях экран будет пустым — нужна досборка UI.
func appShellNeeded(files map[string]string) bool {
	return isScaffoldAppTsx(files["src/App.tsx"]) || !hasPresentationLayer(files)
}

// isAppShellFile — какие файлы разрешено применять из вывода композера: только
// презентационная обвязка (App.tsx, роутер, маршруты/страницы, лейаут).
func isAppShellFile(name string) bool {
	nl := strings.ToLower(name)
	if nl == "src/app.tsx" || nl == "src/router.tsx" || nl == "src/routetree.gen.ts" {
		return true
	}

	return strings.Contains(nl, "/routes/") || strings.Contains(nl, "/pages/") ||
		strings.Contains(nl, "/layout/")
}

// finalizeAppShell — пост-генерационный гейт + досборка. Если src/App.tsx не
// подключает реальный UI, запускает фокусный LLM-шаг, который пишет работающий
// App.tsx (роутер на memory-history + реальные страницы/лейаут). Если LLM не
// справился — ставит детерминированную, гарантированно рендерящуюся заглушку,
// чтобы пользователь НИКОГДА не видел пустой инертный экран.
func (o *Orchestrator) finalizeAppShell(ctx context.Context, specification string, manifest *SystemManifest, files map[string]string) {
	if !appShellNeeded(files) {
		return
	}
	applog(ctx).WarnContext(ctx, "app shell gate tripped — composing UI wiring",
		"appPlaceholder", isScaffoldAppTsx(files["src/App.tsx"]),
		"hasPresentation", hasPresentationLayer(files),
		"totalFiles", len(files))
	o.sendStatus(ctx, RoleCoder, "running", "🔧 Финальная сборка: подключаю UI к приложению...", 92)

	agent := o.agents[RoleCoder]
	// Свежий фоновый контекст: родительский ctx исчерпывается сразу после длинного
	// coder-run (20+ мин), что вызывало context.DeadlineExceeded немедленно.
	// App shell — постгенерационный LLM-шаг, он должен работать независимо.
	shellCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	userPrompt := buildAppShellPrompt(specification, manifest, files)
	content, err := o.callLLMWithReasoning(shellCtx, agent.Model, appShellSystemPrompt, userPrompt, 16384)
	if err != nil {
		applog(ctx).WarnContext(ctx, "app shell composer failed", "error", wrapper.Wrap(err))
	} else {
		gen := o.parseCodeFiles(ctx, content)
		applied := 0
		for name, code := range gen {
			if isAppShellFile(name) && strings.TrimSpace(code) != "" {
				files[name] = code
				o.busFromCtx(ctx).PublishFile(RoleCoder, name, code)
				applied++
			}
		}
		applog(ctx).InfoContext(ctx, "app shell composed", "filesApplied", applied)
	}

	// Hard guarantee: never ship the inert placeholder.
	if isScaffoldAppTsx(files["src/App.tsx"]) {
		files["src/App.tsx"] = fallbackAppShell(manifest)
		o.busFromCtx(ctx).PublishFile(RoleCoder, "src/App.tsx", files["src/App.tsx"])
		applog(ctx).WarnContext(ctx, "app shell deterministic fallback applied")
	}
}

// buildAppShellPrompt собирает фокусный промпт: спека + реально сгенерированные
// файлы по категориям, чтобы композер импортировал СУЩЕСТВУЮЩИЕ модули, а не
// выдумывал пути.
func buildAppShellPrompt(specification string, manifest *SystemManifest, files map[string]string) string {
	pages := collectByMatch(files, "/routes/", "/pages/")
	layout := collectByMatch(files, "/layout/")
	components := collectComponents(files)
	hooks := collectByMatch(files, "/hooks/")
	services := collectByMatch(files, "/services/")
	stores := collectByMatch(files, "/stores/", "/store/")

	projectName := "Application"
	if manifest != nil && manifest.ProjectName != "" {
		projectName = manifest.ProjectName
	}

	spec := specification
	if len(spec) > 1200 {
		spec = spec[:1200] + "..."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Compose the root UI for project %q.\n\nSPEC:\n%s\n\n", projectName, spec)
	b.WriteString("The data layer is already generated. Wire it into a FULLY RENDERED dashboard UI.\n")
	b.WriteString("Import ONLY files that exist in the lists below (via @/* aliases). Do NOT import files that are not listed.\n\n")
	writeFileSection(&b, "EXISTING PAGES/ROUTES", pages)
	writeFileSection(&b, "EXISTING LAYOUT", layout)
	writeFileSection(&b, "EXISTING COMPONENTS", components)
	writeFileSection(&b, "EXISTING HOOKS", hooks)
	writeFileSection(&b, "EXISTING SERVICES", services)
	writeFileSection(&b, "EXISTING STORES", stores)
	b.WriteString("\nTASK:\n")
	b.WriteString("1. Write a COMPLETE src/App.tsx that renders the full dashboard for the spec — NOT a loading placeholder.\n")
	b.WriteString("2. Wrap content in QueryClientProvider (create a QueryClient inline if no client file exists) and <InspectorProvider> from '@/components/InspectorProvider'.\n")
	b.WriteString("3. If you use routing, use @tanstack/react-router with createMemoryHistory (NEVER browser history). If no routes exist, render the dashboard sections directly without a router.\n")
	b.WriteString("4. Build real, data-rich UI sections relevant to the spec using Tailwind + lucide-react icons. Every section must render visible content immediately.\n")
	b.WriteString("5. Add data-component-name to root elements. Provide safe fallbacks for any dynamic lookups (never undefined).\n\n")
	b.WriteString("Output ONLY <file path=\"src/App.tsx\">...</file> (plus optional <file> blocks for src/router.tsx or src/routes/* and src/components/layout/* if you create them). Raw code, no markdown fences, no JSON.\n")

	return b.String()
}

// collectByMatch возвращает отсортированные имена непустых файлов, чьи пути
// содержат любую из подстрок.
func collectByMatch(files map[string]string, subs ...string) []string {
	var out []string
	for name, code := range files {
		if strings.TrimSpace(code) == "" {
			continue
		}
		nl := strings.ToLower(name)
		for _, s := range subs {
			if strings.Contains(nl, s) {
				out = append(out, name)

				break
			}
		}
	}
	sort.Strings(out)

	return out
}

// collectComponents — компоненты вне ui/layout (доменные), непустые.
func collectComponents(files map[string]string) []string {
	var out []string
	for name, code := range files {
		if strings.TrimSpace(code) == "" {
			continue
		}
		nl := strings.ToLower(name)
		if !strings.Contains(nl, "/components/") {
			continue
		}
		if strings.Contains(nl, "/ui/") || strings.Contains(nl, "/layout/") ||
			strings.Contains(nl, "inspectorprovider") {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)

	return out
}

// writeFileSection печатает секцию списка файлов (с потолком, чтобы не раздувать промпт).
func writeFileSection(b *strings.Builder, title string, names []string) {
	if len(names) == 0 {
		fmt.Fprintf(b, "%s: (none)\n", title)

		return
	}
	const maxShown = 40
	shown := names
	if len(shown) > maxShown {
		shown = shown[:maxShown]
	}
	fmt.Fprintf(b, "%s (%d):\n", title, len(names))
	for _, n := range shown {
		fmt.Fprintf(b, "  - %s\n", n)
	}
}

// fallbackAppShell — детерминированный, самодостаточный App.tsx БЕЗ внешних
// импортов (кроме react), который гарантированно рендерится. Последний рубеж,
// когда композер недоступен/провалился: показывает оболочку проекта и список
// фич из манифеста, а не пустой экран.
func fallbackAppShell(manifest *SystemManifest) string {
	projectName := "Application"
	var features []FeatureSpec
	if manifest != nil {
		if manifest.ProjectName != "" {
			projectName = manifest.ProjectName
		}
		features = manifest.Features
	}

	var cards strings.Builder
	if len(features) == 0 {
		cards.WriteString(`        <div className="rounded-2xl border border-zinc-800 bg-zinc-900/50 p-6 shadow-2xl"><h3 className="text-lg font-semibold text-zinc-100">Dashboard</h3><p className="mt-2 text-sm text-zinc-400">Интерфейс готовится к работе.</p></div>`)
	}
	for _, f := range features {
		name := jsEscape(f.Name)
		desc := jsEscape(f.Description)
		fmt.Fprintf(&cards, `        <div className="rounded-2xl border border-zinc-800 bg-zinc-900/50 p-6 shadow-2xl transition-all duration-300 hover:border-zinc-700"><h3 className="text-lg font-semibold text-zinc-100">%s</h3><p className="mt-2 text-sm leading-relaxed text-zinc-400">%s</p></div>
`, name, desc)
	}

	return fmt.Sprintf(`export default function App() {
  return (
    <div data-component-name="AppShell" className="min-h-screen bg-zinc-950 text-zinc-100">
      <header className="border-b border-zinc-800 px-8 py-6">
        <h1 className="text-2xl font-semibold tracking-tight">%s</h1>
        <p className="mt-1 text-sm text-zinc-400">Панель управления</p>
      </header>
      <main className="grid grid-cols-1 gap-6 p-8 sm:grid-cols-2 lg:grid-cols-3">
%s
      </main>
    </div>
  );
}
`, jsEscape(projectName), strings.TrimRight(cards.String(), "\n"))
}

// ensureAppTsxDefaultExport гарантирует наличие default export в src/App.tsx.
// Если Кодер выдал named export (export const App / export function App) без
// сопроводительного export default — main.tsx не сможет его импортировать и
// упадёт с "No matching export in App.tsx for import default".
// Решение детерминировано: добавляет одну строку в конец файла, не трогая логику.
func ensureAppTsxDefaultExport(files map[string]string) {
	const key = "src/App.tsx"
	code := files[key]
	if code == "" {
		return
	}
	if strings.Contains(code, "export default") {
		return
	}
	// Named export без default: добавляем export default App; в конец.
	if strings.Contains(code, "const App") || strings.Contains(code, "function App") {
		files[key] = strings.TrimRight(code, "\n") + "\nexport default App;\n"
	}
}

// ensureApiClientExports — детерминированный guard для src/lib/api-client.ts.
// Если LLM не добавил экспорты queryClient или apiclient — все импортеры
// (компоненты, хуки, сервисы) упадут с "has no exported member".
// Вызывается дважды: после Tier 0 (mid-pipeline, до компонентов) и
// в финальном post-processing `generateCodeChunked` (belt-and-suspenders).
// Возвращает true если была выполнена инъекция (для логирования/thought).
func ensureApiClientExports(files map[string]string) bool {
	const path = "src/lib/api-client.ts"
	content, ok := files[path]
	if !ok {
		return false
	}

	const queryClientInject = `
// --- deterministic guard: queryClient ---
import { QueryClient } from "@tanstack/react-query";
export const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 60_000, retry: 1 } },
});`

	const apiClientInject = `
// --- deterministic guard: apiclient ---
export const apiclient = {
  get: async <T>(url: string, init?: RequestInit): Promise<T> => {
    const res = await fetch(url, { ...init, method: "GET" });
    if (!res.ok) throw new Error("HTTP " + res.status);
    return res.json() as T;
  },
  post: async <T>(url: string, body: unknown, init?: RequestInit): Promise<T> => {
    const res = await fetch(url, {
      ...init, method: "POST",
      headers: { "Content-Type": "application/json", ...((init?.headers as Record<string, string>) ?? {}) },
      body: JSON.stringify(body),
    });
    if (!res.ok) throw new Error("HTTP " + res.status);
    return res.json() as T;
  },
};`

	var additions []string
	if !hasExport(content, "queryClient") {
		additions = append(additions, queryClientInject)
	}
	if !hasExport(content, "apiclient") {
		additions = append(additions, apiClientInject)
	}
	if len(additions) == 0 {
		return false
	}
	files[path] = strings.TrimRight(content, "\n") + "\n" + strings.Join(additions, "\n") + "\n"

	return true
}

// hasExport возвращает true если TypeScript-код содержит именованный или default
// экспорт с заданным именем. Проверяет суффикс после имени (пробел, =, :, ()
// чтобы избежать ложных совпадений с именами-надмножествами (queryClientExtra ≠ queryClient).
func hasExport(content, name string) bool {
	// const/let: name followed by space, =, :, or newline (covers "const x =", "const x:")
	for _, suffix := range []string{" ", "=", ":", "\n", "\t"} {
		if strings.Contains(content, "export const "+name+suffix) ||
			strings.Contains(content, "export let "+name+suffix) {
			return true
		}
	}
	// function: name followed by ( or space
	if strings.Contains(content, "export function "+name+"(") ||
		strings.Contains(content, "export function "+name+" ") {
		return true
	}
	// re-export / named export block: "export { name" (covers "export { name }" and "export { name, ...")
	if strings.Contains(content, "export { "+name) {
		return true
	}
	// default export
	return strings.Contains(content, "export default "+name)
}

// jsEscape экранирует строку для безопасной вставки в JSX-текст/атрибут.
func jsEscape(s string) string {
	r := strings.NewReplacer(
		"\\", "\\\\",
		"{", "&#123;",
		"}", "&#125;",
		"<", "&lt;",
		">", "&gt;",
	)

	return r.Replace(strings.TrimSpace(s))
}

// appShellSystemPrompt — системный промпт композера. Наследует Titan-директивы
// (включая SAFE ROUTING), добавляет роль «финального сборщика приложения».
const appShellSystemPrompt = PremiumDesignSystem + TitanSystemDirectives + criticalMediaContract + `APP COMPOSER ROLE:
You are the final assembler. The project's data layer (types, services, hooks, stores) and some components already exist. Your ONLY job is to produce a working src/App.tsx (and, if helpful, a router + minimal pages/layout) that renders a COMPLETE, data-rich dashboard UI for the spec.
HARD RULES:
- NEVER output a loading spinner or "Загрузка…" placeholder as the final UI. Render real sections with real content immediately.
- Import ONLY modules that the user lists as existing. Do not invent import paths.
- Routing (if any) MUST use in-memory history (createMemoryHistory for TanStack Router / MemoryRouter for react-router-dom). Browser history is FORBIDDEN — it blanks the iframe.
- Wrap the tree in QueryClientProvider and <InspectorProvider>.
- Use Tailwind utility classes + lucide-react icons. Add data-component-name to root elements. Provide fallbacks for every dynamic lookup.
- CRITICAL OUTPUT FORMAT: wrap each file in <file path="...">raw code</file>. NO markdown fences. NO JSON. NO prose outside <file> tags.`
