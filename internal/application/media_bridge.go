package application

import (
	"context"
	"os"

	"github.com/djalben/istok-agent-core/internal/infrastructure/media"
	"github.com/djalben/istok-agent-core/internal/ports"
)

// mediaServiceBridge обёртка для MediaService в слое application
type mediaServiceBridge struct {
	svc *media.MediaService
}

// newMediaService создаёт мост к MediaService для использования в оркестраторе.
// Принимает ports.LLMProvider для совместимости сигнатуры (не используется —
// MediaService использует собственный Replicate-клиент для FLUX image generation).
// API-токен Replicate читается из переменной окружения.
func newMediaService(llm ports.LLMProvider) *mediaServiceBridge {
	apiKey := os.Getenv("REPLICATE_API_TOKEN")
	return &mediaServiceBridge{
		svc: media.NewMediaServiceWithLLM(apiKey, llm),
	}
}

// GenerateUIAssets генерирует UI-ассеты (full pipeline: prompts + images)
func (b *mediaServiceBridge) GenerateUIAssets(ctx context.Context, projectName, spec string, colors []string) (*media.MediaAssets, error) {
	return b.svc.GenerateUIAssets(ctx, projectName, spec, colors)
}

// SynthesizePromptsOnly генерирует промпты для медиа (без вызова Replicate).
// Возвращает MediaAssets с заполненными HeroPrompt и OGImagePrompt.
func (b *mediaServiceBridge) SynthesizePromptsOnly(ctx context.Context, projectName, spec string, colors []string) (*media.MediaAssets, error) {
	return b.svc.SynthesizePromptsOnly(ctx, projectName, spec, colors)
}

// GenerateImage вызывает Replicate nano-banana для одного промпта.
func (b *mediaServiceBridge) GenerateImage(ctx context.Context, prompt string, width, height int) (string, error) {
	return b.svc.GenerateImage(ctx, prompt, width, height)
}

// GeneratePromoVideo генерирует сценарий промо-видео
func (b *mediaServiceBridge) GeneratePromoVideo(ctx context.Context, projectName, spec string) (*media.PromoVideo, error) {
	return b.svc.GeneratePromoVideo(ctx, projectName, spec)
}
