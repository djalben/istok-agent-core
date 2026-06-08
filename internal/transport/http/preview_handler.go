package http

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
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
	cache sync.Map // sessionID -> *previewCacheEntry
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

	html, err := buildPreviewHTML(files)
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
func buildPreviewHTML(files map[string]string) (string, error) {
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
		return "", wrapper.Wrap(ErrPreviewNoEntry)
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
		Alias:   map[string]string{"@": filepath.Join(tmp, "src")},
		Plugins: []api.Plugin{cdnResolverPlugin()},
	})
	if len(result.Errors) > 0 {
		msgs := api.FormatMessages(result.Errors, api.FormatMessagesOptions{})

		return "", wrapper.Wrapf(ErrPreviewBundle, "%s", strings.Join(msgs, "; "))
	}
	if len(result.OutputFiles) == 0 {
		return "", wrapper.Wrap(ErrPreviewNoOutput)
	}

	var bundle strings.Builder
	for _, out := range result.OutputFiles {
		if strings.HasSuffix(out.Path, ".js") {
			bundle.Write(out.Contents)
		}
	}

	return wrapPreviewHTML(bundle.String()), nil
}

// cdnResolverPlugin resolves bare npm specifiers to esm.sh URLs and fetches their
// contents (and transitive https imports) at build time, plus neutralizes CSS imports
// (Tailwind utilities are supplied by the runtime Play CDN, so app CSS is dropped).
func cdnResolverPlugin() api.Plugin {
	return api.Plugin{
		Name: "istok-cdn-resolver",
		Setup: func(build api.PluginBuild) {
			// Drop CSS imported from local files — Play CDN handles Tailwind utilities.
			build.OnLoad(api.OnLoadOptions{Filter: `\.css$`, Namespace: "file"},
				func(_ api.OnLoadArgs) (api.OnLoadResult, error) {
					empty := ""

					return api.OnLoadResult{Contents: &empty, Loader: api.LoaderJS}, nil
				})

			// Bare npm specifiers from on-disk files -> esm.sh.
			build.OnResolve(api.OnResolveOptions{Filter: `.*`, Namespace: "file"},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					p := args.Path
					if strings.HasPrefix(p, ".") || strings.HasPrefix(p, "/") || strings.HasPrefix(p, "@/") {
						return api.OnResolveResult{}, nil // local — default resolution
					}

					return api.OnResolveResult{
						Path:      esmCDN + p + "?target=es2020",
						Namespace: "http-url",
					}, nil
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

			// Fetch remote module contents (cached across builds).
			build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "http-url"},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					body, err := fetchURL(args.Path)
					if err != nil {
						return api.OnLoadResult{}, wrapper.Wrap(err)
					}
					contents := string(body)
					loader := api.LoaderJS
					if strings.Contains(args.Path, ".css") {
						loader = api.LoaderCSS
					}

					return api.OnLoadResult{Contents: &contents, Loader: loader}, nil
				})
		},
	}
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
