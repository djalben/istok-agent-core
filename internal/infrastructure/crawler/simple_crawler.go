package crawler

import (
	"context"
	"strings"
	"time"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
)

type SimpleCrawler struct {
	maxDepth int
	timeout  time.Duration
}

// NewSimpleCrawler создает новый crawler.
func NewSimpleCrawler() *SimpleCrawler {
	return &SimpleCrawler{
		maxDepth: 3,
		timeout:  30 * time.Second,
	}
}

// CrawlWebsite парсит сайт и возвращает данные
// ЗАГЛУШКА: В реальной версии здесь будет интеграция с Colly или Playwright.
func (c *SimpleCrawler) CrawlWebsite(ctx context.Context, req ports.CrawlRequest) (*ports.CrawlResponse, error) {
	ports.LoggerFromContext(ctx).InfoContext(ctx, "crawl website started",
		"url", req.URL,
		"depth", req.Depth,
	)

	// Имитация анализа сайта
	technologies := c.detectTechnologies(req.URL)
	patterns := c.detectPatterns(req.URL)
	insights := c.generateInsights(req.URL, technologies)

	structure := map[string]any{
		"url":       req.URL,
		"analyzed":  true,
		"pages":     []string{req.URL},
		"depth":     req.Depth,
		"mock_data": true,
	}

	return &ports.CrawlResponse{
		URL:          req.URL,
		Title:        c.extractTitle(req.URL),
		Technologies: technologies,
		Patterns:     patterns,
		Insights:     insights,
		Structure:    structure,
		Confidence:   0.85, // высокая уверенность для демо
	}, nil
}

// ExtractTechnologies извлекает технологии из HTML.
func (c *SimpleCrawler) ExtractTechnologies(html string) ([]string, error) {
	technologies := make([]string, 0)

	// Простая эвристика на основе ключевых слов
	keywords := map[string]string{
		"react":      "React",
		"vue":        "Vue.js",
		"angular":    "Angular",
		"next":       "Next.js",
		"nuxt":       "Nuxt.js",
		"tailwind":   "Tailwind CSS",
		"bootstrap":  "Bootstrap",
		"jquery":     "jQuery",
		"typescript": "TypeScript",
		"webpack":    "Webpack",
	}

	htmlLower := strings.ToLower(html)
	for keyword, tech := range keywords {
		if strings.Contains(htmlLower, keyword) {
			technologies = append(technologies, tech)
		}
	}

	return technologies, nil
}

// ExtractPatterns извлекает UI/UX паттерны.
func (c *SimpleCrawler) ExtractPatterns(html string) ([]*domain.Pattern, error) {
	patterns := make([]*domain.Pattern, 0)

	// Заглушка: определяем паттерны на основе URL
	if strings.Contains(html, "component") || strings.Contains(html, "react") {
		pattern := domain.NewPattern(
			domain.PatternTypeUI,
			"Компонентная архитектура",
			"Сайт использует компонентный подход к разработке UI",
		)
		pattern.Confidence = 0.9
		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// GenerateInsights генерирует инсайты на основе данных.
func (c *SimpleCrawler) GenerateInsights(_ map[string]any) ([]*domain.Insight, error) {
	insights := make([]*domain.Insight, 0, 1)

	// Заглушка: создаем базовые инсайты
	insight := domain.NewInsight(
		"Современный технологический стек",
		"Сайт использует актуальные технологии и фреймворки",
		"technology_trend",
		0.85,
	)
	insight.Priority = 7
	insights = append(insights, insight)

	return insights, nil
}

// detectTechnologies определяет технологии по URL (заглушка).
func (c *SimpleCrawler) detectTechnologies(url string) []string {
	technologies := make([]string, 0)

	urlLower := strings.ToLower(url)

	// Эвристика на основе популярных доменов
	switch {
	case strings.Contains(urlLower, "react") || strings.Contains(urlLower, "vercel"):
		technologies = append(technologies, "React", "Next.js", "Vercel")
	case strings.Contains(urlLower, "vue"):
		technologies = append(technologies, "Vue.js", "Nuxt.js")
	case strings.Contains(urlLower, "github"):
		technologies = append(technologies, "Git", "GitHub Pages")
	default:
		technologies = append(technologies, "HTML5", "CSS3", "JavaScript")
	}

	// Всегда добавляем популярные технологии для демо
	technologies = append(technologies, "Tailwind CSS", "TypeScript")

	return technologies
}

// detectPatterns определяет паттерны по URL (заглушка).
func (c *SimpleCrawler) detectPatterns(_ string) []*domain.Pattern {
	patterns := make([]*domain.Pattern, 0, 2)

	// UI паттерн
	uiPattern := domain.NewPattern(
		domain.PatternTypeUI,
		"Современный минималистичный дизайн",
		"Сайт использует чистый, минималистичный дизайн с акцентом на контент",
	)
	uiPattern.Confidence = 0.8
	uiPattern.Frequency = 1
	patterns = append(patterns, uiPattern)

	// Архитектурный паттерн
	archPattern := domain.NewPattern(
		domain.PatternTypeArchitecture,
		"SPA (Single Page Application)",
		"Архитектура одностраничного приложения для быстрой навигации",
	)
	archPattern.Confidence = 0.75
	archPattern.Frequency = 1
	patterns = append(patterns, archPattern)

	return patterns
}

// generateInsights генерирует инсайты по URL (заглушка).
func (c *SimpleCrawler) generateInsights(_ string, technologies []string) []*domain.Insight {
	insights := make([]*domain.Insight, 0)

	// Генерируем инсайты на основе технологий
	if len(technologies) > 0 {
		insight := domain.NewInsight(
			"Технологический стек",
			"Сайт использует: "+strings.Join(technologies, ", "),
			"technical",
			0.8,
		)
		insights = append(insights, insight)
	}

	// UI/UX инсайты
	uiInsight := domain.NewInsight(
		"UI Паттерны",
		"Обнаружены современные UI паттерны: карточный дизайн, градиенты, адаптивная сетка",
		"ux",
		0.75,
	)
	uiInsight.Priority = 7
	uiInsight.Actionable = true
	insights = append(insights, uiInsight)

	// Инсайт по возможностям улучшения
	opportunityInsight := domain.NewInsight(
		"Возможности оптимизации",
		"Можно улучшить производительность и добавить современные анимации",
		"opportunity",
		0.7,
	)
	opportunityInsight.Priority = 6
	opportunityInsight.Actionable = true
	insights = append(insights, opportunityInsight)

	return insights
}

// extractTitle извлекает заголовок из URL (заглушка).
func (c *SimpleCrawler) extractTitle(url string) string {
	// Простая эвристика для демо
	parts := strings.Split(url, "/")
	if len(parts) > 2 {
		domain := parts[2]
		domain = strings.TrimPrefix(domain, "www.")

		return "Анализ сайта: " + domain
	}

	return "Анализ сайта"
}
