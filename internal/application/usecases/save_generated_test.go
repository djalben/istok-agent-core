package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/infrastructure/persistence"
)

// TestSaveGenerated_Create — Layer 2 auto-save: новый проект создаётся с файлами.
func TestSaveGenerated_Create(t *testing.T) {
	repo := persistence.NewProjectRepoMemory()
	svc := usecases.NewProjectService(repo)
	ctx := context.Background()

	files := map[string]string{"index.html": "<h1>hi</h1>", "style.css": "body{}"}
	p, err := svc.SaveGenerated(ctx, "owner-1", "", "Hello World Test", "React", "Create a hello world page", files)
	if err != nil {
		t.Fatalf("SaveGenerated create: %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected generated project ID")
	}
	if p.OwnerID != "owner-1" {
		t.Fatalf("owner mismatch: %s", p.OwnerID)
	}
	if p.Name != "Hello World Test" {
		t.Fatalf("name mismatch: %s", p.Name)
	}
	if len(p.Files) != 2 || p.Files["index.html"] != "<h1>hi</h1>" {
		t.Fatalf("files not persisted: %#v", p.Files)
	}

	// Файлы должны быть доступны через Get
	got, err := svc.Get(ctx, "owner-1", p.ID)
	if err != nil {
		t.Fatalf("Get after create: %v", err)
	}
	if len(got.Files) != 2 {
		t.Fatalf("Get returned %d files, want 2", len(got.Files))
	}
}

// TestSaveGenerated_Update — projectID задан → обновляем существующий проект.
func TestSaveGenerated_Update(t *testing.T) {
	repo := persistence.NewProjectRepoMemory()
	svc := usecases.NewProjectService(repo)
	ctx := context.Background()

	orig, err := svc.SaveGenerated(ctx, "owner-1", "", "v1", "React", "spec", map[string]string{"a.txt": "1"})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	updated, err := svc.SaveGenerated(ctx, "owner-1", orig.ID, "", "React", "spec", map[string]string{"a.txt": "2", "b.txt": "3"})
	if err != nil {
		t.Fatalf("SaveGenerated update: %v", err)
	}
	if updated.ID != orig.ID {
		t.Fatalf("update created new project: %s != %s", updated.ID, orig.ID)
	}
	if updated.Files["a.txt"] != "2" || len(updated.Files) != 2 {
		t.Fatalf("files not updated: %#v", updated.Files)
	}
}

// TestSaveGenerated_ForeignProject — чужой projectID → ErrNotFound (изоляция владельцев).
func TestSaveGenerated_ForeignProject(t *testing.T) {
	repo := persistence.NewProjectRepoMemory()
	svc := usecases.NewProjectService(repo)
	ctx := context.Background()

	other, _ := svc.SaveGenerated(ctx, "owner-A", "", "secret", "React", "spec", map[string]string{"x": "y"})

	_, err := svc.SaveGenerated(ctx, "owner-B", other.ID, "", "", "spec", map[string]string{"hack": "1"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for foreign project, got: %v", err)
	}
}
