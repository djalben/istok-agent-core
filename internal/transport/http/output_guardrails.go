package http

import (
	"regexp"
	"strings"
	"sync"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Output Guardrails (SSE Layer)
//  Сканирует ответы агентов на лету, блокирует утечки
//  системных промптов, jailbreak-маркеров, секретов.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// PromptMaskFilter — фильтр-санитайзер для исходящих SSE-событий.
// Если detectLeak() возвращает true, сообщение подменяется на вежливый отказ.
type PromptMaskFilter struct {
	patterns       []*regexp.Regexp
	literalPhrases []string
	refusal        string
	mu             sync.RWMutex
}

// NewPromptMaskFilter создаёт фильтр с дефолтным набором детекторов утечки.
func NewPromptMaskFilter() *PromptMaskFilter {
	return &PromptMaskFilter{
		patterns: []*regexp.Regexp{
			// Классические markers system prompt
			regexp.MustCompile(`(?i)you\s+are\s+(?:an?\s+)?(?:expert|senior|professional|helpful)\s+(?:software|web|frontend|backend|full[- ]?stack|UI|UX|product|system)`),
			regexp.MustCompile(`(?i)my\s+(?:system\s+)?(?:instructions?|prompt|directives?|guidelines?)`),
			regexp.MustCompile(`(?i)(?:as\s+(?:per|stated\s+in)\s+)?(?:the\s+)?system\s+prompt`),
			regexp.MustCompile(`(?i)i\s+(?:was|am)\s+(?:trained|instructed|configured|programmed)\s+(?:to|by|with)`),
			regexp.MustCompile(`(?i)(?:my|the)\s+(?:internal|secret|hidden)\s+(?:rules?|instructions?|prompt)`),
			regexp.MustCompile(`(?i)ignore\s+(?:all\s+)?(?:previous|prior|above)\s+(?:instructions?|prompts?)`),
			regexp.MustCompile(`(?i)(?:reveal|disclose|show|print|leak|dump)\s+(?:your|the|my)\s+(?:system\s+)?(?:prompt|instructions?)`),
			// Jailbreak / role-override patterns
			regexp.MustCompile(`(?i)\bDAN\b\s+(?:mode|prompt)`),
			regexp.MustCompile(`(?i)pretend\s+(?:you\s+(?:are|have))?\s+(?:no\s+)?(?:rules?|restrictions?|guidelines?)`),
			regexp.MustCompile(`(?i)(?:bypass|override|disable)\s+(?:your\s+)?(?:safety|security|content)\s+(?:filter|policy|rules?)`),
			// Role markers внутри ответа
			regexp.MustCompile(`(?i)<\|im_(?:start|end)\|>`),
			regexp.MustCompile(`(?im)^(?:###\s*)?(?:system|assistant|user)\s*:\s*$`),
		},
		literalPhrases: []string{
			"You are an expert",
			"You are a senior",
			"My system instructions",
			"my system prompt",
			"the system prompt",
			"ARCHITECTURE RULES:",
			"You are a senior software architect",
			"You are a senior product strategist",
			"You are a frontend code fixer",
			"Output only valid JSON",
			"You are an Istok",
			"Do not reveal your",
		},
		refusal: "Извините, я не могу раскрыть внутренние инструкции или системные настройки. Если у вас есть конкретный вопрос по проекту — задайте его, пожалуйста.",
	}
}

// SetRefusal позволяет кастомизировать текст отказа.
func (f *PromptMaskFilter) SetRefusal(msg string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refusal = msg
}

// AddLiteralPhrase добавляет фразу для буквального сравнения (case-insensitive substring).
func (f *PromptMaskFilter) AddLiteralPhrase(phrase string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.literalPhrases = append(f.literalPhrases, phrase)
}

// AddPattern добавляет regex-паттерн для детекции.
func (f *PromptMaskFilter) AddPattern(pat *regexp.Regexp) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.patterns = append(f.patterns, pat)
}

// Detect возвращает (ok, reason).
// ok=true → утечка обнаружена, reason описывает совпавший паттерн.
func (f *PromptMaskFilter) Detect(text string) (bool, string) {
	if text == "" {
		return false, ""
	}
	f.mu.RLock()
	defer f.mu.RUnlock()

	lower := strings.ToLower(text)
	for _, phrase := range f.literalPhrases {
		if strings.Contains(lower, strings.ToLower(phrase)) {
			return true, "literal:" + phrase
		}
	}
	for _, pat := range f.patterns {
		if pat.MatchString(text) {
			return true, "regex:" + pat.String()
		}
	}
	return false, ""
}

// Sanitize возвращает безопасную версию текста.
// Если утечки нет — возвращает оригинал.
// Если утечка найдена — возвращает refusal-сообщение.
// Также возвращает (sanitized, leaked, reason) для логирования.
func (f *PromptMaskFilter) Sanitize(text string) (string, bool, string) {
	if leaked, reason := f.Detect(text); leaked {
		f.mu.RLock()
		refusal := f.refusal
		f.mu.RUnlock()
		return refusal, true, reason
	}
	return text, false, ""
}

// SanitizeMap фильтрует map[string]interface{} — рекурсивно ищет string-поля
// и подменяет их на refusal, если найдена утечка.
// Возвращает (sanitized_map, leaked_count, first_reason).
func (f *PromptMaskFilter) SanitizeMap(data map[string]interface{}) (map[string]interface{}, int, string) {
	if data == nil {
		return data, 0, ""
	}
	leaked := 0
	firstReason := ""
	out := make(map[string]interface{}, len(data))
	for k, v := range data {
		switch val := v.(type) {
		case string:
			sanitized, isLeaked, reason := f.Sanitize(val)
			out[k] = sanitized
			if isLeaked {
				leaked++
				if firstReason == "" {
					firstReason = reason + " (field=" + k + ")"
				}
			}
		case map[string]interface{}:
			subOut, subLeaked, subReason := f.SanitizeMap(val)
			out[k] = subOut
			leaked += subLeaked
			if firstReason == "" && subReason != "" {
				firstReason = subReason
			}
		case []interface{}:
			out[k] = f.sanitizeSlice(val, &leaked, &firstReason)
		default:
			out[k] = v
		}
	}
	return out, leaked, firstReason
}

func (f *PromptMaskFilter) sanitizeSlice(slice []interface{}, leaked *int, firstReason *string) []interface{} {
	out := make([]interface{}, len(slice))
	for i, item := range slice {
		switch val := item.(type) {
		case string:
			sanitized, isLeaked, reason := f.Sanitize(val)
			out[i] = sanitized
			if isLeaked {
				*leaked++
				if *firstReason == "" {
					*firstReason = reason
				}
			}
		case map[string]interface{}:
			subOut, subLeaked, subReason := f.SanitizeMap(val)
			out[i] = subOut
			*leaked += subLeaked
			if *firstReason == "" && subReason != "" {
				*firstReason = subReason
			}
		default:
			out[i] = item
		}
	}
	return out
}
