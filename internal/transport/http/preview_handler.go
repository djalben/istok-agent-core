package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	api "github.com/evanw/esbuild/pkg/api"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// PreviewHandler — серверная сборка предпросмотра (competitor-grade, region-proof).
//
// Проблема: клиентские бандлеры (Sandpack classic -> col.csbops.io, Nodebox ->
// registry.npmjs.org) тянут зависимости из браузера пользователя по зарубежным CDN,
// которые в его сети блокируются (CORS/ERR_FAILED/timeout). Решение: бандлим на СЕРВЕРЕ
// (Railway имеет доступ к esm.sh / cdn.tailwindcss.com) и отдаём самодостаточный HTML,
// который браузер грузит ТОЛЬКО с нашего домена. Никаких клиентских fetch к npm.
type PreviewHandler struct {
	cache  sync.Map // sessionID -> *previewCacheEntry (live-generation flow)
	byHash sync.Map // contentHash -> html (POST-build flow, session-independent)
}

func NewPreviewHandler() *PreviewHandler { return &PreviewHandler{} }

type previewCacheEntry struct {
	hash string
	html string
}

// httpFetchCache — кэш скачанных с esm.sh модулей (общий для всех сборок),
// чтобы не дёргать сеть на каждый импорт react/recharts/radix.
var (
	httpFetchCache = sync.Map{} // url -> []byte
	previewHTTP    = &http.Client{Timeout: 30 * time.Second}
)

const esmCDN = "https://esm.sh/"

// Handle отдаёт собранный HTML предпросмотра: GET /api/v1/preview/{session_id}.
func (h *PreviewHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)

		return
	}
	ctx := r.Context()
	sessionID := r.PathValue("session_id")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("session_id")
	}
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session_id required")

		return
	}

	files := globalFileStore.Get(sessionID)
	if len(files) == 0 {
		writeError(w, http.StatusNotFound, "no files for session")

		return
	}

	hash := hashFiles(files)
	if cached, ok := h.cache.Load(sessionID); ok {
		if entry, valid := cached.(*previewCacheEntry); valid && entry.hash == hash {
			writeHTML(w, entry.html)

			return
		}
	}

	html, err := buildPreviewHTML(ctx, files)
	if err != nil {
		logFrom(ctx).ErrorContext(ctx, "preview build failed",
			"sessionId", sessionID, "error", wrapper.Wrap(err))
		writeError(w, http.StatusUnprocessableEntity, "preview build failed: "+err.Error())

		return
	}
	h.cache.Store(sessionID, &previewCacheEntry{hash: hash, html: html})
	logFrom(ctx).InfoContext(ctx, "preview built",
		"sessionId", sessionID, "files", len(files), "bytes", len(html))
	writeHTML(w, html)
}

type previewBuildRequest struct {
	Files map[string]string `json:"files"`
}

// HandleBuild собирает предпросмотр из файлов в теле запроса (без привязки к сессии
// генерации). Это позволяет показывать серверный предпросмотр для ЛЮБОГО проекта —
// загруженного из БД, отредактированного локально и т.д., а не только во время живой
// генерации. Результат кэшируется по content-hash; ответ — { "id": "<hash>" }, по
// которому фронт открывает iframe на GET /api/v1/preview/view/{id}.
// POST /api/v1/preview
func (h *PreviewHandler) HandleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)

		return
	}
	ctx := r.Context()
	var req previewBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")

		return
	}
	if len(req.Files) == 0 {
		writeError(w, http.StatusBadRequest, "files required")

		return
	}

	hash := hashFiles(req.Files)
	if _, ok := h.byHash.Load(hash); ok {
		_ = writeJSON(w, http.StatusOK, map[string]string{"id": hash})

		return
	}

	html, err := buildPreviewHTML(ctx, req.Files)
	if err != nil {
		logFrom(ctx).ErrorContext(ctx, "preview build failed (post)", "error", wrapper.Wrap(err))
		writeError(w, http.StatusUnprocessableEntity, "preview build failed: "+err.Error())

		return
	}
	h.byHash.Store(hash, html)
	logFrom(ctx).InfoContext(ctx, "preview built (post)", "files", len(req.Files), "bytes", len(html))
	_ = writeJSON(w, http.StatusOK, map[string]string{"id": hash})
}

// HandleView отдаёт ранее собранный (HandleBuild) HTML по content-hash.
// GET /api/v1/preview/view/{id}
func (h *PreviewHandler) HandleView(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)

		return
	}
	cached, ok := h.byHash.Load(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "preview not found")

		return
	}
	html, _ := cached.(string)
	writeHTML(w, html)
}

// HandleTailwind проксирует Tailwind Play CDN через наш домен, чтобы браузер
// пользователя не ходил на cdn.tailwindcss.com напрямую (тот блокируется в его сети).
// GET /api/v1/preview/tailwind.js
func (h *PreviewHandler) HandleTailwind(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)

		return
	}
	ctx := r.Context()
	body, err := fetchURL("https://cdn.tailwindcss.com")
	if err != nil {
		logFrom(ctx).ErrorContext(ctx, "tailwind proxy failed", "error", wrapper.Wrap(err))
		writeError(w, http.StatusBadGateway, "tailwind proxy failed")

		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = w.Write(body)
}

func writeHTML(w http.ResponseWriter, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = io.WriteString(w, html)
}

func hashFiles(files map[string]string) string {
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	hsh := sha256.New()
	for _, k := range keys {
		_, _ = io.WriteString(hsh, k)
		_, _ = io.WriteString(hsh, "\x00")
		_, _ = io.WriteString(hsh, files[k])
		_, _ = io.WriteString(hsh, "\x00")
	}

	return hex.EncodeToString(hsh.Sum(nil))
}

// buildPreviewHTML bundles the generated project into a single self-contained HTML
// document. npm bare-imports resolve from esm.sh at BUILD time (server-side) and are
// inlined; Tailwind is delivered via our proxied Play CDN at runtime.
//
// Resilience contract: a hallucinated/broken import (missing asset, wrong path,
// non-existent npm package) MUST NOT fail the whole build. Such imports are silently
// replaced with safe stubs (data-URI for images, null React component for modules) and
// logged, so the preview always renders even if a few pieces are broken.
func buildPreviewHTML(ctx context.Context, files map[string]string) (string, error) {
	tmp, err := os.MkdirTemp("", "istok-preview-*")
	if err != nil {
		return "", wrapper.Wrap(err)
	}
	defer os.RemoveAll(tmp)

	normalized := make(map[string]string, len(files))
	for path, content := range files {
		rel := strings.TrimPrefix(filepath.ToSlash(path), "/")
		if rel == "" || isPreviewSkippedFile(rel) {
			continue
		}
		normalized[rel] = content
		abs := filepath.Join(tmp, filepath.FromSlash(rel))
		if mkErr := os.MkdirAll(filepath.Dir(abs), 0o755); mkErr != nil {
			return "", wrapper.Wrap(mkErr)
		}
		if wErr := os.WriteFile(abs, []byte(content), 0o644); wErr != nil {
			return "", wrapper.Wrap(wErr)
		}
	}

	entry := detectPreviewEntry(normalized)
	if entry == "" {
		// The coder built components/App but skipped the mount boilerplate. Rather
		// than fail, synthesize a valid entry that imports the root component and
		// mounts it (mirrors the frontend Sandpack path). Only overlay if we can't
		// even find a root component to mount.
		if synthPath, synthContent, ok := synthesizePreviewEntry(normalized); ok {
			normalized[synthPath] = synthContent
			abs := filepath.Join(tmp, filepath.FromSlash(synthPath))
			if mkErr := os.MkdirAll(filepath.Dir(abs), 0o755); mkErr != nil {
				return "", wrapper.Wrap(mkErr)
			}
			if wErr := os.WriteFile(abs, []byte(synthContent), 0o644); wErr != nil {
				return "", wrapper.Wrap(wErr)
			}
			entry = synthPath
			logFrom(ctx).InfoContext(ctx, "preview synthesized missing entry", "entry", synthPath)
		}
	}
	if entry == "" {
		logFrom(ctx).WarnContext(ctx, "preview no entry -> overlay", "files", len(normalized))

		return previewErrorHTML("No React entry point found",
			"Couldn't locate a file that mounts React (looked for src/main.tsx, src/App.tsx, "+
				"or any file calling createRoot / ReactDOM.render), and no root component to synthesize one from."), nil
	}

	result := api.Build(api.BuildOptions{
		AbsWorkingDir:   tmp,
		EntryPoints:     []string{entry},
		Bundle:          true,
		Write:           false,
		Format:          api.FormatESModule,
		Platform:        api.PlatformBrowser,
		Target:          api.ES2020,
		JSX:             api.JSXAutomatic,
		JSXImportSource: "react",
		LogLevel:        api.LogLevelSilent,
		Define: map[string]string{
			"process.env.NODE_ENV": `"production"`,
			"global":               "window",
		},
		Plugins: []api.Plugin{cdnResolverPlugin(ctx, tmp)},
	})
	for _, warn := range result.Warnings {
		logFrom(ctx).WarnContext(ctx, "preview build warning", "text", warn.Text)
	}
	if len(result.Errors) > 0 {
		msgs := api.FormatMessages(result.Errors, api.FormatMessagesOptions{
			Color: false, Kind: api.ErrorMessage, TerminalWidth: 100,
		})
		joined := strings.Join(msgs, "\n")
		logFrom(ctx).WarnContext(ctx, "preview bundle errors -> overlay",
			"count", len(result.Errors), "messages", joined)

		return previewErrorHTML("Build error", joined), nil
	}
	if len(result.OutputFiles) == 0 {
		logFrom(ctx).WarnContext(ctx, "preview empty bundle -> overlay")

		return previewErrorHTML("Empty bundle", "esbuild produced no output for this project."), nil
	}

	var bundle strings.Builder
	for _, out := range result.OutputFiles {
		if strings.HasSuffix(out.Path, ".js") {
			bundle.Write(out.Contents)
		}
	}

	return wrapPreviewHTML(bundle.String()), nil
}

// Stub namespaces for graceful degradation.
const (
	nsAssetStub  = "istok-asset-stub"  // image/font/media import that can't be bundled
	nsStubModule = "istok-stub-module" // missing local file or failed npm package
)

// transparentPNG — 1x1 transparent PNG as a data URI, served as the default export of
// any image import that fails to resolve (broken <img src> instead of a crashed build).
const transparentPNG = "data:image/png;base64," +
	"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="

// cdnResolverPlugin resolves bare npm specifiers to esm.sh URLs and fetches their
// contents at build time. Unresolvable local files, asset imports and failed npm
// packages are replaced with safe stubs (never fatal). CSS imports are neutralized
// (Tailwind utilities come from the runtime Play CDN).
func cdnResolverPlugin(ctx context.Context, tmp string) api.Plugin {
	return api.Plugin{
		Name: "istok-cdn-resolver",
		Setup: func(build api.PluginBuild) {
			// Per-file build logging: every local source file that gets bundled.
			build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "file"},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					logFrom(ctx).InfoContext(ctx, "preview bundling file", "path", filepath.Base(args.Path))

					return api.OnLoadResult{}, nil // pass through to css/default loader
				})

			// Drop CSS imported from local files — Play CDN handles Tailwind utilities.
			build.OnLoad(api.OnLoadOptions{Filter: `\.css$`, Namespace: "file"},
				func(_ api.OnLoadArgs) (api.OnLoadResult, error) {
					empty := ""

					return api.OnLoadResult{Contents: &empty, Loader: api.LoaderJS}, nil
				})

			// Resolve every import originating from a local file.
			build.OnResolve(api.OnResolveOptions{Filter: `.*`, Namespace: "file"},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					return resolveLocalImport(ctx, tmp, args), nil
				})

			// Absolute https imports (e.g. esm.sh internal redirects).
			build.OnResolve(api.OnResolveOptions{Filter: `^https?://`},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					return api.OnResolveResult{Path: args.Path, Namespace: "http-url"}, nil
				})

			// Relative imports inside fetched modules -> resolve against importer URL.
			build.OnResolve(api.OnResolveOptions{Filter: `.*`, Namespace: "http-url"},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					base, err := url.Parse(args.Importer)
					if err != nil {
						return api.OnResolveResult{}, wrapper.Wrap(err)
					}
					ref, err := url.Parse(args.Path)
					if err != nil {
						return api.OnResolveResult{}, wrapper.Wrap(err)
					}

					return api.OnResolveResult{
						Path:      base.ResolveReference(ref).String(),
						Namespace: "http-url",
					}, nil
				})

			// Fetch remote module contents (cached). On failure -> stub, never fatal.
			build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "http-url"},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					body, err := fetchURL(args.Path)
					if err != nil {
						logFrom(ctx).WarnContext(ctx, "preview npm fetch failed, stubbing",
							"url", args.Path, "error", wrapper.Wrap(err))
						contents := stubModuleJS("__c")

						return api.OnLoadResult{Contents: &contents, Loader: api.LoaderJS}, nil
					}
					contents := string(body)
					loader := api.LoaderJS
					if strings.Contains(args.Path, ".css") {
						loader = api.LoaderCSS
					}

					return api.OnLoadResult{Contents: &contents, Loader: loader}, nil
				})

			// Asset stub: images -> data-URI default + null ReactComponent (svgr); else "".
			build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: nsAssetStub},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					def := `""`
					if isImageExt(strings.ToLower(filepath.Ext(args.Path))) {
						def = strconv.Quote(transparentPNG)
					}
					contents := stubModuleJS(def)

					return api.OnLoadResult{Contents: &contents, Loader: api.LoaderJS}, nil
				})

			// Missing-module stub: any import (default/named/namespace) yields a null
			// React component, so a broken reference renders nothing instead of crashing.
			build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: nsStubModule},
				func(_ api.OnLoadArgs) (api.OnLoadResult, error) {
					contents := stubModuleJS("__c")

					return api.OnLoadResult{Contents: &contents, Loader: api.LoaderJS}, nil
				})
		},
	}
}

// resolveLocalImport handles every import seen from a local source file: bare npm ->
// esm.sh, asset -> stub, existing local file -> disk, missing local file -> stub module.
func resolveLocalImport(ctx context.Context, tmp string, args api.OnResolveArgs) api.OnResolveResult {
	p := args.Path

	// Entry point is always a local on-disk file (e.g. "src/main.tsx").
	if args.Kind == api.ResolveEntryPoint {
		if abs, ok := resolveLocalFile(tmp, args.ResolveDir, p); ok {
			return api.OnResolveResult{Path: abs, Namespace: "file"}
		}

		return api.OnResolveResult{}
	}

	isLocal := strings.HasPrefix(p, ".") || strings.HasPrefix(p, "/") || strings.HasPrefix(p, "@/")

	// Bare npm specifier (incl. @scope/pkg).
	if !isLocal {
		if isAssetPath(p) {
			return api.OnResolveResult{Path: p, Namespace: nsAssetStub}
		}

		return api.OnResolveResult{Path: esmCDN + p + "?target=es2020", Namespace: "http-url"}
	}

	// Local asset import -> stub (esbuild has no loader for these here).
	if isAssetPath(p) {
		return api.OnResolveResult{Path: p, Namespace: nsAssetStub}
	}

	// Existing local source file -> let esbuild load it from disk.
	if abs, ok := resolveLocalFile(tmp, args.ResolveDir, p); ok {
		return api.OnResolveResult{Path: abs, Namespace: "file"}
	}

	// Hallucinated/missing local import -> stub, log, keep building.
	logFrom(ctx).WarnContext(ctx, "preview stubbed missing import",
		"import", p, "importer", filepath.Base(args.Importer))

	return api.OnResolveResult{Path: p, Namespace: nsStubModule}
}

// resolveLocalFile mirrors esbuild's local resolution (extension + index probing) for
// relative ("./"), root-absolute ("/") and "@/" alias paths rooted at the temp project.
func resolveLocalFile(tmp, resolveDir, p string) (string, bool) {
	var base, sub string
	switch {
	case strings.HasPrefix(p, "@/"):
		base, sub = filepath.Join(tmp, "src"), p[2:]
	case strings.HasPrefix(p, "/"):
		base, sub = tmp, strings.TrimPrefix(p, "/")
	default:
		base, sub = resolveDir, p
	}
	cand := filepath.Join(base, filepath.FromSlash(sub))
	exts := []string{"", ".tsx", ".ts", ".jsx", ".js", ".mjs", ".cjs", ".json", ".css"}
	for _, e := range exts {
		if isRegularFile(cand + e) {
			return cand + e, true
		}
	}
	for _, e := range exts[1:] { // directory index
		if idx := filepath.Join(cand, "index"+e); isRegularFile(idx) {
			return idx, true
		}
	}

	return "", false
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)

	return err == nil && !info.IsDir()
}

func isImageExt(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".avif", ".bmp", ".ico", ".svg":
		return true
	}

	return false
}

func isAssetPath(p string) bool {
	ext := strings.ToLower(filepath.Ext(p))
	if isImageExt(ext) {
		return true
	}
	switch ext {
	case ".woff", ".woff2", ".ttf", ".otf", ".eot",
		".mp4", ".webm", ".mov", ".mp3", ".wav", ".ogg", ".pdf":
		return true
	}

	return false
}

// stubModuleJS builds a CommonJS module whose every export (default, named, namespace)
// is safe: defaultExpr drives the default export; any other access returns a null React
// component. Using CJS + Proxy avoids esbuild "no matching export" errors for arbitrary
// named imports from a hallucinated module.
func stubModuleJS(defaultExpr string) string {
	return `var __c = function(){ return null; };
var __d = ` + defaultExpr + `;
module.exports = new Proxy(__c, {
  get: function(t, p){
    if (p === 'default') return __d;
    if (p === '__esModule') return true;
    if (typeof p === 'symbol') return t[p];
    return __c;
  }
});
`
}

func fetchURL(rawURL string) ([]byte, error) {
	if cached, ok := httpFetchCache.Load(rawURL); ok {
		if b, valid := cached.([]byte); valid {
			return b, nil
		}
	}
	resp, err := previewHTTP.Get(rawURL) //nolint:noctx // build-time fetch, client has timeout
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, wrapper.Wrapf(ErrPreviewFetch, "%s: status %d", rawURL, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	httpFetchCache.Store(rawURL, body)

	return body, nil
}

// isPreviewSkippedFile drops AI-generated infra/config that the server build replaces
// or doesn't need (matches the frontend's INFRA_FILES intent), plus binary assets.
func isPreviewSkippedFile(rel string) bool {
	base := strings.ToLower(filepath.Base(rel))
	switch base {
	case "package.json", "vite.config.ts", "vite.config.js",
		"tsconfig.json", "tsconfig.node.json",
		"postcss.config.js", "postcss.config.cjs",
		"tailwind.config.js", "tailwind.config.ts", "index.html":
		return true
	}
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".avif", ".bmp", ".ico",
		".woff", ".woff2", ".ttf", ".otf", ".eot", ".mp4", ".webm", ".mov",
		".mp3", ".wav", ".ogg", ".pdf", ".zip":
		return true
	}

	return false
}

// detectPreviewEntry mirrors the frontend entry detection: prefer conventional mount
// files, then any file that actually mounts React.
func detectPreviewEntry(files map[string]string) string {
	candidates := []string{
		"src/main.tsx", "src/main.jsx", "src/main.ts",
		"src/index.tsx", "src/index.jsx",
		"src/App.tsx", "src/App.jsx",
		"index.tsx", "index.jsx", "App.tsx",
	}
	for _, c := range candidates {
		if _, ok := files[c]; ok {
			return c
		}
	}
	// Fallback: first file that mounts React.
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if strings.HasSuffix(k, ".tsx") || strings.HasSuffix(k, ".jsx") {
			if strings.Contains(files[k], "createRoot") || strings.Contains(files[k], "ReactDOM.render") {
				return k
			}
		}
	}

	return ""
}

// previewErrorHTML renders a self-contained, readable error overlay (Vite-style) that
// is served with HTTP 200 inside the preview iframe. This keeps a broken/hallucinated
// build from returning 422 (which forced the frontend onto the OOM-prone Sandpack
// runtime) and instead shows exactly what failed — without losing the project code.
func previewErrorHTML(title, detail string) string {
	return `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Preview build error</title>
    <style>
      html,body{height:100%;margin:0;background:#0b0b0f;color:#e5e7eb;
        font:14px/1.6 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}
      .wrap{max-width:880px;margin:0 auto;padding:40px 24px}
      .badge{display:inline-flex;align-items:center;gap:8px;color:#fca5a5;
        background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.3);
        border-radius:8px;padding:6px 12px;font-weight:600;font-size:12px}
      h1{font-size:18px;margin:18px 0 8px;color:#f9fafb}
      p{color:#9ca3af;margin:0 0 16px}
      pre{white-space:pre-wrap;word-break:break-word;background:#15151c;
        border:1px solid #27272a;border-radius:10px;padding:16px;color:#fda4af;overflow:auto}
    </style>
  </head>
  <body>
    <div class="wrap">
      <span class="badge">Preview build failed</span>
      <h1>` + html.EscapeString(title) + `</h1>
      <p>Серверная сборка нашла ошибку в сгенерированном коде. Детали ниже — код проекта в безопасности.</p>
      <pre>` + html.EscapeString(detail) + `</pre>
    </div>
  </body>
</html>
`
}

// synthEntryRuntime is the body of the synthesized entry (after the imports). It wraps
// the generated root component in a React error boundary + global error/rejection
// handlers so a runtime crash paints a visible red overlay instead of a blank white
// screen, and renders an explicit error when no App export could be resolved. Written
// with createElement (no JSX) to avoid any jsx-runtime coupling in the synthesized file.
const synthEntryRuntime = `
function __overlay(title, detail) {
  return createElement("div", { style: {
    position: "fixed", inset: "0", zIndex: 2147483647, overflow: "auto",
    background: "#0b0b0f", color: "#e5e7eb", padding: "40px 24px",
    font: "14px/1.6 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
  } }, [
    createElement("div", { key: "b", style: { display: "inline-flex", gap: "8px",
      color: "#fca5a5", background: "rgba(239,68,68,.1)", border: "1px solid rgba(239,68,68,.3)",
      borderRadius: "8px", padding: "6px 12px", fontWeight: 600, fontSize: "12px" } }, "Runtime error"),
    createElement("h1", { key: "h", style: { fontSize: "18px", margin: "18px 0 8px",
      color: "#f9fafb" } }, title),
    createElement("pre", { key: "p", style: { whiteSpace: "pre-wrap", wordBreak: "break-word",
      background: "#15151c", border: "1px solid #27272a", borderRadius: "10px", padding: "16px",
      color: "#fda4af" } }, detail || "(no details)")
  ]);
}

class __ErrorBoundary extends Component {
  constructor(p) { super(p); this.state = { error: null }; }
  static getDerivedStateFromError(error) { return { error: error }; }
  componentDidCatch(error, info) { console.error("[preview] runtime error:", error, info); }
  render() {
    if (this.state.error) {
      var e = this.state.error;
      return __overlay(String((e && e.message) || e), String((e && e.stack) || ""));
    }
    return this.props.children;
  }
}

var __App = (__AppModule.default || __AppModule.App || Object.values(__AppModule)[0]);
var __el = document.getElementById("root");
var __root = __el ? createRoot(__el) : null;

function __paintFatal(title, detail) {
  if (__root) { __root.render(__overlay(title, detail)); }
  else if (document.body) {
    document.body.innerHTML = '<pre style="color:#fda4af;padding:24px;font:14px/1.6 monospace;white-space:pre-wrap">' + title + '\n\n' + detail + '</pre>';
  }
}

window.addEventListener("error", function (ev) {
  __paintFatal(String((ev && ev.message) || "Script error"), String((ev && ev.error && ev.error.stack) || ""));
});
window.addEventListener("unhandledrejection", function (ev) {
  var r = ev && ev.reason;
  __paintFatal(String((r && r.message) || r || "Unhandled promise rejection"), String((r && r.stack) || ""));
});

if (!__root) {
  if (document.body) {
    document.body.innerHTML = '<pre style="color:#fda4af;padding:24px;font:14px/1.6 monospace">Error: #root element not found</pre>';
  }
} else if (!__App) {
  __root.render(__overlay("Could not find exported App component to mount",
    "synthesizePreviewEntry resolved no default / named (App) / first export from the root module. Make sure the root component is exported."));
} else {
  __root.render(createElement(__ErrorBoundary, null, createElement(__App)));
}
`

// synthesizePreviewEntry builds a valid React entry file when the generated project
// has a root component (App / a default-exporting page) but no mount file. It returns
// the synthesized path ("<dir>/__istok_main.tsx"), its content, and ok=false if no
// mountable component could be found. The import is resilient: default OR named OR
// first export is used, so a slightly-off App export shape still renders.
func synthesizePreviewEntry(files map[string]string) (string, string, bool) {
	comp := findRootComponentFile(files)
	if comp == "" {
		return "", "", false
	}
	dir := path.Dir(comp)
	if dir == "." {
		dir = ""
	}
	base := strings.TrimSuffix(filepath.Base(comp), filepath.Ext(comp))
	importPath := "./" + base

	// Import the global stylesheet only if it actually exists (avoid a stub import).
	cssImport := ""
	for _, css := range []string{"src/index.css", "src/App.css", "index.css", "src/styles.css"} {
		if _, ok := files[css]; ok {
			rel, err := filepath.Rel(dir, css)
			if err == nil {
				slash := filepath.ToSlash(rel)
				if !strings.HasPrefix(slash, ".") {
					slash = "./" + slash // bare path would be read as an npm specifier
				}
				cssImport = "import \"" + slash + "\";\n"
			}

			break
		}
	}

	content := "import { createRoot } from \"react-dom/client\";\n" +
		"import { Component, createElement } from \"react\";\n" +
		"import * as __AppModule from \"" + importPath + "\";\n" +
		cssImport +
		synthEntryRuntime

	synthPath := "__istok_main.tsx"
	if dir != "" {
		synthPath = path.Join(dir, "__istok_main.tsx")
	}

	return synthPath, content, true
}

// findRootComponentFile locates the most likely root React component to mount when the
// project lacks an explicit entry: conventional App files first, then any .tsx/.jsx with
// a default export that isn't already a mount file.
func findRootComponentFile(files map[string]string) string {
	for _, c := range []string{"src/App.tsx", "src/App.jsx", "App.tsx", "App.jsx", "src/app.tsx"} {
		if _, ok := files[c]; ok {
			return c
		}
	}
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !strings.HasSuffix(k, ".tsx") && !strings.HasSuffix(k, ".jsx") {
			continue
		}
		if strings.Contains(files[k], "export default") {
			return k
		}
	}

	return ""
}

// wrapPreviewHTML embeds the bundled JS into a self-contained HTML page. Tailwind is
// loaded from our proxied Play CDN (same-origin to the user), not cdn.tailwindcss.com.
func wrapPreviewHTML(bundleJS string) string {
	var b strings.Builder
	b.WriteString(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Istok Preview</title>
    <script src="/api/v1/preview/tailwind.js"></script>
    <style>html,body,#root{height:100%;margin:0}</style>
  </head>
  <body>
    <div id="root"></div>
    <script type="module">
`)
	b.WriteString(bundleJS)
	b.WriteString(`
    </script>
  </body>
</html>
`)

	return b.String()
}
