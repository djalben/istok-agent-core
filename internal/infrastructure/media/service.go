package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/djalben/istok-agent-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Media Service V3
//  Replicate-only: nano-banana (image) + Veo 3 (video).
//  Text prompt synthesis делегируется ports.LLMProvider (Anthropic).
//  OpenRouter полностью удалён.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	// ModelImageNanoBanana — Replicate: Google Nano Banana (text→image).
	ModelImageNanoBanana = "google/nano-banana"
	// ModelVideoVeo3 — Replicate: Google Veo 3 (text→video).
	ModelVideoVeo3 = "google/veo-3"

	// AspectRatio16x9 — соотношение сторон по умолчанию для видео.
	AspectRatio16x9 = "16:9"

	// ModelPromptAnthropic — модель для синтеза промптов (если задан LLMProvider).
	ModelPromptAnthropic = "anthropic/claude-sonnet-4-6"

	replicateAPIBase = "https://api.replicate.com/v1"
)

// Assets — результат генерации UI-ассетов.
type Assets struct {
	LogoSVG       string            `json:"logo_svg"`
	ColorPalette  []string          `json:"color_palette"`
	IconSet       map[string]string `json:"icon_set"`
	HeroPrompt    string            `json:"hero_prompt"`
	OGImagePrompt string            `json:"og_image_prompt"`
	VideoPrompts  []string          `json:"video_prompts,omitempty"` // 3 варианта промо-ролика
	HeroImageURL  string            `json:"hero_image_url,omitempty"`
	OGImageURL    string            `json:"og_image_url,omitempty"`
	GeneratedAt   time.Time         `json:"generated_at"`
}

// PromoVideo — концепт + ссылка на сгенерированное промо-видео.
type PromoVideo struct {
	Script      string    `json:"script"`
	Duration    string    `json:"duration"`
	Scenes      []string  `json:"scenes"`
	Voiceover   string    `json:"voiceover"`
	MusicStyle  string    `json:"music_style"`
	VideoURL    string    `json:"video_url,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

// Service генерирует медиа через Replicate (nano-banana + Veo 3).
// LLMProvider опционален — используется только для генерации промптов/сценариев.
type Service struct {
	replicateToken string
	imageModel     string
	videoModel     string
	llm            ports.LLMProvider // optional — nil-safe
	httpClient     *http.Client
}

// NewService — конструктор без LLM (использует шаблонные ассеты).
func NewService(replicateToken string) *Service {
	return NewServiceWithLLM(replicateToken, nil)
}

// NewServiceWithLLM — конструктор Service с LLMProvider для синтеза промптов.
func NewServiceWithLLM(replicateToken string, llm ports.LLMProvider) *Service {
	img := os.Getenv("IMAGE_MODEL_ID")
	if img == "" {
		img = ModelImageNanoBanana
	}
	vid := os.Getenv("VIDEO_MODEL_ID")
	if vid == "" {
		vid = ModelVideoVeo3
	}

	return &Service{
		replicateToken: replicateToken,
		imageModel:     img,
		videoModel:     vid,
		llm:            llm,
		httpClient:     &http.Client{Timeout: 5 * time.Minute},
	}
}

// ──────────────────────────────────────────────────────────────
//  UI ASSETS
// ──────────────────────────────────────────────────────────────

// GenerateUIAssets — синтезирует промпты (через LLM, если доступен) и
// запускает nano-banana для hero/OG-изображений.
func (s *Service) GenerateUIAssets(ctx context.Context, projectName, spec string, colors []string) (*Assets, error) {
	l := ports.LoggerFromContext(ctx)
	l.InfoContext(ctx, "ui assets generation started", "projectName", projectName)

	assets := s.defaultAssets(projectName, colors)

	// 1) Если LLM есть — обновляем logo_svg / icon_set / hero_prompt / og_prompt.
	if s.llm != nil {
		synthesized, synthErr := s.synthesizePrompts(ctx, projectName, spec, colors)
		if synthErr == nil {
			s.mergeAssets(assets, synthesized)
		} else {
			l.WarnContext(ctx, "prompt synthesis failed, using defaults", "error", wrapper.Wrap(synthErr))
		}
	}

	// 2) Генерация hero (Replicate nano-banana).
	if assets.HeroPrompt != "" {
		url, imgErr := s.GenerateImage(ctx, assets.HeroPrompt, 1344, 768)
		if imgErr == nil {
			assets.HeroImageURL = url
			l.InfoContext(ctx, "nano-banana hero generated", "url", url)
		} else {
			l.WarnContext(ctx, "nano-banana hero failed", "error", wrapper.Wrap(imgErr))
		}
	}
	// 3) Генерация OG.
	if assets.OGImagePrompt != "" {
		url, imgErr := s.GenerateImage(ctx, assets.OGImagePrompt, 1200, 630)
		if imgErr == nil {
			assets.OGImageURL = url
			l.InfoContext(ctx, "nano-banana og generated", "url", url)
		} else {
			l.WarnContext(ctx, "nano-banana og failed", "error", wrapper.Wrap(imgErr))
		}
	}

	return assets, nil
}

// SynthesizePromptsOnly — генерирует промпты для медиа (без вызова Replicate).
// Используется для Human-in-the-Loop: показать пользователю промпты ДО оплаты.
// Также генерирует 3 варианта промо-ролика.
func (s *Service) SynthesizePromptsOnly(ctx context.Context, projectName, spec string, colors []string) (*Assets, error) {
	assets := s.defaultAssets(projectName, colors)
	s.enrichPromptAssetsFromLLM(ctx, assets, projectName, spec, colors)

	return assets, nil
}

// ──────────────────────────────────────────────────────────────
//  IMAGE — Replicate nano-banana
// ──────────────────────────────────────────────────────────────

// GenerateImage — text→image через Replicate nano-banana.
func (s *Service) GenerateImage(ctx context.Context, prompt string, width, height int) (string, error) {
	if s.replicateToken == "" {
		return "", ErrReplicateTokenNotSet
	}
	ports.LoggerFromContext(ctx).InfoContext(ctx, "nano-banana image generation",
		"width", width,
		"height", height,
	)

	endpoint := fmt.Sprintf("%s/models/%s/predictions", replicateAPIBase, s.imageModel)
	payload, err := json.Marshal(map[string]any{
		"input": map[string]any{
			"prompt":        prompt,
			"aspect_ratio":  aspectRatio(width, height),
			"output_format": "png",
			"safety_filter": "block_only_high",
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal replicate payload: %w", err)
	}

	pred, err := s.replicateCreate(ctx, endpoint, payload, true)
	if err != nil {
		return "", err
	}
	if url := extractURL(pred.Output); url != "" && pred.Status == "succeeded" {
		return url, nil
	}
	poll, err := s.replicatePoll(ctx, pred, 2*time.Minute, 2*time.Second)
	if err != nil {
		return "", err
	}

	return extractURL(poll.Output), nil
}

// ──────────────────────────────────────────────────────────────
//  VIDEO — Replicate Veo 3
// ──────────────────────────────────────────────────────────────

// VeoRequest — запрос на генерацию видео.
type VeoRequest struct {
	Prompt   string `json:"prompt"`
	Duration string `json:"duration"`
	Style    string `json:"style"`
	Aspect   string `json:"aspect"`
}

// VeoResult — ссылка на сгенерированное видео.
type VeoResult struct {
	VideoURL string `json:"video_url"`
	Status   string `json:"status"`
	Duration string `json:"duration"`
	Error    string `json:"error,omitempty"`
}

// GenerateVideoVeo — text→video через Replicate google/veo-3.
func (s *Service) GenerateVideoVeo(ctx context.Context, req VeoRequest) (*VeoResult, error) {
	if s.replicateToken == "" {
		return nil, ErrReplicateTokenNotSet
	}
	ports.LoggerFromContext(ctx).InfoContext(ctx, "veo3 video generation",
		"duration", req.Duration,
		"style", req.Style,
		"aspect", req.Aspect,
	)

	aspect := req.Aspect
	if aspect == "" {
		aspect = AspectRatio16x9
	}
	endpoint := fmt.Sprintf("%s/models/%s/predictions", replicateAPIBase, s.videoModel)
	payload, err := json.Marshal(map[string]any{
		"input": map[string]any{
			"prompt":          req.Prompt,
			"aspect_ratio":    aspect,
			"duration":        req.Duration,
			"style":           req.Style,
			"negative_prompt": "low quality, blurry, watermark",
		},
	})
	if err != nil {
		return &VeoResult{Status: "failed", Error: err.Error()}, fmt.Errorf("marshal replicate payload: %w", err)
	}

	pred, err := s.replicateCreate(ctx, endpoint, payload, false)
	if err != nil {
		return &VeoResult{Status: "failed", Error: err.Error()}, err
	}

	poll, err := s.replicatePoll(ctx, pred, 10*time.Minute, 5*time.Second)
	if err != nil {
		return &VeoResult{Status: "failed", Error: err.Error()}, err
	}

	return &VeoResult{
		VideoURL: extractURL(poll.Output),
		Status:   "completed",
		Duration: req.Duration,
	}, nil
}

// GeneratePromoVideo — сценарий (через LLM) + запуск Veo 3.
func (s *Service) GeneratePromoVideo(ctx context.Context, projectName, spec string) (*PromoVideo, error) {
	l := ports.LoggerFromContext(ctx)
	l.InfoContext(ctx, "promo video generation started", "projectName", projectName)

	video := s.defaultPromoVideo(projectName)
	s.applyPromoScriptFromLLM(ctx, video, projectName, spec)

	// Запуск Veo 3 (best-effort).
	if os.Getenv("VEO_ENABLED") == "1" {
		result, err := s.GenerateVideoVeo(ctx, VeoRequest{
			Prompt:   video.Voiceover,
			Duration: video.Duration,
			Style:    "cinematic",
			Aspect:   AspectRatio16x9,
		})
		if err == nil && result.VideoURL != "" {
			video.VideoURL = result.VideoURL
			l.InfoContext(ctx, "veo3 video generated", "url", result.VideoURL)
		} else if err != nil {
			l.WarnContext(ctx, "veo3 generation failed", "error", wrapper.Wrap(err))
		}
	}

	return video, nil
}

func (s *Service) enrichPromptAssetsFromLLM(ctx context.Context, assets *Assets, projectName, spec string, colors []string) {
	l := ports.LoggerFromContext(ctx)
	if s.llm == nil {
		assets.VideoPrompts = defaultVideoPrompts(projectName)

		return
	}
	synthesized, err := s.synthesizePrompts(ctx, projectName, spec, colors)
	if err == nil {
		s.mergeAssets(assets, synthesized)
	} else {
		l.WarnContext(ctx, "prompt synthesis failed, using defaults", "error", wrapper.Wrap(err))
	}
	videoPrompts, err := s.synthesizeVideoVariants(ctx, projectName, spec)
	if err == nil {
		assets.VideoPrompts = videoPrompts

		return
	}
	l.WarnContext(ctx, "video variants failed, using defaults", "error", wrapper.Wrap(err))
	assets.VideoPrompts = defaultVideoPrompts(projectName)
}

func defaultVideoPrompts(projectName string) []string {
	return []string{
		fmt.Sprintf("Cinematic 30-second promo for %s. Dark tech aesthetic, smooth camera movements, modern UI showcase.", projectName),
		fmt.Sprintf("Dynamic product demo of %s. Split-screen transitions, code-to-visual morphs, energetic electronic music.", projectName),
		fmt.Sprintf("Minimalist brand story for %s. Soft gradients, typography animation, ambient soundtrack, premium feel.", projectName),
	}
}

func (s *Service) applyPromoScriptFromLLM(ctx context.Context, video *PromoVideo, projectName, spec string) {
	if s.llm == nil {
		return
	}
	prompt := fmt.Sprintf(`Create a 30-second promo video script for %q. Spec: %s.
Return ONLY JSON:
{
  "script": "...", "duration": "30s",
  "scenes": ["Scene 1: ...","Scene 2: ...","Scene 3: ..."],
  "voiceover": "...", "music_style": "..."
}`, projectName, spec)

	resp, err := s.llm.Complete(ctx, ports.LLMRequest{
		Model:       ModelPromptAnthropic,
		UserPrompt:  prompt,
		MaxTokens:   1024,
		Temperature: 0.6,
	})
	if err != nil {
		return
	}
	var parsed PromoVideo
	if json.Unmarshal([]byte(stripFences(resp.Content)), &parsed) != nil {
		return
	}
	if parsed.Script != "" {
		video.Script = parsed.Script
	}
	if parsed.Duration != "" {
		video.Duration = parsed.Duration
	}
	if len(parsed.Scenes) > 0 {
		video.Scenes = parsed.Scenes
	}
	if parsed.Voiceover != "" {
		video.Voiceover = parsed.Voiceover
	}
	if parsed.MusicStyle != "" {
		video.MusicStyle = parsed.MusicStyle
	}
}

// synthesizeVideoVariants — через LLM получает 3 варианта промо-ролика.
func (s *Service) synthesizeVideoVariants(ctx context.Context, projectName, spec string) ([]string, error) {
	prompt := fmt.Sprintf(`Generate exactly 3 different creative prompts for a 30-second promo video for %q.
Spec: %s

Each prompt should describe the visual style, camera work, transitions, and mood.
Return ONLY a JSON array of 3 strings, no other text:
["prompt 1", "prompt 2", "prompt 3"]`, projectName, spec)

	resp, err := s.llm.Complete(ctx, ports.LLMRequest{
		Model:       ModelPromptAnthropic,
		UserPrompt:  prompt,
		MaxTokens:   1024,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	body := stripFences(resp.Content)
	var variants []string
	err = json.Unmarshal([]byte(body), &variants)
	if err != nil {
		return nil, fmt.Errorf("parse video variants: %w", err)
	}
	if len(variants) < 3 {
		return nil, fmt.Errorf("%w, got %d", ErrExpectedThreeVideoVars, len(variants))
	}

	return variants[:3], nil
}

// synthesizePrompts — через ports.LLMProvider (Anthropic) получает JSON-ассеты.
func (s *Service) synthesizePrompts(ctx context.Context, projectName, spec string, colors []string) (*Assets, error) {
	colorCtx := strings.Join(colors, ", ")
	if colorCtx == "" {
		colorCtx = "#5b4cdb, #0e0e11, #ffffff"
	}
	prompt := fmt.Sprintf(`Create design assets for %q. Spec: %s. Colors: %s.
Return ONLY JSON:
{
  "logo_svg": "<svg ...>",
  "color_palette": ["#primary","#secondary","#accent","#background","#foreground"],
  "icon_set": {"home":"M...","star":"M...","check":"M..."},
  "hero_prompt": "detailed image prompt for hero background",
  "og_image_prompt": "detailed prompt for Open Graph preview"
}`, projectName, spec, colorCtx)

	resp, err := s.llm.Complete(ctx, ports.LLMRequest{
		Model:       ModelPromptAnthropic,
		UserPrompt:  prompt,
		MaxTokens:   2048,
		Temperature: 0.4,
	})
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	body := stripFences(resp.Content)
	var parsed Assets
	err = json.Unmarshal([]byte(body), &parsed)
	if err != nil {
		return nil, fmt.Errorf("parse assets JSON: %w", err)
	}

	return &parsed, nil
}

func (s *Service) mergeAssets(dst, src *Assets) {
	if src.LogoSVG != "" {
		dst.LogoSVG = src.LogoSVG
	}
	if len(src.ColorPalette) > 0 {
		dst.ColorPalette = src.ColorPalette
	}
	if len(src.IconSet) > 0 {
		dst.IconSet = src.IconSet
	}
	if src.HeroPrompt != "" {
		dst.HeroPrompt = src.HeroPrompt
	}
	if src.OGImagePrompt != "" {
		dst.OGImagePrompt = src.OGImagePrompt
	}
}

// ──────────────────────────────────────────────────────────────
//  REPLICATE HTTP (create + poll)
// ──────────────────────────────────────────────────────────────

type replicatePrediction struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Output any    `json:"output"`
	Error  any    `json:"error"`
	URLs   struct {
		Get string `json:"get"`
	} `json:"urls"`
}

func (s *Service) replicateCreate(ctx context.Context, endpoint string, payload []byte, preferWait bool) (*replicatePrediction, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	req.Header.Set("Authorization", "Bearer "+s.replicateToken)
	req.Header.Set("Content-Type", "application/json")
	if preferWait {
		req.Header.Set("Prefer", "wait")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("replicate create: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		maxLog := min(len(body), 300)

		return nil, fmt.Errorf("%w %d: %s", ErrReplicateHTTPError, resp.StatusCode, string(body[:maxLog]))
	}
	var pred replicatePrediction
	err = json.Unmarshal(body, &pred)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if pred.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrReplicatePredictionError, pred.Error)
	}

	return &pred, nil
}

func (s *Service) replicatePoll(ctx context.Context, pred *replicatePrediction, timeout, interval time.Duration) (*replicatePrediction, error) {
	pollURL := pred.URLs.Get
	if pollURL == "" {
		pollURL = fmt.Sprintf("%s/predictions/%s", replicateAPIBase, pred.ID)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	deadline := time.After(timeout)

	for {
		select {
		case <-ctx.Done():
			return nil, wrapper.Wrap(ctx.Err())
		case <-deadline:
			return nil, fmt.Errorf("%w (id=%s)", ErrReplicatePollTimeout, pred.ID)
		case <-ticker.C:
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, pollURL, nil)
			req.Header.Set("Authorization", "Bearer "+s.replicateToken)
			resp, err := s.httpClient.Do(req)
			if err != nil {
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			var poll replicatePrediction
			err = json.Unmarshal(body, &poll)
			if err != nil {
				continue
			}
			switch poll.Status {
			case "succeeded":
				return &poll, nil
			case "failed", "canceled":
				return nil, fmt.Errorf("%w %s: %v", ErrReplicatePollStatus, poll.Status, poll.Error)
			}
		}
	}
}

// ──────────────────────────────────────────────────────────────
//  Helpers
// ──────────────────────────────────────────────────────────────

func extractURL(output any) string {
	if s, ok := output.(string); ok {
		return s
	}
	if arr, ok := output.([]any); ok && len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			return s
		}
	}

	return ""
}

func stripFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")

	return strings.TrimSpace(s)
}

func aspectRatio(w, h int) string {
	if w <= 0 || h <= 0 {
		return AspectRatio16x9
	}
	switch {
	case w*9 == h*16:
		return AspectRatio16x9
	case w*16 == h*9:
		return "9:16"
	case w == h:
		return "1:1"
	case w*2 == h*3:
		return "3:2"
	case w*3 == h*2:
		return "2:3"
	}

	return AspectRatio16x9
}

func (s *Service) defaultAssets(name string, colors []string) *Assets {
	palette := colors
	if len(palette) == 0 {
		palette = []string{"#5b4cdb", "#0e0e11", "#ffffff", "#f0f0f5", "#8b7cf8"}
	}

	return &Assets{
		LogoSVG: fmt.Sprintf(
			`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="50" cy="50" r="45" fill="%s"/><text x="50" y="65" font-size="40" text-anchor="middle" fill="white" font-family="Inter">И</text></svg>`,
			palette[0],
		),
		ColorPalette: palette,
		IconSet: map[string]string{
			"home":  "M10 20 L50 5 L90 20 L90 90 L10 90 Z",
			"star":  "M50 5 L61 35 L95 35 L68 57 L79 91 L50 70 L21 91 L32 57 L5 35 L39 35 Z",
			"check": "M10 50 L35 75 L90 20",
		},
		HeroPrompt:    fmt.Sprintf("Futuristic dark tech background for %s, gradient mesh, purple and blue tones", name),
		OGImagePrompt: fmt.Sprintf("Professional social preview for %s app, dark theme, modern typography", name),
		GeneratedAt:   time.Now(),
	}
}

func (s *Service) defaultPromoVideo(name string) *PromoVideo {
	return &PromoVideo{
		Script:      fmt.Sprintf("Introducing %s — the future of AI-powered development.", name),
		Duration:    "30s",
		Scenes:      []string{"Scene 1: Dark intro with logo reveal", "Scene 2: Feature showcase", "Scene 3: CTA"},
		Voiceover:   fmt.Sprintf("Meet %s. AI that builds your vision.", name),
		MusicStyle:  "Epic Electronic Cinematic",
		VideoURL:    "",
		GeneratedAt: time.Now(),
	}
}
