package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/istok/agent-core/internal/ports"
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

	// Default text model for prompt synthesis (if LLMProvider supplied).
	ModelPromptAnthropic = "anthropic/claude-3-7-sonnet"

	replicateAPIBase = "https://api.replicate.com/v1"
)

// MediaAssets — результат генерации UI-ассетов.
type MediaAssets struct {
	LogoSVG       string            `json:"logo_svg"`
	ColorPalette  []string          `json:"color_palette"`
	IconSet       map[string]string `json:"icon_set"`
	HeroPrompt    string            `json:"hero_prompt"`
	OGImagePrompt string            `json:"og_image_prompt"`
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

// MediaService генерирует медиа через Replicate (nano-banana + Veo 3).
// LLMProvider опционален — используется только для генерации промптов/сценариев.
type MediaService struct {
	replicateToken string
	imageModel     string
	videoModel     string
	llm            ports.LLMProvider // optional — nil-safe
	httpClient     *http.Client
}

// NewMediaService — конструктор без LLM (использует шаблонные ассеты).
func NewMediaService(replicateToken string) *MediaService {
	return NewMediaServiceWithLLM(replicateToken, nil)
}

// NewMediaServiceWithLLM — конструктор с LLMProvider для синтеза промптов.
func NewMediaServiceWithLLM(replicateToken string, llm ports.LLMProvider) *MediaService {
	img := os.Getenv("IMAGE_MODEL_ID")
	if img == "" {
		img = ModelImageNanoBanana
	}
	vid := os.Getenv("VIDEO_MODEL_ID")
	if vid == "" {
		vid = ModelVideoVeo3
	}
	return &MediaService{
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
func (s *MediaService) GenerateUIAssets(ctx context.Context, projectName, spec string, colors []string) (*MediaAssets, error) {
	log.Printf("🎨 MediaService: UI-assets for %q", projectName)

	assets := s.defaultAssets(projectName, colors)

	// 1) Если LLM есть — обновляем logo_svg / icon_set / hero_prompt / og_prompt.
	if s.llm != nil {
		if synthesized, err := s.synthesizePrompts(ctx, projectName, spec, colors); err == nil {
			s.mergeAssets(assets, synthesized)
		} else {
			log.Printf("⚠️ MediaService: prompt synthesis failed, using defaults: %v", err)
		}
	}

	// 2) Генерация hero (Replicate nano-banana).
	if assets.HeroPrompt != "" {
		if url, err := s.GenerateImage(ctx, assets.HeroPrompt, 1344, 768); err == nil {
			assets.HeroImageURL = url
			log.Printf("✅ nano-banana: hero → %s", url)
		} else {
			log.Printf("⚠️ nano-banana hero: %v", err)
		}
	}
	// 3) Генерация OG.
	if assets.OGImagePrompt != "" {
		if url, err := s.GenerateImage(ctx, assets.OGImagePrompt, 1200, 630); err == nil {
			assets.OGImageURL = url
			log.Printf("✅ nano-banana: OG → %s", url)
		} else {
			log.Printf("⚠️ nano-banana OG: %v", err)
		}
	}

	return assets, nil
}

// synthesizePrompts — через ports.LLMProvider (Anthropic) получает JSON-ассеты.
func (s *MediaService) synthesizePrompts(ctx context.Context, projectName, spec string, colors []string) (*MediaAssets, error) {
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
		return nil, err
	}

	body := stripFences(resp.Content)
	var parsed MediaAssets
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return nil, fmt.Errorf("parse assets JSON: %w", err)
	}
	return &parsed, nil
}

func (s *MediaService) mergeAssets(dst, src *MediaAssets) {
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
//  IMAGE — Replicate nano-banana
// ──────────────────────────────────────────────────────────────

// GenerateImage — text→image через Replicate nano-banana.
func (s *MediaService) GenerateImage(ctx context.Context, prompt string, width, height int) (string, error) {
	if s.replicateToken == "" {
		return "", fmt.Errorf("REPLICATE_API_TOKEN not set")
	}
	log.Printf("🎨 nano-banana: %dx%d", width, height)

	endpoint := fmt.Sprintf("%s/models/%s/predictions", replicateAPIBase, s.imageModel)
	payload, _ := json.Marshal(map[string]interface{}{
		"input": map[string]interface{}{
			"prompt":          prompt,
			"aspect_ratio":    aspectRatio(width, height),
			"output_format":   "png",
			"safety_filter":   "block_only_high",
		},
	})

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
func (s *MediaService) GenerateVideoVeo(ctx context.Context, req VeoRequest) (*VeoResult, error) {
	if s.replicateToken == "" {
		return nil, fmt.Errorf("REPLICATE_API_TOKEN not set")
	}
	log.Printf("🎬 Veo 3: %s %s %s", req.Duration, req.Style, req.Aspect)

	aspect := req.Aspect
	if aspect == "" {
		aspect = "16:9"
	}
	endpoint := fmt.Sprintf("%s/models/%s/predictions", replicateAPIBase, s.videoModel)
	payload, _ := json.Marshal(map[string]interface{}{
		"input": map[string]interface{}{
			"prompt":         req.Prompt,
			"aspect_ratio":   aspect,
			"duration":       req.Duration,
			"style":          req.Style,
			"negative_prompt": "low quality, blurry, watermark",
		},
	})

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
func (s *MediaService) GeneratePromoVideo(ctx context.Context, projectName, spec string) (*PromoVideo, error) {
	log.Printf("🎬 MediaService: promo video for %q", projectName)

	video := s.defaultPromoVideo(projectName)

	// LLM-генерация сценария (если доступен).
	if s.llm != nil {
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
		if err == nil {
			var parsed PromoVideo
			if e := json.Unmarshal([]byte(stripFences(resp.Content)), &parsed); e == nil {
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
		}
	}

	// Запуск Veo 3 (best-effort).
	if os.Getenv("VEO_ENABLED") == "1" {
		result, err := s.GenerateVideoVeo(ctx, VeoRequest{
			Prompt:   video.Voiceover,
			Duration: video.Duration,
			Style:    "cinematic",
			Aspect:   "16:9",
		})
		if err == nil && result.VideoURL != "" {
			video.VideoURL = result.VideoURL
			log.Printf("✅ Veo 3: video → %s", result.VideoURL)
		} else if err != nil {
			log.Printf("⚠️ Veo 3: %v", err)
		}
	}

	return video, nil
}

// ──────────────────────────────────────────────────────────────
//  REPLICATE HTTP (create + poll)
// ──────────────────────────────────────────────────────────────

type replicatePrediction struct {
	ID     string      `json:"id"`
	Status string      `json:"status"`
	Output interface{} `json:"output"`
	Error  interface{} `json:"error"`
	URLs   struct {
		Get string `json:"get"`
	} `json:"urls"`
}

func (s *MediaService) replicateCreate(ctx context.Context, endpoint string, payload []byte, preferWait bool) (*replicatePrediction, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
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
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		maxLog := len(body)
		if maxLog > 300 {
			maxLog = 300
		}
		return nil, fmt.Errorf("replicate HTTP %d: %s", resp.StatusCode, string(body[:maxLog]))
	}
	var pred replicatePrediction
	if err := json.Unmarshal(body, &pred); err != nil {
		return nil, err
	}
	if pred.Error != nil {
		return nil, fmt.Errorf("replicate error: %v", pred.Error)
	}
	return &pred, nil
}

func (s *MediaService) replicatePoll(ctx context.Context, pred *replicatePrediction, timeout, interval time.Duration) (*replicatePrediction, error) {
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
			return nil, ctx.Err()
		case <-deadline:
			return nil, fmt.Errorf("replicate poll timeout (id=%s)", pred.ID)
		case <-ticker.C:
			req, _ := http.NewRequestWithContext(ctx, "GET", pollURL, nil)
			req.Header.Set("Authorization", "Bearer "+s.replicateToken)
			resp, err := s.httpClient.Do(req)
			if err != nil {
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			var poll replicatePrediction
			if err := json.Unmarshal(body, &poll); err != nil {
				continue
			}
			switch poll.Status {
			case "succeeded":
				return &poll, nil
			case "failed", "canceled":
				return nil, fmt.Errorf("replicate %s: %v", poll.Status, poll.Error)
			}
		}
	}
}

// ──────────────────────────────────────────────────────────────
//  Helpers
// ──────────────────────────────────────────────────────────────

func extractURL(output interface{}) string {
	if s, ok := output.(string); ok {
		return s
	}
	if arr, ok := output.([]interface{}); ok && len(arr) > 0 {
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
		return "16:9"
	}
	switch {
	case w*9 == h*16:
		return "16:9"
	case w*16 == h*9:
		return "9:16"
	case w == h:
		return "1:1"
	case w*2 == h*3:
		return "3:2"
	case w*3 == h*2:
		return "2:3"
	}
	return "16:9"
}

func (s *MediaService) defaultAssets(name string, colors []string) *MediaAssets {
	palette := colors
	if len(palette) == 0 {
		palette = []string{"#5b4cdb", "#0e0e11", "#ffffff", "#f0f0f5", "#8b7cf8"}
	}
	return &MediaAssets{
		LogoSVG: fmt.Sprintf(
			`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="50" cy="50" r="45" fill="%s"/><text x="50" y="65" font-size="40" text-anchor="middle" fill="white" font-family="Inter">И</text></svg>`,
			palette[0]),
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

func (s *MediaService) defaultPromoVideo(name string) *PromoVideo {
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
