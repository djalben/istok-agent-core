package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
)

// CreateProjectInput — данные для POST /projects.
type CreateProjectInput struct {
	Name        string
	Description string
	Framework   string
	Prompt      string
	IsPublic    bool
	Files       map[string]string
}

// ProjectService — бизнес-логика проектов с проверкой владения.
// Зависит только от порта ProjectRepository (Dependency Rule).
type ProjectService struct {
	repo ports.ProjectRepository
}

func NewProjectService(repo ports.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

// List возвращает проекты пользователя (без files).
func (s *ProjectService) List(ctx context.Context, ownerID string) ([]*domain.Project, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

// Get возвращает полный проект, если он принадлежит пользователю.
// Чужой/несуществующий → ErrNotFound (не раскрываем существование).
func (s *ProjectService) Get(ctx context.Context, ownerID, id string) (*domain.Project, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.OwnerID != ownerID {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

// Create сохраняет новый проект.
func (s *ProjectService) Create(ctx context.Context, ownerID string, in CreateProjectInput) (*domain.Project, error) {
	now := time.Now().UTC()
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = "Новый проект"
	}
	files := in.Files
	if files == nil {
		files = map[string]string{}
	}
	p := &domain.Project{
		ID:          domain.GenerateID(),
		OwnerID:     ownerID,
		Name:        name,
		Description: in.Description,
		Framework:   in.Framework,
		IsPublic:    in.IsPublic,
		Prompt:      in.Prompt,
		Files:       files,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Update применяет частичный патч к проекту пользователя.
// folder_id / workspace_id принимаются, но пока не персистятся (stub-фаза).
func (s *ProjectService) Update(ctx context.Context, ownerID, id string, patch domain.ProjectPatch) (*domain.Project, error) {
	p, err := s.Get(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	if patch.Name != nil {
		p.Name = *patch.Name
	}
	if patch.Description != nil {
		p.Description = *patch.Description
	}
	if patch.Framework != nil {
		p.Framework = *patch.Framework
	}
	if patch.IsPublic != nil {
		p.IsPublic = *patch.IsPublic
	}
	if patch.Files != nil {
		p.Files = patch.Files
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Remix клонирует проект пользователя в новый (с новым id).
func (s *ProjectService) Remix(ctx context.Context, ownerID, id, name string, includeHistory bool) (*domain.Project, error) {
	src, err := s.Get(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	cloneName := strings.TrimSpace(name)
	if cloneName == "" {
		cloneName = src.Name + " (ремикс)"
	}
	prompt := ""
	if includeHistory {
		prompt = src.Prompt
	}
	files := map[string]string{}
	for k, v := range src.Files {
		files[k] = v
	}
	now := time.Now().UTC()
	clone := &domain.Project{
		ID:          domain.GenerateID(),
		OwnerID:     ownerID,
		Name:        cloneName,
		Description: src.Description,
		Framework:   src.Framework,
		IsPublic:    false,
		Prompt:      prompt,
		Files:       files,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, clone); err != nil {
		return nil, err
	}
	return clone, nil
}

// Delete удаляет проект пользователя.
func (s *ProjectService) Delete(ctx context.Context, ownerID, id string) error {
	if _, err := s.Get(ctx, ownerID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// SaveGenerated персистит результат LLM-генерации (Layer 2 auto-save).
// projectID пуст → создаём новый проект; иначе обновляем существующий проект владельца.
// Чужой/несуществующий projectID при обновлении → ErrNotFound.
func (s *ProjectService) SaveGenerated(ctx context.Context, ownerID, projectID, name, framework, prompt string, files map[string]string) (*domain.Project, error) {
	if files == nil {
		files = map[string]string{}
	}
	if projectID != "" {
		p, err := s.Get(ctx, ownerID, projectID)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(name) != "" {
			p.Name = strings.TrimSpace(name)
		}
		if framework != "" {
			p.Framework = framework
		}
		if prompt != "" {
			p.Prompt = prompt
		}
		p.Files = files
		if err := s.repo.Update(ctx, p); err != nil {
			return nil, err
		}
		return p, nil
	}
	return s.Create(ctx, ownerID, CreateProjectInput{
		Name:      deriveProjectName(name, prompt),
		Framework: framework,
		Prompt:    prompt,
		Files:     files,
	})
}

// deriveProjectName выбирает имя: явное name, иначе первая строка промпта, иначе дефолт.
func deriveProjectName(name, prompt string) string {
	if n := strings.TrimSpace(name); n != "" {
		return n
	}
	line := strings.TrimSpace(strings.SplitN(prompt, "\n", 2)[0])
	if line == "" {
		return "Сгенерированный проект"
	}
	if len(line) > 60 {
		line = line[:60]
	}
	return line
}

// Stats агрегирует статистику для страницы профиля.
func (s *ProjectService) Stats(ctx context.Context, ownerID string) (domain.ProfileStats, error) {
	total, err := s.repo.CountByOwner(ctx, ownerID)
	if err != nil {
		return domain.ProfileStats{}, err
	}
	published, err := s.repo.CountPublishedByOwner(ctx, ownerID)
	if err != nil {
		return domain.ProfileStats{}, err
	}
	return domain.ProfileStats{
		TotalProjects:     total,
		PublishedProjects: published,
		TotalGenerations:  total, // proxy until a generations table exists
		DaysActive:        0,
		CurrentStreak:     0,
	}, nil
}
