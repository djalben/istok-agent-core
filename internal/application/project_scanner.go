package application

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ProjectScanner — сканирует package.json и tsconfig.json
//  чтобы агенты знали точные версии библиотек и алиасы.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ProjectEnv — результат сканирования проектных конфигов.
// Передаётся агентам как часть контекста перед каждой итерацией.
type ProjectEnv struct {
	// package.json
	PackageName    string            `json:"package_name,omitempty"`
	Dependencies   map[string]string `json:"dependencies,omitempty"`
	DevDeps        map[string]string `json:"dev_dependencies,omitempty"`
	Scripts        map[string]string `json:"scripts,omitempty"`
	PackageManager string            `json:"package_manager,omitempty"` // "bun" | "npm" | "pnpm"

	// tsconfig.json
	TSTarget       string            `json:"ts_target,omitempty"`
	TSModule       string            `json:"ts_module,omitempty"`
	TSPaths        map[string]string `json:"ts_paths,omitempty"` // "@/*" → "./src/*"
	TSStrict       bool              `json:"ts_strict,omitempty"`
	TSBaseURL      string            `json:"ts_base_url,omitempty"`
}

// ScanPackageJSON парсит содержимое package.json и извлекает зависимости и скрипты.
func ScanPackageJSON(content []byte) (*ProjectEnv, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("package.json is empty")
	}

	var pkg struct {
		Name            string            `json:"name"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
		PackageManager  string            `json:"packageManager"`
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}

	env := &ProjectEnv{
		PackageName:  pkg.Name,
		Dependencies: pkg.Dependencies,
		DevDeps:      pkg.DevDependencies,
		Scripts:      pkg.Scripts,
	}

	// Detect package manager
	switch {
	case strings.HasPrefix(pkg.PackageManager, "bun"):
		env.PackageManager = "bun"
	case strings.HasPrefix(pkg.PackageManager, "pnpm"):
		env.PackageManager = "pnpm"
	case strings.HasPrefix(pkg.PackageManager, "npm"):
		env.PackageManager = "npm"
	default:
		env.PackageManager = "npm"
	}

	return env, nil
}

// ScanTSConfig парсит содержимое tsconfig.json и извлекает пути, таргет, strict.
func ScanTSConfig(content []byte) (*ProjectEnv, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("tsconfig.json is empty")
	}

	var tsconfig struct {
		CompilerOptions struct {
			Target  string                `json:"target"`
			Module  string                `json:"module"`
			BaseURL string                `json:"baseUrl"`
			Strict  bool                  `json:"strict"`
			Paths   map[string][]string   `json:"paths"`
		} `json:"compilerOptions"`
	}

	if err := json.Unmarshal(content, &tsconfig); err != nil {
		return nil, fmt.Errorf("parse tsconfig.json: %w", err)
	}

	env := &ProjectEnv{
		TSTarget:  tsconfig.CompilerOptions.Target,
		TSModule:  tsconfig.CompilerOptions.Module,
		TSBaseURL: tsconfig.CompilerOptions.BaseURL,
		TSStrict:  tsconfig.CompilerOptions.Strict,
	}

	// Flatten paths: "@/*" → ["./src/*"] → "@/*" → "./src/*"
	if len(tsconfig.CompilerOptions.Paths) > 0 {
		env.TSPaths = make(map[string]string)
		for alias, targets := range tsconfig.CompilerOptions.Paths {
			if len(targets) > 0 {
				env.TSPaths[alias] = targets[0]
			}
		}
	}

	return env, nil
}

// ProjectScanner сканирует package.json и tsconfig.json (как raw bytes)
// и объединяет результат в единый ProjectEnv для передачи агентам.
func ProjectScanner(packageJSON, tsconfigJSON []byte) *ProjectEnv {
	env := &ProjectEnv{}

	if pkg, err := ScanPackageJSON(packageJSON); err == nil && pkg != nil {
		env.PackageName = pkg.PackageName
		env.Dependencies = pkg.Dependencies
		env.DevDeps = pkg.DevDeps
		env.Scripts = pkg.Scripts
		env.PackageManager = pkg.PackageManager
	}

	if ts, err := ScanTSConfig(tsconfigJSON); err == nil && ts != nil {
		env.TSTarget = ts.TSTarget
		env.TSModule = ts.TSModule
		env.TSBaseURL = ts.TSBaseURL
		env.TSStrict = ts.TSStrict
		env.TSPaths = ts.TSPaths
	}

	return env
}

// ForPrompt генерирует компактный текстовый блок для вставки в LLM промпт.
func (env *ProjectEnv) ForPrompt() string {
	if env == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n## PROJECT ENVIRONMENT (scanned from config files)\n")

	if env.PackageName != "" {
		b.WriteString(fmt.Sprintf("Package: %s (manager: %s)\n", env.PackageName, env.PackageManager))
	}

	if len(env.Dependencies) > 0 {
		b.WriteString("Dependencies:\n")
		for pkg, ver := range env.Dependencies {
			b.WriteString(fmt.Sprintf("  %s: %s\n", pkg, ver))
		}
	}

	if len(env.DevDeps) > 0 {
		b.WriteString("DevDependencies:\n")
		for pkg, ver := range env.DevDeps {
			b.WriteString(fmt.Sprintf("  %s: %s\n", pkg, ver))
		}
	}

	if env.TSTarget != "" || env.TSModule != "" {
		b.WriteString(fmt.Sprintf("TypeScript: target=%s module=%s strict=%v\n", env.TSTarget, env.TSModule, env.TSStrict))
	}

	if len(env.TSPaths) > 0 {
		b.WriteString("Path Aliases:\n")
		for alias, target := range env.TSPaths {
			b.WriteString(fmt.Sprintf("  %s → %s\n", alias, target))
		}
	}

	return b.String()
}
