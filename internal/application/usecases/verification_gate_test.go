package usecases_test

import (
	"strings"
	"testing"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
)

// TestBackfillMissingImports verifies that unresolved local imports get stub files
// exporting the exact symbols importers reference — guaranteeing the bundle resolves.
func TestBackfillMissingImports(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"src/main.tsx": `import App from './App';
import { Hero } from '@/components/Hero';
import Button from '@/components/ui/Button';
import { cn, formatDate } from '@/lib/utils';
import './index.css';
import React from 'react';`,
		"src/App.tsx": `export default function App() { return null; }`,
		// Hero, ui/Button, lib/utils are MISSING — must be stubbed.
	}

	created := usecases.BackfillMissingImports(files)

	// react (node_module), ./App (exists), ./index.css (css) must NOT be stubbed.
	for _, p := range created {
		if strings.Contains(p, "react") || strings.Contains(p, "App") || strings.HasSuffix(p, ".css") {
			t.Errorf("must not stub %q", p)
		}
	}

	// Hero stub must export named Hero.
	hero, ok := files["src/components/Hero.tsx"]
	if !ok {
		t.Fatalf("Hero stub not created; got %v", created)
	}
	if !strings.Contains(hero, "export const Hero") {
		t.Errorf("Hero stub missing named export:\n%s", hero)
	}

	// Button stub must have a default export.
	btn, ok := files["src/components/ui/Button.tsx"]
	if !ok {
		t.Fatalf("Button stub not created; got %v", created)
	}
	if !strings.Contains(btn, "export default") {
		t.Errorf("Button stub missing default export:\n%s", btn)
	}

	// utils stub must export both named symbols.
	utils, ok := files["src/lib/utils.tsx"]
	if !ok {
		t.Fatalf("utils stub not created; got %v", created)
	}
	if !strings.Contains(utils, "export const cn") || !strings.Contains(utils, "export const formatDate") {
		t.Errorf("utils stub missing named exports:\n%s", utils)
	}
}

// TestBackfillNoFalsePositives ensures a fully-resolved project gets zero stubs.
func TestBackfillNoFalsePositives(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"src/main.tsx":            `import { Hero } from '@/components/Hero';`,
		"src/components/Hero.tsx": `export const Hero = () => null;`,
	}
	if created := usecases.BackfillMissingImports(files); len(created) != 0 {
		t.Errorf("expected no stubs for resolved project, got %v", created)
	}
}

// TestBackfillSkipsTypeOnly ensures type-only imports (stripped by bundler) aren't stubbed.
func TestBackfillSkipsTypeOnly(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"src/main.tsx": `import type { User } from '@/types/user';`,
	}
	if created := usecases.BackfillMissingImports(files); len(created) != 0 {
		t.Errorf("type-only import must not be stubbed, got %v", created)
	}
}
