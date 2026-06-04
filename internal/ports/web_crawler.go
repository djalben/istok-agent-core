package ports

import (
	"context"

	"github.com/djalben/istok-agent-core/internal/domain"
)

// CrawlRequest - запрос на парсинг сайта.
type CrawlRequest struct {
	URL   string
	Depth int
}

// CrawlResponse - результат парсинга.
type CrawlResponse struct {
	URL          string
	Title        string
	Technologies []string
	Patterns     []*domain.Pattern
	Insights     []*domain.Insight
	Structure    map[string]any
	Confidence   float64
}

// WebCrawler - интерфейс для парсинга сайтов.
type WebCrawler interface {
	CrawlWebsite(ctx context.Context, req CrawlRequest) (*CrawlResponse, error)
	ExtractTechnologies(html string) ([]string, error)
	ExtractPatterns(html string) ([]*domain.Pattern, error)
	GenerateInsights(data map[string]any) ([]*domain.Insight, error)
}
