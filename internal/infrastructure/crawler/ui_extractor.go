package crawler

import (
	"strings"
)

// UIExtractor извлекает UI-компоненты и дизайн-систему из HTML
type UIExtractor struct{}

// NewUIExtractor создает новый экстрактор UI
func NewUIExtractor() *UIExtractor {
	return &UIExtractor{}
}

// ExtractUIComponents извлекает UI компоненты из HTML
func (e *UIExtractor) ExtractUIComponents(html string) []string {
	components := make([]string, 0)
	
	// Определяем компоненты по ключевым словам
	if strings.Contains(html, "button") || strings.Contains(html, "btn") {
		components = append(components, "Button")
	}
	if strings.Contains(html, "card") {
		components = append(components, "Card")
	}
	if strings.Contains(html, "modal") || strings.Contains(html, "dialog") {
		components = append(components, "Modal")
	}
	if strings.Contains(html, "navbar") || strings.Contains(html, "header") {
		components = append(components, "Navigation")
	}
	if strings.Contains(html, "form") || strings.Contains(html, "input") {
		components = append(components, "Form")
	}
	if strings.Contains(html, "table") || strings.Contains(html, "grid") {
		components = append(components, "DataGrid")
	}
	if strings.Contains(html, "dropdown") || strings.Contains(html, "select") {
		components = append(components, "Dropdown")
	}
	if strings.Contains(html, "tabs") || strings.Contains(html, "tab-") {
		components = append(components, "Tabs")
	}
	
	return components
}

// ExtractColorPalette извлекает цветовую палитру
func (e *UIExtractor) ExtractColorPalette(html string) []string {
	colors := make([]string, 0)
	
	// Определяем доминирующие цвета по классам и стилям
	if strings.Contains(html, "blue") || strings.Contains(html, "indigo") {
		colors = append(colors, "#4F46E5") // Indigo
	}
	if strings.Contains(html, "purple") || strings.Contains(html, "violet") {
		colors = append(colors, "#7C3AED") // Violet
	}
	if strings.Contains(html, "green") {
		colors = append(colors, "#10B981") // Green
	}
	if strings.Contains(html, "red") {
		colors = append(colors, "#EF4444") // Red
	}
	if strings.Contains(html, "yellow") || strings.Contains(html, "amber") {
		colors = append(colors, "#F59E0B") // Amber
	}
	
	// Добавляем нейтральные цвета
	colors = append(colors, "#18181B") // Zinc-900
	colors = append(colors, "#FFFFFF") // White
	
	return colors
}

// ExtractTypography извлекает типографику
func (e *UIExtractor) ExtractTypography(html string) map[string]string {
	typography := make(map[string]string)
	
	// Определяем шрифты
	if strings.Contains(html, "inter") {
		typography["primary"] = "Inter"
	} else if strings.Contains(html, "roboto") {
		typography["primary"] = "Roboto"
	} else if strings.Contains(html, "geist") {
		typography["primary"] = "Geist Sans"
	} else {
		typography["primary"] = "system-ui"
	}
	
	if strings.Contains(html, "mono") || strings.Contains(html, "code") {
		typography["mono"] = "JetBrains Mono"
	}
	
	return typography
}

// ExtractLayoutPatterns извлекает паттерны компоновки
func (e *UIExtractor) ExtractLayoutPatterns(html string) []string {
	patterns := make([]string, 0)
	
	if strings.Contains(html, "grid") {
		patterns = append(patterns, "Grid Layout")
	}
	if strings.Contains(html, "flex") {
		patterns = append(patterns, "Flexbox")
	}
	if strings.Contains(html, "sidebar") {
		patterns = append(patterns, "Sidebar Navigation")
	}
	if strings.Contains(html, "hero") {
		patterns = append(patterns, "Hero Section")
	}
	if strings.Contains(html, "footer") {
		patterns = append(patterns, "Footer")
	}
	
	return patterns
}

// ExtractDesignSystem извлекает элементы дизайн-системы
func (e *UIExtractor) ExtractDesignSystem(html string) map[string]interface{} {
	designSystem := make(map[string]interface{})
	
	designSystem["components"] = e.ExtractUIComponents(html)
	designSystem["colors"] = e.ExtractColorPalette(html)
	designSystem["typography"] = e.ExtractTypography(html)
	designSystem["layouts"] = e.ExtractLayoutPatterns(html)
	
	// Определяем стиль дизайна
	styles := make([]string, 0)
	if strings.Contains(html, "glass") || strings.Contains(html, "blur") {
		styles = append(styles, "Glassmorphism")
	}
	if strings.Contains(html, "gradient") {
		styles = append(styles, "Gradients")
	}
	if strings.Contains(html, "shadow") {
		styles = append(styles, "Shadows")
	}
	if strings.Contains(html, "rounded") {
		styles = append(styles, "Rounded Corners")
	}
	designSystem["styles"] = styles
	
	return designSystem
}
