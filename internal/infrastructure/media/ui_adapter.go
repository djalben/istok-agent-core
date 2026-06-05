package media

import (
	"context"
	"os"

	"github.com/djalben/istok-agent-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// uiAdapter реализует ports.UIMediaService поверх media.Service.
type uiAdapter struct {
	svc *Service
}

var _ ports.UIMediaService = (*uiAdapter)(nil)

// NewUIMediaService создаёт адаптер UI-медиа для application/transport.
func NewUIMediaService(llm ports.LLMProvider) ports.UIMediaService {
	return &uiAdapter{svc: NewServiceWithLLM(os.Getenv("REPLICATE_API_TOKEN"), llm)}
}

func (a *uiAdapter) GenerateUIAssets(ctx context.Context, projectName, spec string, colors []string) (*ports.UIAssets, error) {
	out, err := a.svc.GenerateUIAssets(ctx, projectName, spec, colors)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return toPortsUIAssets(out), nil
}

func (a *uiAdapter) SynthesizePromptsOnly(ctx context.Context, projectName, spec string, colors []string) (*ports.UIAssets, error) {
	out, err := a.svc.SynthesizePromptsOnly(ctx, projectName, spec, colors)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return toPortsUIAssets(out), nil
}

func (a *uiAdapter) GenerateImage(ctx context.Context, prompt string, width, height int) (string, error) {
	out, err := a.svc.GenerateImage(ctx, prompt, width, height)
	if err != nil {
		return "", wrapper.Wrap(err)
	}

	return out, nil
}

func (a *uiAdapter) GeneratePromoVideo(ctx context.Context, projectName, spec string) (*ports.PromoVideo, error) {
	out, err := a.svc.GeneratePromoVideo(ctx, projectName, spec)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return toPortsPromoVideo(out), nil
}

func toPortsUIAssets(a *Assets) *ports.UIAssets {
	if a == nil {
		return nil
	}

	return &ports.UIAssets{
		LogoSVG: a.LogoSVG, ColorPalette: a.ColorPalette, IconSet: a.IconSet,
		HeroPrompt: a.HeroPrompt, OGImagePrompt: a.OGImagePrompt, VideoPrompts: a.VideoPrompts,
		HeroImageURL: a.HeroImageURL, OGImageURL: a.OGImageURL, GeneratedAt: a.GeneratedAt,
	}
}

func toPortsPromoVideo(v *PromoVideo) *ports.PromoVideo {
	if v == nil {
		return nil
	}

	return &ports.PromoVideo{
		Script: v.Script, Duration: v.Duration, Scenes: v.Scenes,
		Voiceover: v.Voiceover, MusicStyle: v.MusicStyle, VideoURL: v.VideoURL, GeneratedAt: v.GeneratedAt,
	}
}
