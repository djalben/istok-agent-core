package ports

import (
	"context"
	"time"
)

// UIAssets — результат синтеза/генерации UI-медиа (hero, OG, промпты).
type UIAssets struct {
	LogoSVG       string
	ColorPalette  []string
	IconSet       map[string]string
	HeroPrompt    string
	OGImagePrompt string
	VideoPrompts  []string
	HeroImageURL  string
	OGImageURL    string
	GeneratedAt   time.Time
}

// PromoVideo — сценарий и метаданные промо-ролика.
type PromoVideo struct {
	Script      string
	Duration    string
	Scenes      []string
	Voiceover   string
	MusicStyle  string
	VideoURL    string
	GeneratedAt time.Time
}

// UIMediaService — оркестрация UI-медиа (промпты, изображения, промо-видео).
type UIMediaService interface {
	GenerateUIAssets(ctx context.Context, projectName, spec string, colors []string) (*UIAssets, error)
	SynthesizePromptsOnly(ctx context.Context, projectName, spec string, colors []string) (*UIAssets, error)
	GenerateImage(ctx context.Context, prompt string, width, height int) (string, error)
	GeneratePromoVideo(ctx context.Context, projectName, spec string) (*PromoVideo, error)
}
