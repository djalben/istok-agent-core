package ports

import "context"

// MediaType определяет тип генерируемого медиа-контента
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

// ImageGenerationRequest запрос на генерацию изображения
type ImageGenerationRequest struct {
	Prompt      string
	Model       string // "nano-banana-2", "gemini-flash-image", "dall-e-3"
	Width       int
	Height      int
	Quality     string // "standard", "hd"
	Style       string // "natural", "vivid"
	NumImages   int
}

// ImageGenerationResponse ответ с сгенерированным изображением
type ImageGenerationResponse struct {
	ImageURL    string
	ImageBase64 string
	Model       string
	Prompt      string
	RevisedPrompt string
	Width       int
	Height      int
	Cost        float64
}

// VideoGenerationRequest запрос на генерацию видео
type VideoGenerationRequest struct {
	Prompt      string
	Model       string // "veo", "runway-gen3"
	Duration    int    // секунды
	Resolution  string // "720p", "1080p", "4k"
	FPS         int
	Style       string
}

// VideoGenerationResponse ответ с сгенерированным видео
type VideoGenerationResponse struct {
	VideoURL    string
	ThumbnailURL string
	Model       string
	Prompt      string
	Duration    int
	Resolution  string
	Cost        float64
	Status      string // "processing", "completed", "failed"
}

// MediaGenerator интерфейс для генерации медиа-контента
type MediaGenerator interface {
	// GenerateImage генерирует изображение по текстовому описанию
	GenerateImage(ctx context.Context, req ImageGenerationRequest) (*ImageGenerationResponse, error)
	
	// GenerateVideo генерирует видео по текстовому описанию
	GenerateVideo(ctx context.Context, req VideoGenerationRequest) (*VideoGenerationResponse, error)
	
	// GetVideoStatus получает статус генерации видео (для асинхронных операций)
	GetVideoStatus(ctx context.Context, jobID string) (*VideoGenerationResponse, error)
	
	// ListAvailableModels возвращает список доступных моделей для генерации
	ListAvailableModels(ctx context.Context, mediaType MediaType) ([]string, error)
}
