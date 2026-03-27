package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Media Service
//  Nano Banana 2 (UI ассеты) + Veo (промо-видео)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	ModelNanoBanana = "google/gemini-3.1-flash-image-preview" // Nano Banana 2 — UI-ассеты
	ModelVeo        = "google/veo-3.1"                        // Veo 3.1 — промо-видео
)

// MediaAssets результат генерации медиа-ассетов
type MediaAssets struct {
	LogoSVG       string            `json:"logo_svg"`
	ColorPalette  []string          `json:"color_palette"`
	IconSet       map[string]string `json:"icon_set"`
	HeroPrompt    string            `json:"hero_prompt"`
	OGImagePrompt string            `json:"og_image_prompt"`
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

// MediaService сервис для генерации медиа-контента
type MediaService struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewMediaService создает новый медиа-сервис
func NewMediaService(apiKey string) *MediaService {
	return &MediaService{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
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

	content, err := s.callLLM(ctx, ModelNanoBanana, prompt, 2048)
	if err != nil {
		log.Printf("⚠️ MediaService UI Assets error: %v — using defaults", err)
		return s.defaultAssets(projectName, colors), nil
	}

	assets := s.parseMediaAssets(content, colors)
	log.Printf("✅ MediaService: UI-ассеты сгенерированы для '%s'", projectName)
	return assets, nil
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

	content, err := s.callLLM(ctx, ModelVeo, prompt, 1024)
	if err != nil {
		log.Printf("⚠️ MediaService Promo Video error: %v — using defaults", err)
		return s.defaultPromoVideo(projectName), nil
	}

	video := s.parsePromoVideo(content, projectName)
	log.Printf("✅ MediaService: промо-видео концепт готов для '%s'", projectName)
	return video, nil
}

// callLLM выполняет запрос к OpenRouter
func (s *MediaService) callLLM(ctx context.Context, model, prompt string, maxTokens int) (string, error) {
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
