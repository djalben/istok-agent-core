package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/istok/agent-core/internal/application/usecases"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Architect (DefineArchitecture)
//  Ядро Истока → Full-Stack JSON Manifest
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// SystemManifest полная архитектурная схема системы
type SystemManifest struct {
	ProjectName string           `json:"project_name"`
	Type        string           `json:"type"` // "fullstack" | "frontend" | "api"
	Frontend    FrontendManifest `json:"frontend"`
	Backend     BackendManifest  `json:"backend"`
	Database    DatabaseManifest `json:"database"`
	Features    []FeatureSpec    `json:"features"`
	FileMap     []string         `json:"file_map"`
	CreatedAt   time.Time        `json:"created_at"`
}

// FrontendManifest описание фронтенда
type FrontendManifest struct {
	Framework       string   `json:"framework"`        // "react" | "vue" | "vanilla" | "nextjs"
	Styling         string   `json:"styling"`          // "tailwindcss" | "css-modules"
	Pages           []string `json:"pages"`            // ["index.html", "dashboard.html", "auth.html"]
	Components      []string `json:"components"`       // ["Navbar", "Sidebar", "Card", "Modal"]
	StateManagement string   `json:"state_management"` // "zustand" | "context" | "redux"
}

// BackendManifest описание бэкенда
type BackendManifest struct {
	Language   string         `json:"language"`  // "go" | "node" | "python"
	Framework  string         `json:"framework"` // "fiber" | "gin" | "echo" | "express"
	Modules    []string       `json:"modules"`   // ["auth", "api-router", "db-connect", "payments"]
	Endpoints  []EndpointSpec `json:"endpoints"`
	Middleware []string       `json:"middleware"` // ["cors", "jwt-auth", "rate-limit", "logging"]
}

// EndpointSpec описание API-эндпоинта
type EndpointSpec struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Handler     string `json:"handler"`
	Auth        bool   `json:"auth"`
	Description string `json:"description"`
}

// DatabaseManifest описание базы данных
type DatabaseManifest struct {
	Engine  string      `json:"engine"` // "postgresql" | "sqlite" | "mysql"
	Tables  []TableSpec `json:"tables"`
	Indexes []string    `json:"indexes"`
}

// TableSpec описание таблицы БД
type TableSpec struct {
	Name    string       `json:"name"`
	Columns []ColumnSpec `json:"columns"`
}

// ColumnSpec описание колонки БД
type ColumnSpec struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	PrimaryKey bool   `json:"primary_key,omitempty"`
	Nullable   bool   `json:"nullable,omitempty"`
	Reference  string `json:"reference,omitempty"` // "users.id"
}

// FeatureSpec описание фичи системы
type FeatureSpec struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"` // "critical" | "high" | "medium"
	Endpoints   []string `json:"endpoints"`
	Frontend    []string `json:"frontend"`
}

// defineArchitecture вызывает ядро Истока для создания полной архитектурной схемы
// Это первый этап перед любой генерацией кода
func (o *Orchestrator) defineArchitecture(ctx context.Context, spec string, audit *ReverseEngineeringResult, features []CompetitorFeature) (*SystemManifest, error) {
	agent := o.agents[RoleBrain]
	ctx, cancel := context.WithTimeout(ctx, agent.Timeout)
	defer cancel()

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("--- DEBUG: ЗАПУСК АРХИТЕКТОРА ---\n")
	fmt.Printf("Spec: %s\n", spec)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	o.sendStatus(RoleArchitect, "running", "🏗️ Архитектор проектирует систему...", 15)

	// ── Phase 0: Reflective Reasoning (Thought Chain) ──
	reflectionTask := fmt.Sprintf("Design full-stack architecture for: %s", spec[:min(len(spec), 300)])
	thoughtChain, _ := o.ThoughtChain(ctx, RoleBrain, reflectionTask)
	reflectionCtx := ThoughtChainContext(thoughtChain)

	// Build feature context if synthesis produced features
	featureCtx := ""
	if len(features) > 0 {
		var featureLines []string
		for _, f := range features {
			featureLines = append(featureLines, fmt.Sprintf("- [%s] %s: %s", f.Priority, f.Name, f.Description))
		}
		featureCtx = fmt.Sprintf("\n\nFEATURES FROM COMPETITOR ANALYSIS:\n%s", strings.Join(featureLines, "\n"))
	}

	// Build audit context
	auditCtx := ""
	if audit != nil {
		auditCtx = fmt.Sprintf("\n\nDESIGN AUDIT:\n- Colors: %v\n- Components: %v\n- Technologies: %v\n- Layout: %s",
			audit.Colors, audit.Components, audit.Technologies, audit.Layout)
	}

	// Inject ProjectScanner context (exact library versions from package.json/tsconfig.json)
	envCtx := ""
	o.mu.RLock()
	if o.projectEnv != nil {
		envCtx = o.projectEnv.ForPrompt()
	}
	o.mu.RUnlock()

	prompt := fmt.Sprintf(`Design a full-stack architecture with FUNCTIONAL requirements. Output ONLY valid JSON, no markdown.

SPEC: %s%s%s%s%s

LOVABLE KNOWLEDGE BASE (mandatory stack):
- Bundler: Vite 5 (with Bun as package manager)
- Framework: React 18 + TypeScript
- Routing: @tanstack/react-router (file-based routes in src/routes/)
- Data fetching: @tanstack/react-query (hooks in src/hooks/)
- UI library: shadcn/ui (Radix primitives + Tailwind)
- Styling: TailwindCSS 3 with @/* import aliases
- Icons: lucide-react
- Forms: react-hook-form + zod validation
- State: zustand (lightweight) or TanStack Query cache

MANDATORY DIRECTORY STRUCTURE:
  src/
    components/ui/       — shadcn primitives (Button, Card, Dialog, etc.)
    components/layout/   — AppLayout, Sidebar, Header, MobileNav
    hooks/               — useAuth, useProducts, useMutation wrappers
    services/            — API client, auth service, storage helpers
    routes/              — TanStack Router file-based routes
    lib/                 — utils.ts (cn helper), constants
    types/               — shared TypeScript interfaces

IMPORT RULE: ALL imports MUST use @/* aliases (e.g. import { Button } from "@/components/ui/button").
Never use relative paths like "../../components".

JSON keys: project_name, type("fullstack"), frontend{framework,styling,pages[],components[],state_management}, backend{language,framework,modules[],endpoints[{method,path,handler,auth,description}],middleware[]}, database{engine,tables[{name,columns[{name,type,primary_key}]}],indexes[]}, features[{name,description,priority,endpoints[],frontend[]}], file_map[].

CRITICAL: Each feature MUST include concrete frontend interactivity:
- Forms with validation logic (what fields, what validation rules, zod schemas)
- Business logic (cart calculation, order total, quantity controls)
- Data structures (menu items with name/price/category, products with filters)
- User interactions (add to cart, submit order, toggle menu, smooth scroll)
- TanStack Query hooks for data fetching with optimistic updates

Example feature:
{"name":"Order System","description":"Menu with categories, Add to Cart with quantity, cart sidebar with +/- controls, order total calculation, checkout form with react-hook-form+zod validation, TanStack Query mutations, zustand cart store","priority":"critical","endpoints":["/api/orders"],"frontend":["MenuGrid","CartSidebar","CheckoutForm","OrderConfirmation"]}

Be production-grade. Start with {.`, spec, auditCtx, featureCtx, envCtx, reflectionCtx)

	result, err := o.callLLMWithReasoning(ctx, agent.Model,
		`You are a senior system architect with deep expertise in the Lovable/shadcn stack.
Design architectures with FUNCTIONAL specifications — every component must have clear interactivity and business logic requirements.
KNOWLEDGE BASE:
- Stack: Vite 5, Bun, React 18+TS, TanStack Router+Query, shadcn/ui, TailwindCSS, lucide-react
- Imports: ONLY @/* aliases. Never relative paths.
- Structure: components/ui, components/layout, hooks, services, routes, lib, types
- All external deps through interfaces (ports pattern).

## REFLECTION BLOCK (execute BEFORE outputting JSON):
Before producing your final manifest, verify against these 3 critical errors:
1. [ENTITY INTEGRITY] — Does every DB table have proper primary keys, timestamps (created_at/updated_at), and foreign key references? Fix missing relationships.
2. [FILE MAP COMPLETENESS] — Does file_map include ALL files needed to implement ALL features? Cross-check: every endpoint needs a handler, every page needs a route file, every component needs its file.
3. [IMPORT CONSISTENCY] — Will all components be importable via @/* aliases? Verify directory structure matches import paths.
If any check fails, silently correct the manifest before output.

Output pure JSON only.`,
		prompt, 16384)

	if err != nil {
		errMsg := fmt.Sprintf("⚠️ Architect fallback: %v", err)
		log.Printf("DEBUG [Architect] LLM call FAILED: %v", err)
		log.Printf("%s", errMsg)
		if len(errMsg) > 200 {
			errMsg = errMsg[:200]
		}
		o.sendStatus(RoleArchitect, "error", errMsg, 20)
		fallback := o.defaultManifest(spec, features)
		expandFileMap(fallback)
		log.Printf("📂 Architect fallback: expanded FileMap to %d files", len(fallback.FileMap))
		return fallback, nil
	}

	// DEBUG: FULL raw Architect LLM output (ADR / Architectural Decision Record)
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("--- DEBUG: ОТВЕТ АРХИТЕКТОРА (raw, %d chars) ---\n", len(result))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("%s\n", result)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	log.Printf("DEBUG [Architect] raw LLM output: %d chars", len(result))

	manifest := o.parseManifest(result, spec, features)

	// ── Post-parse: expand FileMap from manifest structure ──
	preExpand := len(manifest.FileMap)
	expandFileMap(manifest)
	if len(manifest.FileMap) > preExpand {
		log.Printf("📂 Architect FileMap expanded: %d → %d files", preExpand, len(manifest.FileMap))
	}

	// Print parsed ADR summary
	fmt.Printf("\n┌─── ARCHITECT ADR (Architectural Decision Record) ───┐\n")
	fmt.Printf("│ Project:    %s\n", manifest.ProjectName)
	fmt.Printf("│ Type:       %s\n", manifest.Type)
	fmt.Printf("│ Frontend:   %s + %s (state: %s)\n", manifest.Frontend.Framework, manifest.Frontend.Styling, manifest.Frontend.StateManagement)
	fmt.Printf("│ Backend:    %s / %s\n", manifest.Backend.Language, manifest.Backend.Framework)
	fmt.Printf("│ Database:   %s\n", manifest.Database.Engine)
	fmt.Printf("│ Pages:      %v\n", manifest.Frontend.Pages)
	fmt.Printf("│ Components: %v\n", manifest.Frontend.Components)
	fmt.Printf("│ Endpoints:  %d\n", len(manifest.Backend.Endpoints))
	for _, ep := range manifest.Backend.Endpoints {
		fmt.Printf("│   %s %s → %s\n", ep.Method, ep.Path, ep.Handler)
	}
	fmt.Printf("│ Tables:     %d\n", len(manifest.Database.Tables))
	for _, t := range manifest.Database.Tables {
		fmt.Printf("│   %s (%d cols)\n", t.Name, len(t.Columns))
	}
	fmt.Printf("│ Features:   %d\n", len(manifest.Features))
	for _, f := range manifest.Features {
		fmt.Printf("│   [%s] %s\n", f.Priority, f.Name)
	}
	fmt.Printf("│ FileMap:    %d files\n", len(manifest.FileMap))
	fmt.Printf("└──────────────────────────────────────────────────────┘\n\n")

	o.sendStatus(RoleArchitect, "completed",
		fmt.Sprintf("✅ Архитектура: %d endpoints, %d tables, %d files",
			len(manifest.Backend.Endpoints), len(manifest.Database.Tables), len(manifest.FileMap)), 100)

	log.Printf("✅ Architect: manifest ready — %d endpoints, %d tables, %d features, %d files",
		len(manifest.Backend.Endpoints), len(manifest.Database.Tables), len(manifest.Features), len(manifest.FileMap))
	return manifest, nil
}

// parseManifest парсит JSON-манифест от ядра Истока
func (o *Orchestrator) parseManifest(content, spec string, features []CompetitorFeature) *SystemManifest {
	jsonBlock, ok := usecases.ExtractFirstJSONObject(content)
	if !ok {
		log.Printf("⚠️ parseManifest: no JSON object found in response (len=%d)", len(content))
		return o.defaultManifest(spec, features)
	}

	var manifest SystemManifest
	if err := json.Unmarshal([]byte(jsonBlock), &manifest); err != nil {
		log.Printf("⚠️ parseManifest strict parse failed: %v — trying relaxed parse", err)
		// Relaxed parse: extract file_map and key fields from untyped map
		manifest = o.parseManifestRelaxed(jsonBlock, spec)
	}

	manifest.CreatedAt = time.Now()
	if manifest.ProjectName == "" {
		manifest.ProjectName = "IstokProject"
	}
	return &manifest
}

// parseManifestRelaxed extracts manifest data from a loosely-typed JSON map.
// Handles cases where LLM returns objects instead of strings for pages/components.
func (o *Orchestrator) parseManifestRelaxed(jsonBlock, spec string) SystemManifest {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonBlock), &raw); err != nil {
		log.Printf("⚠️ parseManifestRelaxed: even raw parse failed: %v", err)
		return *o.defaultManifest(spec, nil)
	}

	m := SystemManifest{
		ProjectName: getStringField(raw, "project_name"),
		Type:        getStringField(raw, "type"),
	}

	// Extract file_map (most critical field for chunked coder)
	if fm, ok := raw["file_map"]; ok {
		if arr, ok := fm.([]interface{}); ok {
			for _, item := range arr {
				if s, ok := item.(string); ok {
					m.FileMap = append(m.FileMap, s)
				}
			}
		}
	}

	// Extract frontend
	if fe, ok := raw["frontend"].(map[string]interface{}); ok {
		m.Frontend.Framework = getStringField(fe, "framework")
		m.Frontend.Styling = getStringField(fe, "styling")
		m.Frontend.StateManagement = getStringField(fe, "state_management")
		m.Frontend.Pages = extractStringArray(fe, "pages")
		m.Frontend.Components = extractStringArray(fe, "components")
	}

	// Extract backend
	if be, ok := raw["backend"].(map[string]interface{}); ok {
		m.Backend.Language = getStringField(be, "language")
		m.Backend.Framework = getStringField(be, "framework")
		m.Backend.Modules = extractStringArray(be, "modules")
		m.Backend.Middleware = extractStringArray(be, "middleware")
		// Endpoints — try to re-marshal and parse
		if eps, ok := be["endpoints"].([]interface{}); ok {
			for _, ep := range eps {
				if epMap, ok := ep.(map[string]interface{}); ok {
					m.Backend.Endpoints = append(m.Backend.Endpoints, EndpointSpec{
						Method:      getStringField(epMap, "method"),
						Path:        getStringField(epMap, "path"),
						Handler:     getStringField(epMap, "handler"),
						Auth:        getBoolField(epMap, "auth"),
						Description: getStringField(epMap, "description"),
					})
				}
			}
		}
	}

	// Extract database
	if db, ok := raw["database"].(map[string]interface{}); ok {
		m.Database.Engine = getStringField(db, "engine")
		if tables, ok := db["tables"].([]interface{}); ok {
			for _, t := range tables {
				if tMap, ok := t.(map[string]interface{}); ok {
					table := TableSpec{Name: getStringField(tMap, "name")}
					if cols, ok := tMap["columns"].([]interface{}); ok {
						for _, c := range cols {
							if cMap, ok := c.(map[string]interface{}); ok {
								table.Columns = append(table.Columns, ColumnSpec{
									Name:       getStringField(cMap, "name"),
									Type:       getStringField(cMap, "type"),
									PrimaryKey: getBoolField(cMap, "primary_key"),
								})
							}
						}
					}
					m.Database.Tables = append(m.Database.Tables, table)
				}
			}
		}
	}

	// Extract features
	if feats, ok := raw["features"].([]interface{}); ok {
		for _, f := range feats {
			if fMap, ok := f.(map[string]interface{}); ok {
				m.Features = append(m.Features, FeatureSpec{
					Name:        getStringField(fMap, "name"),
					Description: getStringField(fMap, "description"),
					Priority:    getStringField(fMap, "priority"),
				})
			}
		}
	}

	log.Printf("✅ parseManifestRelaxed: recovered %d files, %d endpoints, %d tables",
		len(m.FileMap), len(m.Backend.Endpoints), len(m.Database.Tables))
	return m
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBoolField(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func extractStringArray(m map[string]interface{}, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	var result []string
	for _, item := range arr {
		switch val := item.(type) {
		case string:
			result = append(result, val)
		case map[string]interface{}:
			// LLM returned object instead of string — extract name/path/component field
			for _, field := range []string{"path", "name", "component", "page"} {
				if s, ok := val[field].(string); ok {
					result = append(result, s)
					break
				}
			}
		}
	}
	return result
}

// expandFileMap enriches manifest's FileMap by synthesizing files from its structure.
// Ensures comprehensive coverage: every page, component, endpoint, table → gets files.
// Called after parsing to guarantee chunked generation threshold (≥5 files).
func expandFileMap(m *SystemManifest) {
	seen := make(map[string]bool, len(m.FileMap))
	for _, f := range m.FileMap {
		seen[f] = true
	}

	add := func(path string) {
		if path != "" && !seen[path] {
			seen[path] = true
			m.FileMap = append(m.FileMap, path)
		}
	}

	// ── Infrastructure files (always needed) ──
	infra := []string{
		"package.json", "vite.config.ts", "tsconfig.json", "tsconfig.node.json",
		"tailwind.config.ts", "postcss.config.js", "index.html",
		"src/main.tsx", "src/App.tsx", "src/index.css",
		"src/lib/utils.ts", "src/lib/constants.ts", "src/lib/api-client.ts",
		"src/types/index.ts",
		"src/components/ui/button.tsx", "src/components/ui/card.tsx",
		"src/components/ui/input.tsx", "src/components/ui/dialog.tsx",
		"src/components/ui/badge.tsx", "src/components/ui/toast.tsx",
		"src/components/ui/select.tsx", "src/components/ui/tabs.tsx",
		"src/components/ui/dropdown-menu.tsx", "src/components/ui/avatar.tsx",
		"src/components/ui/separator.tsx", "src/components/ui/sheet.tsx",
		"src/components/ui/tooltip.tsx", "src/components/ui/progress.tsx",
		"src/components/ui/skeleton.tsx", "src/components/ui/scroll-area.tsx",
		"src/components/ui/table.tsx", "src/components/ui/checkbox.tsx",
		"src/components/layout/AppLayout.tsx", "src/components/layout/Sidebar.tsx",
		"src/components/layout/Header.tsx", "src/components/layout/MobileNav.tsx",
		"src/components/layout/Footer.tsx",
	}
	for _, f := range infra {
		add(f)
	}

	// ── Pages → route files ──
	for _, page := range m.Frontend.Pages {
		name := strings.ToLower(strings.ReplaceAll(page, " ", "-"))
		name = strings.TrimSuffix(name, ".tsx")
		name = strings.TrimPrefix(name, "/")
		if name == "" || name == "index" || name == "index.html" {
			add("src/routes/index.tsx")
		} else {
			add(fmt.Sprintf("src/routes/%s.tsx", name))
		}
		// Each page gets a page component
		pascal := toPascalCase(page)
		if pascal != "" {
			add(fmt.Sprintf("src/pages/%s.tsx", pascal))
		}
	}

	// ── Components → component files ──
	for _, comp := range m.Frontend.Components {
		pascal := toPascalCase(comp)
		if pascal != "" {
			add(fmt.Sprintf("src/components/%s.tsx", pascal))
		}
	}

	// ── Endpoints → service files + hooks ──
	handlerModules := make(map[string]bool)
	for _, ep := range m.Backend.Endpoints {
		// Extract module name from path (e.g. /api/users → users)
		parts := strings.Split(strings.TrimPrefix(ep.Path, "/api/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			mod := strings.ToLower(parts[0])
			if !handlerModules[mod] {
				handlerModules[mod] = true
				add(fmt.Sprintf("src/services/%s-service.ts", mod))
				add(fmt.Sprintf("src/hooks/use-%s.ts", mod))
				add(fmt.Sprintf("src/types/%s.ts", mod))
			}
		}
	}

	// ── Backend modules → service files ──
	for _, mod := range m.Backend.Modules {
		name := strings.ToLower(strings.ReplaceAll(mod, " ", "-"))
		add(fmt.Sprintf("src/services/%s-service.ts", name))
	}

	// ── Tables → type definitions ──
	for _, table := range m.Database.Tables {
		name := strings.ToLower(table.Name)
		add(fmt.Sprintf("src/types/%s.ts", name))
	}

	// ── Features → dedicated feature folders ──
	for _, feat := range m.Features {
		name := strings.ToLower(strings.ReplaceAll(feat.Name, " ", "-"))
		if name == "" {
			continue
		}
		add(fmt.Sprintf("src/features/%s/index.tsx", name))
		add(fmt.Sprintf("src/features/%s/hooks.ts", name))
		add(fmt.Sprintf("src/features/%s/types.ts", name))
		// Feature frontend components
		for _, comp := range feat.Frontend {
			pascal := toPascalCase(comp)
			if pascal != "" {
				add(fmt.Sprintf("src/features/%s/components/%s.tsx", name, pascal))
			}
		}
	}

	// ── Common hooks ──
	commonHooks := []string{"use-auth", "use-theme", "use-toast", "use-debounce", "use-local-storage", "use-media-query"}
	for _, h := range commonHooks {
		add(fmt.Sprintf("src/hooks/%s.ts", h))
	}

	// ── Auth/RBAC (if spec mentions roles/auth) ──
	add("src/services/auth-service.ts")
	add("src/hooks/use-auth.ts")
	add("src/contexts/AuthContext.tsx")
	add("src/components/auth/LoginForm.tsx")
	add("src/components/auth/SignupForm.tsx")
	add("src/components/auth/ProtectedRoute.tsx")
	add("src/components/auth/RoleGuard.tsx")

	originalCount := len(m.FileMap) - len(seen) + len(m.FileMap) // approximate
	log.Printf("📂 expandFileMap: %d → %d files (synthesized from %d pages, %d components, %d endpoints, %d features)",
		originalCount, len(m.FileMap),
		len(m.Frontend.Pages), len(m.Frontend.Components),
		len(m.Backend.Endpoints), len(m.Features))
}

// toPascalCase converts "my component" or "my-component" to "MyComponent".
func toPascalCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Split on spaces, hyphens, underscores
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '/'
	})
	var b strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]))
			if len(p) > 1 {
				b.WriteString(p[1:])
			}
		}
	}
	return b.String()
}

// defaultManifest возвращает базовый манифест при ошибке
func (o *Orchestrator) defaultManifest(spec string, features []CompetitorFeature) *SystemManifest {
	m := &SystemManifest{
		ProjectName: "IstokProject",
		Type:        "fullstack",
		Frontend: FrontendManifest{
			Framework:       "vanilla",
			Styling:         "tailwindcss",
			Pages:           []string{"index.html"},
			Components:      []string{"Navbar", "Hero", "Features", "CTA", "Footer"},
			StateManagement: "vanilla",
		},
		Backend: BackendManifest{
			Language:  "go",
			Framework: "fiber",
			Modules:   []string{"auth", "api-router", "db-connect"},
			Endpoints: []EndpointSpec{
				{Method: "POST", Path: "/api/auth/login", Handler: "AuthLogin", Auth: false, Description: "User login"},
				{Method: "POST", Path: "/api/auth/register", Handler: "AuthRegister", Auth: false, Description: "User registration"},
				{Method: "GET", Path: "/api/users/me", Handler: "GetProfile", Auth: true, Description: "Get current user"},
			},
			Middleware: []string{"cors", "jwt-auth", "logging"},
		},
		Database: DatabaseManifest{
			Engine: "postgresql",
			Tables: []TableSpec{
				{
					Name: "users",
					Columns: []ColumnSpec{
						{Name: "id", Type: "UUID", PrimaryKey: true},
						{Name: "email", Type: "VARCHAR(255)"},
						{Name: "password_hash", Type: "VARCHAR(255)"},
						{Name: "created_at", Type: "TIMESTAMP"},
					},
				},
			},
		},
		FileMap:   []string{"index.html", "backend/main.go", "backend/handlers/auth.go", "backend/db/connect.go"},
		CreatedAt: time.Now(),
	}

	// Convert competitor features into FeatureSpecs
	for _, f := range features {
		m.Features = append(m.Features, FeatureSpec{
			Name:        f.Name,
			Description: f.Description,
			Priority:    f.Priority,
		})
	}

	return m
}
