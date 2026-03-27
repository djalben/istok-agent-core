package application

import (
	"context"

	"github.com/istok/agent-core/internal/infrastructure/media"
)

// mediaServiceBridge обёртка для MediaService в слое application
type mediaServiceBridge struct {
	svc *media.MediaService
}

// newMediaService создаёт мост к MediaService для использования в оркестраторе
func newMediaService(apiKey string) *mediaServiceBridge {
	return &mediaServiceBridge{
		svc: media.NewMediaService(apiKey),
	}
}

// GenerateUIAssets генерирует UI-ассеты
func (b *mediaServiceBridge) GenerateUIAssets(ctx context.Context, projectName, spec string, colors []string) (*media.MediaAssets, error) {
	return b.svc.GenerateUIAssets(ctx, projectName, spec, colors)
}

// GeneratePromoVideo генерирует сценарий промо-видео
func (b *mediaServiceBridge) GeneratePromoVideo(ctx context.Context, projectName, spec string) (*media.PromoVideo, error) {
	return b.svc.GeneratePromoVideo(ctx, projectName, spec)
}
