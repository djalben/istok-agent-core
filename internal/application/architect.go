package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Architect (DefineArchitecture)
//  Claude Opus 4.6 → Full-Stack JSON Manifest
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

// defineArchitecture вызывает Claude Opus 4.6 для создания полной архитектурной схемы
// Это первый этап перед любой генерацией кода
func (o *Orchestrator) defineArchitecture(ctx context.Context, spec string, audit *ReverseEngineeringResult, features []CompetitorFeature) (*SystemManifest, error) {
	agent := o.agents[RoleBrain]
	ctx, cancel := context.WithTimeout(ctx, agent.Timeout)
	defer cancel()

	o.sendStatus(RoleBrain, "running", "🏗️ Claude Opus 4.6 проектирует архитектуру системы...", 15)

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

	prompt := fmt.Sprintf(`You are a world-class system architect. Design a COMPLETE full-stack architecture for this project.

SPECIFICATION:
%s%s%s

OUTPUT ONLY a valid JSON object with this EXACT structure (no markdown, no explanation, start with {):
{
  "project_name": "ProjectName",
  "type": "fullstack",
  "frontend": {
    "framework": "react",
    "styling": "tailwindcss",
    "pages": ["index.html", "dashboard.html"],
    "components": ["Navbar", "Sidebar", "Card"],
    "state_management": "zustand"
  },
  "backend": {
    "language": "go",
    "framework": "fiber",
    "modules": ["auth", "api-router", "db-connect"],
    "endpoints": [
      {"method": "POST", "path": "/api/auth/login", "handler": "AuthLogin", "auth": false, "description": "User login"},
      {"method": "GET", "path": "/api/users/me", "handler": "GetProfile", "auth": true, "description": "Get user profile"}
    ],
    "middleware": ["cors", "jwt-auth", "rate-limit", "logging"]
  },
  "database": {
    "engine": "postgresql",
    "tables": [
      {
        "name": "users",
        "columns": [
          {"name": "id", "type": "UUID", "primary_key": true},
          {"name": "email", "type": "VARCHAR(255)"},
          {"name": "password_hash", "type": "VARCHAR(255)"},
          {"name": "created_at", "type": "TIMESTAMP"}
        ]
      }
    ],
    "indexes": ["CREATE INDEX idx_users_email ON users(email)"]
  },
  "features": [
    {
      "name": "Authentication",
      "description": "JWT-based auth with login/register/refresh",
      "priority": "critical",
      "endpoints": ["/api/auth/login", "/api/auth/register"],
      "frontend": ["LoginPage", "RegisterPage"]
    }
  ],
  "file_map": [
    "index.html",
    "dashboard.html",
    "backend/main.go",
    "backend/handlers/auth.go",
    "backend/middleware/jwt.go",
    "backend/db/connect.go",
    "backend/db/migrations.sql"
  ]
}

RULES:
- Design for a REAL production system, not a toy
- Include ALL necessary tables, endpoints, and components
- Be specific with column types, HTTP methods, and auth requirements
- The file_map must list EVERY file that needs to be generated
- Adapt the architecture to match the project specification precisely`, spec, auditCtx, featureCtx)

	result, err := o.callLLMWithReasoning(ctx, agent.Model,
		"You are a world-class system architect. Output pure JSON only. Design production-grade architectures.",
		prompt, 8192, agent.ThinkingBudget)

	if err != nil {
		errMsg := fmt.Sprintf("⚠️ Architect fallback: %v", err)
		log.Printf("%s", errMsg)
		if len(errMsg) > 200 {
			errMsg = errMsg[:200]
		}
		o.sendStatus(RoleBrain, "error", errMsg, 20)
		return o.defaultManifest(spec, features), nil
	}

	manifest := o.parseManifest(result, spec, features)
	o.sendStatus(RoleBrain, "completed",
		fmt.Sprintf("✅ Архитектура: %d endpoints, %d tables, %d files",
			len(manifest.Backend.Endpoints), len(manifest.Database.Tables), len(manifest.FileMap)), 25)

	log.Printf("✅ Architect: manifest ready — %d endpoints, %d tables, %d features, %d files",
		len(manifest.Backend.Endpoints), len(manifest.Database.Tables), len(manifest.Features), len(manifest.FileMap))
	return manifest, nil
}

// parseManifest парсит JSON-манифест от Claude Opus
func (o *Orchestrator) parseManifest(content, spec string, features []CompetitorFeature) *SystemManifest {
	// Strip thinking blocks
	for strings.Contains(content, "<thinking>") {
		start := strings.Index(content, "<thinking>")
		end := strings.Index(content, "</thinking>")
		if end == -1 {
			break
		}
		content = content[:start] + content[end+len("</thinking>"):]
	}

	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	if first := strings.Index(content, "{"); first != -1 {
		if last := strings.LastIndex(content, "}"); last > first {
			content = content[first : last+1]
		}
	}

	var manifest SystemManifest
	if err := json.Unmarshal([]byte(content), &manifest); err != nil {
		log.Printf("⚠️ parseManifest JSON error: %v", err)
		return o.defaultManifest(spec, features)
	}

	manifest.CreatedAt = time.Now()
	if manifest.ProjectName == "" {
		manifest.ProjectName = "IstokProject"
	}
	return &manifest
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
