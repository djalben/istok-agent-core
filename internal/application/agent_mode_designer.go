package application

import (
	"fmt"
	"strings"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (run *agentModeRun) phaseAgentDesigner() {
	svc := run.o.uiMedia
	run.imageURLs = map[string]string{}

	assets, mediaAssets := run.designerSynthesizePrompts(svc)
	run.designerFulfillMedia(svc, assets, mediaAssets, run.designerColors())
	applog(run.ctx).InfoContext(run.ctx, "designer phase complete", "imageUrls", len(run.imageURLs))
}

func (run *agentModeRun) designerColors() []string {
	if run.result.VisualAudit != nil {
		return run.result.VisualAudit.Colors
	}

	return nil
}

func (run *agentModeRun) designerSynthesizePrompts(svc ports.UIMediaService) (*ports.UIAssets, []domain.MediaAsset) {
	applog(run.ctx).DebugContext(run.ctx, "designer prompt synthesis start")
	run.o.sendStatus(run.ctx, RoleDesigner, "running", "🎨 Designer: Синтезирую описания для изображений...", 35)

	assets, synthErr := svc.SynthesizePromptsOnly(run.ctx, run.specification, run.specification, run.designerColors())
	if synthErr != nil {
		applog(run.ctx).WarnContext(run.ctx, "designer prompt synthesis error", "error", wrapper.Wrap(synthErr))
		run.o.sendStatus(run.ctx, RoleDesigner, "error", "⚠️ Не удалось сгенерировать описания для медиа", 0)
	}

	return assets, run.designerBuildMediaAssets(assets)
}

func (run *agentModeRun) designerBuildMediaAssets(assets *ports.UIAssets) []domain.MediaAsset {
	if assets == nil {
		return nil
	}
	var mediaAssets []domain.MediaAsset
	if assets.HeroPrompt != "" {
		kw := stockKeywords(assets.HeroPrompt)
		mediaAssets = append(mediaAssets, domain.MediaAsset{
			ID: "hero", Type: "image", Placement: "hero",
			Label: "Главный фон (Hero)", Prompt: assets.HeroPrompt,
			StockKeywords: kw, PreviewURL: "https://source.unsplash.com/1344x768/?" + kw,
		})
	}
	if assets.OGImagePrompt != "" {
		kw := stockKeywords(assets.OGImagePrompt)
		mediaAssets = append(mediaAssets, domain.MediaAsset{
			ID: "og", Type: "image", Placement: "og",
			Label: "Превью для соцсетей (OG)", Prompt: assets.OGImagePrompt,
			StockKeywords: kw, PreviewURL: "https://source.unsplash.com/1200x630/?" + kw,
		})
	}
	for i, vp := range assets.VideoPrompts {
		mediaAssets = append(mediaAssets, domain.MediaAsset{
			ID: fmt.Sprintf("video_%d", i+1), Type: "video", Placement: "promo_video",
			Label: fmt.Sprintf("Промо-ролик (вариант %d)", i+1), Prompt: vp,
		})
	}

	return mediaAssets
}

func (run *agentModeRun) designerFulfillMedia(svc ports.UIMediaService, assets *ports.UIAssets, mediaAssets []domain.MediaAsset, designColors []string) {
	switch {
	case len(mediaAssets) > 0 && run.sessionID != "" && run.o.approvalRegistry != nil:
		run.designerWithApproval(svc, assets, mediaAssets)
	case len(mediaAssets) == 0:
		applog(run.ctx).InfoContext(run.ctx, "no media prompts, skipping approval")
		run.o.sendStatus(run.ctx, RoleDesigner, "completed", "⏭️ Медиа-промпты отсутствуют — пропуск", 100)
	default:
		run.designerFallbackGenerate(svc, designColors)
	}
}

func (run *agentModeRun) designerWithApproval(svc ports.UIMediaService, assets *ports.UIAssets, mediaAssets []domain.MediaAsset) {
	run.o.approvalRegistry.RegisterMedia(run.ctx, run.sessionID)
	run.o.busFromCtx(run.ctx).PublishMediaApproval(domain.RoleDesigner, mediaAssets, run.sessionID)
	run.o.sendStatus(run.ctx, RoleDesigner, "running", "⏸️ Ожидание утверждения медиа-ассетов...", 38)

	mediaDecision, mediaErr := run.o.approvalRegistry.WaitForMediaApproval(run.ctx, run.sessionID)
	switch {
	case mediaErr != nil:
		applog(run.ctx).WarnContext(run.ctx, "media approval wait failed", "error", wrapper.Wrap(mediaErr))
		run.o.sendStatus(run.ctx, RoleDesigner, "error", "⚠️ Медиа пропущено (соединение потеряно)", 0)
	case !mediaDecision.Approved:
		applog(run.ctx).InfoContext(run.ctx, "media generation skipped by user")
		run.o.sendStatus(run.ctx, RoleDesigner, "completed", "⏭️ Генерация медиа пропущена пользователем", 100)
	default:
		run.designerApplyApprovedPrompts(assets, mediaDecision.Assets)
		run.designerGenerateImages(svc, assets)
		run.designerStoreAssets(assets)
		run.o.sendStatus(run.ctx, RoleDesigner, "completed",
			fmt.Sprintf("✅ Дизайн готов: %d изображений, SVG логотип", len(run.imageURLs)), 100)
		if guide := buildMediaGuidelines(mediaDecision.Assets, run.imageURLs); guide != "" {
			run.specification += guide
			applog(run.ctx).InfoContext(run.ctx, "media guidelines injected", "chars", len(guide))
		}
	}
}

func (run *agentModeRun) designerApplyApprovedPrompts(assets *ports.UIAssets, approved []domain.MediaAsset) {
	applog(run.ctx).InfoContext(run.ctx, "media assets approved", "count", len(approved))
	for _, a := range approved {
		switch a.Placement {
		case "hero":
			if a.Prompt != "" {
				assets.HeroPrompt = a.Prompt
			}
		case "og":
			if a.Prompt != "" {
				assets.OGImagePrompt = a.Prompt
			}
		}
	}
}

func (run *agentModeRun) designerGenerateImages(svc ports.UIMediaService, assets *ports.UIAssets) {
	if assets == nil {
		return
	}
	run.o.sendStatus(run.ctx, RoleDesigner, "running", "🎨 Designer: Генерирую изображения с утверждёнными промптами...", 42)
	if assets.HeroPrompt != "" {
		url, err := svc.GenerateImage(run.ctx, assets.HeroPrompt, 1344, 768)
		if err == nil {
			run.imageURLs["hero"] = url
			applog(run.ctx).InfoContext(run.ctx, "hero image generated", "url", url)
		} else {
			applog(run.ctx).WarnContext(run.ctx, "hero image failed", "error", wrapper.Wrap(err))
			if strings.Contains(err.Error(), "402") {
				run.o.sendStatus(run.ctx, RoleDesigner, "error", "⚠️ Визуализация временно недоступна (превышен бюджет)", 0)
			}
		}
	}
	if assets.OGImagePrompt != "" {
		url, err := svc.GenerateImage(run.ctx, assets.OGImagePrompt, 1200, 630)
		if err == nil {
			run.imageURLs["og"] = url
			applog(run.ctx).InfoContext(run.ctx, "og image generated", "url", url)
		} else {
			applog(run.ctx).WarnContext(run.ctx, "og image failed", "error", wrapper.Wrap(err))
		}
	}
}

func (run *agentModeRun) designerStoreAssets(assets *ports.UIAssets) {
	if assets == nil {
		return
	}
	run.o.mu.Lock()
	run.result.Assets = map[string]string{
		"logo.svg": assets.LogoSVG, "hero_prompt": assets.HeroPrompt,
		"og_prompt": assets.OGImagePrompt, "color_palette": fmt.Sprintf("%v", assets.ColorPalette),
	}
	if run.imageURLs["hero"] != "" {
		run.result.Assets["hero_image_url"] = run.imageURLs["hero"]
	}
	if run.imageURLs["og"] != "" {
		run.result.Assets["og_image_url"] = run.imageURLs["og"]
	}
	run.o.mu.Unlock()
}

func (run *agentModeRun) designerFallbackGenerate(svc ports.UIMediaService, designColors []string) {
	applog(run.ctx).WarnContext(run.ctx, "media generation without approval (no session)")
	run.o.sendStatus(run.ctx, RoleDesigner, "running", "🎨 Designer: Генерирую визуальные ассеты...", 40)
	fullAssets, designErr := svc.GenerateUIAssets(run.ctx, run.specification, run.specification, designColors)
	if designErr != nil {
		applog(run.ctx).WarnContext(run.ctx, "designer fallback error", "error", wrapper.Wrap(designErr))
		userMsg := "⚠️ Визуализация временно недоступна"
		if strings.Contains(designErr.Error(), "402") {
			userMsg = "⚠️ Визуализация временно недоступна (превышен бюджет медиа-сервиса)"
		}
		run.o.sendStatus(run.ctx, RoleDesigner, "error", userMsg, 0)

		return
	}
	if fullAssets.HeroImageURL != "" {
		run.imageURLs["hero"] = fullAssets.HeroImageURL
	}
	if fullAssets.OGImageURL != "" {
		run.imageURLs["og"] = fullAssets.OGImageURL
	}
	run.o.mu.Lock()
	run.result.Assets = map[string]string{
		"logo.svg": fullAssets.LogoSVG, "hero_prompt": fullAssets.HeroPrompt,
		"og_prompt": fullAssets.OGImagePrompt, "color_palette": fmt.Sprintf("%v", fullAssets.ColorPalette),
	}
	if fullAssets.HeroImageURL != "" {
		run.result.Assets["hero_image_url"] = fullAssets.HeroImageURL
	}
	if fullAssets.OGImageURL != "" {
		run.result.Assets["og_image_url"] = fullAssets.OGImageURL
	}
	run.o.mu.Unlock()
	run.o.sendStatus(run.ctx, RoleDesigner, "completed", fmt.Sprintf("✅ Дизайн готов: %d изображений", len(run.imageURLs)), 100)
}
