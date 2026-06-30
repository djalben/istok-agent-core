package application

import (
	"strings"
	"testing"
)

// ─── fixRemovedLucideIcons ────────────────────────────────────────────────────

func TestFixRemovedLucideIcons_ReplacesKnown(t *testing.T) {
	files := map[string]string{
		"src/components/Footer.tsx": `import { Mail, Github, Linkedin, Twitter, ExternalLink } from 'lucide-react';

export function Footer() {
  return (
    <div>
      <Github size={16} />
      <Linkedin />
      <Twitter className="icon" />
    </div>
  );
}`,
	}

	n := fixRemovedLucideIcons(files)
	if n != 1 {
		t.Fatalf("expected 1 file modified, got %d", n)
	}

	result := files["src/components/Footer.tsx"]
	for _, removed := range []string{"Github", "Linkedin", "Twitter"} {
		if strings.Contains(result, removed) {
			t.Errorf("removed icon %q still present in result", removed)
		}
	}
	for _, expected := range []string{"Code2", "Globe", "Share2"} {
		if !strings.Contains(result, expected) {
			t.Errorf("replacement %q not found in result", expected)
		}
	}
	// Original icons should still be there
	if !strings.Contains(result, "Mail") {
		t.Error("Mail icon should not be removed")
	}
	if !strings.Contains(result, "ExternalLink") {
		t.Error("ExternalLink icon should not be removed")
	}
}

func TestFixRemovedLucideIcons_SkipsNonLucideFiles(t *testing.T) {
	original := `import { Github } from 'some-other-lib';`
	files := map[string]string{
		"src/components/Foo.tsx": original,
	}
	n := fixRemovedLucideIcons(files)
	if n != 0 {
		t.Fatalf("expected 0 files modified, got %d", n)
	}
	if files["src/components/Foo.tsx"] != original {
		t.Error("file was modified when it should not have been")
	}
}

func TestFixRemovedLucideIcons_NoChangeWhenClean(t *testing.T) {
	original := `import { Mail, Search, ArrowRight } from 'lucide-react';`
	files := map[string]string{
		"src/components/Nav.tsx": original,
	}
	n := fixRemovedLucideIcons(files)
	if n != 0 {
		t.Fatalf("expected 0 files modified, got %d", n)
	}
}

func TestFixRemovedLucideIcons_DeduplicatesAfterReplacement(t *testing.T) {
	// Code2 already imported; Github → Code2 would create duplicate
	files := map[string]string{
		"src/components/Icons.tsx": `import { Code2, Github } from 'lucide-react';
export const A = <Code2 />;
export const B = <Github />;`,
	}
	n := fixRemovedLucideIcons(files)
	if n != 1 {
		t.Fatalf("expected 1 file modified, got %d", n)
	}
	result := files["src/components/Icons.tsx"]
	// Should not have duplicate Code2 in import
	importLine := result[:strings.Index(result, "\n")]
	occurrences := strings.Count(importLine, "Code2")
	if occurrences != 1 {
		t.Errorf("expected exactly 1 Code2 in import line, got %d: %q", occurrences, importLine)
	}
}

func TestFixRemovedLucideIcons_SkipsNonTSXFiles(t *testing.T) {
	files := map[string]string{
		"README.md": `Use the Github icon from lucide-react`,
	}
	n := fixRemovedLucideIcons(files)
	if n != 0 {
		t.Fatalf("expected 0 files modified, got %d", n)
	}
}

// ─── fixHallucinatedShadcnExports ────────────────────────────────────────────

func TestFixHallucinatedShadcnExports_ReplacesHallucinatedName(t *testing.T) {
	files := map[string]string{
		"src/components/CTA.tsx": `import { ModernButton } from '@/components/ui/button';
import { PremiumCard } from '@/components/ui/card';

export function CTA() {
  return (
    <PremiumCard>
      <ModernButton>Click</ModernButton>
    </PremiumCard>
  );
}`,
	}

	n := fixHallucinatedShadcnExports(files)
	if n != 1 {
		t.Fatalf("expected 1 file modified, got %d", n)
	}

	result := files["src/components/CTA.tsx"]
	for _, bad := range []string{"ModernButton", "PremiumCard"} {
		if strings.Contains(result, bad) {
			t.Errorf("hallucinated name %q still present", bad)
		}
	}
	if !strings.Contains(result, "Button") {
		t.Error("canonical Button not found")
	}
	if !strings.Contains(result, "Card") {
		t.Error("canonical Card not found")
	}
}

func TestFixHallucinatedShadcnExports_LeavesValidExportsAlone(t *testing.T) {
	original := `import { Button, buttonVariants } from '@/components/ui/button';
import { Card, CardHeader, CardContent } from '@/components/ui/card';

export function Comp() { return <Card><Button>OK</Button></Card>; }`
	files := map[string]string{
		"src/components/Valid.tsx": original,
	}
	n := fixHallucinatedShadcnExports(files)
	if n != 0 {
		t.Fatalf("expected 0 files modified, got %d; content:\n%s", n, files["src/components/Valid.tsx"])
	}
}

func TestFixHallucinatedShadcnExports_HandlesUnknownComponent(t *testing.T) {
	// Component not in knownShadcnExports — guard must not touch it
	original := `import { WeirdThing } from '@/components/ui/custom-widget';`
	files := map[string]string{
		"src/components/Weird.tsx": original,
	}
	n := fixHallucinatedShadcnExports(files)
	if n != 0 {
		t.Fatalf("expected 0 files modified, got %d", n)
	}
	if files["src/components/Weird.tsx"] != original {
		t.Error("unknown component was modified — guard should be conservative")
	}
}

func TestFixHallucinatedShadcnExports_DeduplicatesAfterReplacement(t *testing.T) {
	// Button already present; StyledButton → Button would duplicate
	files := map[string]string{
		"src/components/Dup.tsx": `import { Button, StyledButton } from '@/components/ui/button';

export const A = <Button />;
export const B = <StyledButton />;`,
	}
	n := fixHallucinatedShadcnExports(files)
	if n != 1 {
		t.Fatalf("expected 1 file modified, got %d", n)
	}
	result := files["src/components/Dup.tsx"]
	importLine := result[:strings.Index(result, "\n")]
	occurrences := strings.Count(importLine, "Button")
	if occurrences != 1 {
		t.Errorf("expected exactly 1 Button in import, got %d: %q", occurrences, importLine)
	}
}

func TestFixHallucinatedShadcnExports_ReplacesInJSXBody(t *testing.T) {
	files := map[string]string{
		"src/components/Form.tsx": `import { GlassInput } from '@/components/ui/input';

export function MyForm() {
  return <GlassInput placeholder="Email" />;
}`,
	}
	n := fixHallucinatedShadcnExports(files)
	if n != 1 {
		t.Fatalf("expected 1 file modified, got %d", n)
	}
	result := files["src/components/Form.tsx"]
	if strings.Contains(result, "GlassInput") {
		t.Error("GlassInput still present in JSX body")
	}
	if !strings.Contains(result, "Input") {
		t.Error("canonical Input not found in result")
	}
}

// ─── deduplicateImportSpecifiers ─────────────────────────────────────────────

func TestDeduplicateImportSpecifiers_RemovesDuplicates(t *testing.T) {
	input := `import { Button, Button, Card } from '@/components/ui/button';`
	result := deduplicateImportSpecifiers(input)
	if strings.Count(result, "Button") != 1 {
		t.Errorf("expected 1 Button, got result: %q", result)
	}
	if !strings.Contains(result, "Card") {
		t.Error("Card should be preserved")
	}
}

func TestDeduplicateImportSpecifiers_NoChangeWhenClean(t *testing.T) {
	input := `import { Button, Card, Input } from '@/components/ui/foo';`
	result := deduplicateImportSpecifiers(input)
	if result != input {
		t.Errorf("clean input should not be modified\ngot:  %q\nwant: %q", result, input)
	}
}
