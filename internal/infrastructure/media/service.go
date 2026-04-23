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
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Media Service V2
//  Nano Banana 2 (FLUX image gen) + Veo (video gen hooks)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	ModelDesigner     = "google/gemini-3-pro"   // Gemini 3 Pro — UI-ассеты (Replicate)
	ModelVideographer = "google/gemini-3.1-pro" // Gemini 3.1 Pro — промо-видео (Replicate)

	// Nano Banana 2 — text-to-image via Replicate FLUX
	DefaultImageModel = "black-forest-labs/flux-1.1-pro"
)

// MediaAssets результат генерации медиа-ассетов
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

// PromoVideo результат генерации промо-видео
type PromoVideo struct {
	Script      string    `json:"script"`
	Duration    string    `json:"duration"`
	Scenes      []string  `json:"scenes"`
	Voiceover   string    `json:"voiceover"`
	MusicStyle  string    `json:"music_style"`
	VideoURL    string    `json:"video_url,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

// VeoRequest запрос на генерацию видео через Veo
type VeoRequest struct {
	Prompt   string `json:"prompt"`
	Duration string `json:"duration"` // "5s", "10s", "15s"
	Style    string `json:"style"`    // "cinematic", "animated", "realistic"
	Aspect   string `json:"aspect"`   // "16:9", "9:16", "1:1"
}

// VeoResult результат генерации видео
type VeoResult struct {
	VideoURL string `json:"video_url"`
	Status   string `json:"status"` // "pending", "processing", "completed", "failed"
	Duration string `json:"duration"`
	Error    string `json:"error,omitempty"`
}

// MediaService сервис для генерации медиа-контента
type MediaService struct {
	apiKey     string
	baseURL    string
	imageModel string
	httpClient *http.Client
}

// NewMediaService создает новый медиа-сервис
func NewMediaService(apiKey string) *MediaService {
	imgModel := os.Getenv("IMAGE_MODEL_ID")
	if imgModel == "" {
		imgModel = DefaultImageModel
	}
	return &MediaService{
		apiKey:     apiKey,
		baseURL:    "https://openrouter.ai/api/v1",
		imageModel: imgModel,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// GenerateUIAssets генерирует UI-ассеты через Nano Banana 2
func (s *MediaService) GenerateUIAssets(ctx context.Context, projectName, spec string, colors []string) (*MediaAssets, error) {
	log.Printf("🎨 MediaService: генерация UI-ассетов для '%s'", projectName)

	colorCtx := strings.Join(colors, ", ")
	if colorCtx == "" {
		colorCtx = "#5b4cdb, #0e0e11, #ffffff"
	}

	prompt := fmt.Sprintf(`You are a professional UI/UX designer. Create design assets for a project called "%s".

Project specification: %s
Color palette: %s

Return ONLY a valid JSON object:
{
  "logo_svg": "<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'>...minimal logo SVG...</svg>",
  "color_palette": ["#primary", "#secondary", "#accent", "#background", "#foreground"],
  "icon_set": {
    "home": "M10 20 L50 5 L90 20 L90 90 L10 90 Z",
    "star": "M50 5 L61 35 L95 35 L68 57 L79 91 L50 70 L21 91 L32 57 L5 35 L39 35 Z",
    "check": "M10 50 L35 75 L90 20"
  },
  "hero_prompt": "detailed image generation prompt for hero section background",
  "og_image_prompt": "detailed prompt for Open Graph social preview image"
}

Return ONLY the JSON, no explanation.`, projectName, spec, colorCtx)

	content, err := s.callLLM(ctx, ModelDesigner, prompt, 2048)
	if err != nil {
		log.Printf("⚠️ MediaService UI Assets error: %v — using defaults", err)
		return s.defaultAssets(projectName, colors), nil
	}

	assets := s.parseMediaAssets(content, colors)

	// ── Nano Banana 2: генерация реальных изображений ──
	if assets.HeroPrompt != "" {
		heroURL, err := s.GenerateImage(ctx, assets.HeroPrompt, 1344, 768)
		if err != nil {
			log.Printf("⚠️ Nano Banana 2: hero image failed: %v", err)
		} else {
			assets.HeroImageURL = heroURL
			log.Printf("✅ Nano Banana 2: hero image → %s", heroURL)
		}
	}
	if assets.OGImagePrompt != "" {
		ogURL, err := s.GenerateImage(ctx, assets.OGImagePrompt, 1200, 630)
		if err != nil {
			log.Printf("⚠️ Nano Banana 2: OG image failed: %v", err)
		} else {
			assets.OGImageURL = ogURL
			log.Printf("✅ Nano Banana 2: OG image → %s", ogURL)
		}
	}

	log.Printf("✅ MediaService: UI-ассеты сгенерированы для '%s'", projectName)
	return assets, nil
}

// GenerateImage генерирует изображение через Nano Banana 2 (Replicate FLUX)
func (s *MediaService) GenerateImage(ctx context.Context, prompt string, width, height int) (string, error) {
	token := os.Getenv("REPLICATE_API_TOKEN")
	if token == "" {
		return "", fmt.Errorf("REPLICATE_API_TOKEN not set")
	}

	log.Printf("🎨 Nano Banana 2: generating %dx%d image...", width, height)

	endpoint := fmt.Sprintf("https://api.replicate.com/v1/models/%s/predictions", s.imageModel)
	payload, _ := json.Marshal(map[string]interface{}{
		"input": map[string]interface{}{
			"prompt":              prompt,
			"width":               width,
			"height":              height,
			"num_inference_steps": 28,
			"guidance_scale":      3.5,
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "wait")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("image request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", fmt.Errorf("Nano Banana 2 API error (HTTP %d): %s", resp.StatusCode, string(respBody[:min(len(respBody), 300)]))
	}

	var pred struct {
		ID     string      `json:"id"`
		Status string      `json:"status"`
		Output interface{} `json:"output"`
		Error  interface{} `json:"error"`
		URLs   struct {
			Get string `json:"get"`
		} `json:"urls"`
	}
	json.Unmarshal(respBody, &pred)

	// If Prefer:wait resolved immediately
	if pred.Status == "succeeded" {
		url := extractImageURL(pred.Output)
		if url != "" {
			return url, nil
		}
	}
	if pred.Error != nil {
		return "", fmt.Errorf("Nano Banana 2 error: %v", pred.Error)
	}

	// Poll for completion
	pollURL := pred.URLs.Get
	if pollURL == "" {
		pollURL = fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", pred.ID)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	timeout := time.After(2 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timeout:
			return "", fmt.Errorf("Nano Banana 2 timed out (id=%s)", pred.ID)
		case <-ticker.C:
			pollReq, _ := http.NewRequestWithContext(ctx, "GET", pollURL, nil)
			pollReq.Header.Set("Authorization", "Bearer "+token)
			pollResp, err := client.Do(pollReq)
			if err != nil {
				continue
			}
			pollBody, _ := io.ReadAll(pollResp.Body)
			pollResp.Body.Close()

			var poll struct {
				Status string      `json:"status"`
				Output interface{} `json:"output"`
				Error  interface{} `json:"error"`
			}
			json.Unmarshal(pollBody, &poll)

			switch poll.Status {
			case "succeeded":
				url := extractImageURL(poll.Output)
				if url != "" {
					log.Printf("✅ Nano Banana 2: image ready (id=%s)", pred.ID)
					return url, nil
				}
				return "", fmt.Errorf("empty image output (id=%s)", pred.ID)
			case "failed", "canceled":
				return "", fmt.Errorf("Nano Banana 2 %s: %v", poll.Status, poll.Error)
			}
		}
	}
}

// extractImageURL extracts URL from Replicate image model output
func extractImageURL(output interface{}) string {
	// FLUX returns a single URL string
	if s, ok := output.(string); ok {
		return s
	}
	// Some models return []string
	if arr, ok := output.([]interface{}); ok && len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			return s
		}
	}
	return ""
}

// ──────────────────────────────────────────────────────────────────
// VEO VIDEO GENERATION HOOKS
// ──────────────────────────────────────────────────────────────────

// GenerateVideoVeo запускает генерацию промо-видео через Google Veo API
// TODO: Подключить когда Veo API станет доступен
// Сейчас: возвращает промпт + сценарий для ручной генерации
func (s *MediaService) GenerateVideoVeo(ctx context.Context, req VeoRequest) (*VeoResult, error) {
	veoEndpoint := os.Getenv("VEO_API_ENDPOINT")
	veoKey := os.Getenv("VEO_API_KEY")

	if veoEndpoint == "" || veoKey == "" {
		log.Printf("🎬 Veo: API not configured — returning prompt-only result")
		return &VeoResult{
			Status:   "pending",
			Duration: req.Duration,
			VideoURL: "",
		}, nil
	}

	// ── Veo API call (ready to connect) ──
	log.Printf("🎬 Veo: generating %s video, style=%s aspect=%s", req.Duration, req.Style, req.Aspect)

	payload, _ := json.Marshal(map[string]interface{}{
		"prompt":   req.Prompt,
		"duration": req.Duration,
		"style":    req.Style,
		"aspect":   req.Aspect,
	})

	httpReq, err := http.NewRequestWithContext(ctx, "POST", veoEndpoint+"/generate", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+veoKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Veo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("Veo API error (HTTP %d): %s", resp.StatusCode, string(body[:min(len(body), 300)]))
	}

	var result VeoResult
	json.Unmarshal(body, &result)
	log.Printf("✅ Veo: video generation started, status=%s", result.Status)
	return &result, nil
}

// GeneratePromoVideo генерирует сценарий и описание промо-видео через Veo
func (s *MediaService) GeneratePromoVideo(ctx context.Context, projectName, spec string) (*PromoVideo, error) {
	log.Printf("🎬 MediaService: генерация промо-видео для '%s'", projectName)

	prompt := fmt.Sprintf(`You are a professional video producer and scriptwriter. Create a promo video concept for "%s".

Project: %s

Return ONLY a valid JSON object:
{
  "script": "Full video script with narrator text",
  "duration": "30s",
  "scenes": [
    "Scene 1: Opening shot — ...",
    "Scene 2: Feature showcase — ...",
    "Scene 3: Call to action — ..."
  ],
  "voiceover": "Complete voiceover text for the video",
  "music_style": "Epic cinematic / Electronic / Ambient",
  "video_url": ""
}

Return ONLY the JSON.`, projectName, spec)

	content, err := s.callLLM(ctx, ModelVideographer, prompt, 1024)
	if err != nil {
		log.Printf("⚠️ MediaService Promo Video error: %v — using defaults", err)
		return s.defaultPromoVideo(projectName), nil
	}

	video := s.parsePromoVideo(content, projectName)
	log.Printf("✅ MediaService: промо-видео концепт готов для '%s'", projectName)
	return video, nil
}

// callLLM выполняет запрос к LLM — Google модели через Replicate, остальные через OpenRouter
func (s *MediaService) callLLM(ctx context.Context, model, prompt string, maxTokens int) (string, error) {
	// Route Google models through Replicate (banned on OpenRouter)
	if strings.HasPrefix(model, "google/") || strings.HasPrefix(model, "anthropic/") {
		return s.callReplicate(ctx, model, prompt, maxTokens)
	}
	return s.callOpenRouterLLM(ctx, model, prompt, maxTokens)
}

// callReplicate вызывает модель через Replicate Predictions API с async polling
func (s *MediaService) callReplicate(ctx context.Context, model, prompt string, maxTokens int) (string, error) {
	token := os.Getenv("REPLICATE_API_TOKEN")
	if token == "" {
		return "", fmt.Errorf("REPLICATE_API_TOKEN not set")
	}

	if maxTokens < 1024 {
		maxTokens = 1024
	}

	endpoint := fmt.Sprintf("https://api.replicate.com/v1/models/%s/predictions", model)
	payload, _ := json.Marshal(map[string]interface{}{
		"input": map[string]interface{}{
			"prompt":      prompt,
			"max_tokens":  maxTokens,
			"temperature": 0.7,
		},
	})

	log.Printf("🔗 MediaService Replicate: %s (%d bytes)", model, len(payload))

	// Create prediction
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.Printf("🚨 MediaService Replicate error | model=%s status=%d | %s", model, resp.StatusCode, string(respBody[:min(len(respBody), 300)]))
		return "", fmt.Errorf("Replicate API error (HTTP %d)", resp.StatusCode)
	}

	var pred struct {
		ID     string      `json:"id"`
		Status string      `json:"status"`
		Output interface{} `json:"output"`
		Error  interface{} `json:"error"`
		URLs   struct {
			Get string `json:"get"`
		} `json:"urls"`
	}
	json.Unmarshal(respBody, &pred)

	if pred.Status == "succeeded" {
		return extractOutput(pred.Output), nil
	}
	if pred.Error != nil {
		return "", fmt.Errorf("Replicate error: %v", pred.Error)
	}

	// Poll for completion
	pollURL := pred.URLs.Get
	if pollURL == "" {
		pollURL = fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", pred.ID)
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timeout:
			return "", fmt.Errorf("Replicate prediction timed out (id=%s)", pred.ID)
		case <-ticker.C:
			pollReq, _ := http.NewRequestWithContext(ctx, "GET", pollURL, nil)
			pollReq.Header.Set("Authorization", "Bearer "+token)
			pollResp, err := client.Do(pollReq)
			if err != nil {
				continue
			}
			pollBody, _ := io.ReadAll(pollResp.Body)
			pollResp.Body.Close()

			var poll struct {
				Status string      `json:"status"`
				Output interface{} `json:"output"`
				Error  interface{} `json:"error"`
			}
			json.Unmarshal(pollBody, &poll)

			switch poll.Status {
			case "succeeded":
				out := extractOutput(poll.Output)
				log.Printf("✅ MediaService Replicate: %s → %d chars", model, len(out))
				return out, nil
			case "failed", "canceled":
				return "", fmt.Errorf("Replicate prediction %s: %v", poll.Status, poll.Error)
			}
		}
	}
}

// extractOutput handles Replicate output (string or []string)
func extractOutput(output interface{}) string {
	if s, ok := output.(string); ok {
		return s
	}
	if arr, ok := output.([]interface{}); ok {
		var sb strings.Builder
		for _, chunk := range arr {
			if s, ok := chunk.(string); ok {
				sb.WriteString(s)
			}
		}
		return sb.String()
	}
	b, _ := json.Marshal(output)
	return string(b)
}

// callOpenRouterLLM выполняет запрос к OpenRouter (DeepSeek, Qwen)
func (s *MediaService) callOpenRouterLLM(ctx context.Context, model, prompt string, maxTokens int) (string, error) {
	if s.apiKey == "" || strings.HasPrefix(s.apiKey, "MISSING") {
		return "", fmt.Errorf("OPENROUTER_API_KEY не установлен")
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  maxTokens,
		"temperature": 0.7,
	})

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("HTTP-Referer", "https://istok-agent-core.vercel.app")
	httpReq.Header.Set("X-Title", "ИСТОК Агент")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("🚨 MediaService error | model=%s status=%d | %s", model, resp.StatusCode, string(body))
		return "", fmt.Errorf("API error (HTTP %d)", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return result.Choices[0].Message.Content, nil
}

// parseMediaAssets парсит JSON ответ
func (s *MediaService) parseMediaAssets(content string, colors []string) *MediaAssets {
	assets := s.defaultAssets("Project", colors)

	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var parsed struct {
		LogoSVG       string            `json:"logo_svg"`
		ColorPalette  []string          `json:"color_palette"`
		IconSet       map[string]string `json:"icon_set"`
		HeroPrompt    string            `json:"hero_prompt"`
		OGImagePrompt string            `json:"og_image_prompt"`
	}

	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		log.Printf("⚠️ MediaService: не удалось распарсить assets JSON: %v", err)
		return assets
	}

	if parsed.LogoSVG != "" {
		assets.LogoSVG = parsed.LogoSVG
	}
	if len(parsed.ColorPalette) > 0 {
		assets.ColorPalette = parsed.ColorPalette
	}
	if len(parsed.IconSet) > 0 {
		assets.IconSet = parsed.IconSet
	}
	if parsed.HeroPrompt != "" {
		assets.HeroPrompt = parsed.HeroPrompt
	}
	if parsed.OGImagePrompt != "" {
		assets.OGImagePrompt = parsed.OGImagePrompt
	}

	return assets
}

// parsePromoVideo парсит JSON ответ для видео
func (s *MediaService) parsePromoVideo(content, projectName string) *PromoVideo {
	video := s.defaultPromoVideo(projectName)

	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var parsed struct {
		Script     string   `json:"script"`
		Duration   string   `json:"duration"`
		Scenes     []string `json:"scenes"`
		Voiceover  string   `json:"voiceover"`
		MusicStyle string   `json:"music_style"`
		VideoURL   string   `json:"video_url"`
	}

	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		log.Printf("⚠️ MediaService: не удалось распарсить video JSON: %v", err)
		return video
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
	if parsed.VideoURL != "" {
		video.VideoURL = parsed.VideoURL
	}

	return video
}

// defaultAssets возвращает дефолтные ассеты при ошибке
func (s *MediaService) defaultAssets(name string, colors []string) *MediaAssets {
	palette := colors
	if len(palette) == 0 {
		palette = []string{"#5b4cdb", "#0e0e11", "#ffffff", "#f0f0f5", "#8b7cf8"}
	}
	return &MediaAssets{
		LogoSVG:      fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="50" cy="50" r="45" fill="%s"/><text x="50" y="65" font-size="40" text-anchor="middle" fill="white" font-family="Inter">И</text></svg>`, palette[0]),
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

// defaultPromoVideo возвращает дефолтный концепт видео
func (s *MediaService) defaultPromoVideo(name string) *PromoVideo {
	return &PromoVideo{
		Script:      fmt.Sprintf("Introducing %s — the future of AI-powered development. Build faster, smarter, better.", name),
		Duration:    "30s",
		Scenes:      []string{"Scene 1: Dark intro with logo reveal", "Scene 2: Feature showcase with UI animation", "Scene 3: CTA — Start building today"},
		Voiceover:   fmt.Sprintf("Meet %s. The AI agent that builds your vision in seconds.", name),
		MusicStyle:  "Epic Electronic Cinematic",
		VideoURL:    "",
		GeneratedAt: time.Now(),
	}
}
