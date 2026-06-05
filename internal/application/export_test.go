package application

import (
	"context"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
)

type noopUIMedia struct{}

func (noopUIMedia) GenerateUIAssets(context.Context, string, string, []string) (*ports.UIAssets, error) {
	return &ports.UIAssets{}, nil
}

func (noopUIMedia) SynthesizePromptsOnly(context.Context, string, string, []string) (*ports.UIAssets, error) {
	return &ports.UIAssets{}, nil
}

func (noopUIMedia) GenerateImage(context.Context, string, int, int) (string, error) {
	return "", nil
}

func (noopUIMedia) GeneratePromoVideo(context.Context, string, string) (*ports.PromoVideo, error) {
	return &ports.PromoVideo{}, nil
}

// NewOrchestratorForTest создаёт оркестратор с noop UI media (для application_test).
func NewOrchestratorForTest(llm ports.LLMProvider) *Orchestrator {
	return NewOrchestrator(llm, noopUIMedia{})
}

// MaxFilesPerGroupForTest — лимит файлов в одном чанке (для application_test).
const MaxFilesPerGroupForTest = maxFilesPerGroup

// InspectorProviderPathForTest — путь InspectorProvider в сгенерированных проектах.
const InspectorProviderPathForTest = inspectorProviderPath

// ParseCodeFilesForTest вызывает parseCodeFiles для внешних тестов (application_test).
func ParseCodeFilesForTest(o *Orchestrator, content string) map[string]string {
	return o.parseCodeFiles(context.Background(), content)
}

// FileGroupForTest — снимок fileGroup для application_test.
type FileGroupForTest struct {
	Name  string
	Tier  int
	Files []string
}

// GroupFileMapForTest вызывает groupFileMap для внешних тестов.
func GroupFileMapForTest(fileMap []string) []FileGroupForTest {
	groups := groupFileMap(fileMap)
	out := make([]FileGroupForTest, len(groups))
	for i, g := range groups {
		out[i] = FileGroupForTest{Name: g.Name, Tier: g.Tier, Files: g.Files}
	}

	return out
}

// InjectInspectorProviderForTest вызывает injectInspectorProvider для внешних тестов.
func InjectInspectorProviderForTest(files map[string]string) {
	injectInspectorProvider(context.Background(), files)
}

// BusFromCtxForTest возвращает шину событий сессии из контекста.
func BusFromCtxForTest(ctx context.Context, o *Orchestrator) *domain.EventBus {
	return o.busFromCtx(ctx)
}

// SubscribeOrchestratorEventsForTest подписывается на дефолтную шину оркестратора.
func SubscribeOrchestratorEventsForTest(o *Orchestrator) <-chan domain.AgentEvent {
	return o.events.Subscribe()
}

// GenerateCodeChunkedForTest вызывает generateCodeChunked для внешних тестов.
func GenerateCodeChunkedForTest(
	ctx context.Context,
	o *Orchestrator,
	specification string,
	manifest *SystemManifest,
	plan *MasterPlan,
	audit *ReverseEngineeringResult,
	features []CompetitorFeature,
	imageURLs map[string]string,
) (map[string]string, error) {
	return o.generateCodeChunked(ctx, specification, manifest, plan, audit, features, imageURLs)
}
