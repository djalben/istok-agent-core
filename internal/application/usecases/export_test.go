package usecases

import "github.com/djalben/istok-agent-core/internal/ports"

// NewMemoryProjectRepoForTest — in-memory ProjectRepository для white-box тестов пакета usecases.
func NewMemoryProjectRepoForTest() ports.ProjectRepository {
	return newMemoryProjectRepo()
}
