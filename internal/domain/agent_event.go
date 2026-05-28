package domain

import "time"

// AgentRole определяет роль агента в системе (domain-уровень).
type AgentRole string

const (
	RoleResearcher   AgentRole = "researcher"
	RoleBrain        AgentRole = "brain"
	RoleDirector     AgentRole = "director"
	RoleArchitect    AgentRole = "architect"
	RolePlanner      AgentRole = "planner"
	RoleCoder        AgentRole = "coder"
	RoleDesigner     AgentRole = "designer"
	RoleVideographer AgentRole = "videographer"
	RoleValidator    AgentRole = "validator"
	RoleSecurity     AgentRole = "security"
	RoleTester       AgentRole = "tester"
	RoleUIReviewer   AgentRole = "ui_reviewer"
)

// EventKind тип события в пайплайне
type EventKind string

const (
	EventStatus        EventKind = "status"         // обновление статуса агента
	EventFSM           EventKind = "fsm"            // переход FSM
	EventFile          EventKind = "file"           // сгенерированный файл
	EventPlan          EventKind = "plan"           // утверждённый план
	EventError         EventKind = "error"          // ошибка агента
	EventDone          EventKind = "done"           // завершение пайплайна
	EventReflection    EventKind = "reflection"     // скрытый лог рассуждений (Thought Chain)
	EventUserAction    EventKind = "user_action"    // пауза FSM — ожидание решения пользователя
	EventReplan        EventKind = "replan"         // сигнал фронтенду: перезапустить стрим с фидбеком
	EventMediaApproval EventKind = "media_approval" // пауза FSM — ожидание утверждения медиа-промптов
)

// MediaAsset — описание одного медиа-ассета для дизайн-ревью.
type MediaAsset struct {
	ID            string `json:"id"`
	Type          string `json:"type"`      // "image" | "video"
	Placement     string `json:"placement"` // "hero" | "og" | "logo" | "promo_video"
	Label         string `json:"label"`     // человекочитаемое название: "Главный фон (Hero)"
	Prompt        string `json:"prompt"`
	PreviewURL    string `json:"preview_url,omitempty"`    // stock photo URL (default) or AI-generated preview
	StockKeywords string `json:"stock_keywords,omitempty"` // comma-separated keywords for Unsplash fallback
	URL           string `json:"url,omitempty"`            // final generated URL (post-approval)
}

// AgentState — статус выполнения агента в пайплайне.
type AgentState string

const (
	AgentStateIdle       AgentState = "idle"
	AgentStateRunning    AgentState = "running"
	AgentStateReflecting AgentState = "reflecting" // Thought Chain: [Goal]->[Hypothesis]->[Verification]->[Action]
	AgentStateCompleted  AgentState = "completed"
	AgentStateError      AgentState = "error"
)

// AgentEvent — событие, публикуемое агентом в шину событий.
// Транспортный слой (SSE) подписывается на канал и транслирует
// эти события клиенту в реальном времени.
type AgentEvent struct {
	Kind      EventKind `json:"kind"`
	Agent     AgentRole `json:"agent"`
	State     TaskState `json:"state,omitempty"`
	Message   string    `json:"message"`
	Progress  int       `json:"progress"`
	Timestamp time.Time `json:"timestamp"`

	// Payload — опциональные данные (файл, план, мета и т.д.)
	Filename    string       `json:"filename,omitempty"`
	Content     string       `json:"content,omitempty"`
	AgentPhase  AgentState   `json:"agent_phase,omitempty"`  // reflecting / running / completed
	DraftPlan   string       `json:"draft_plan,omitempty"`   // предложенная архитектура (для user_action)
	SessionID   string       `json:"session_id,omitempty"`   // идентификатор сессии (для approve endpoint)
	MediaAssets []MediaAsset `json:"media_assets,omitempty"` // ассеты для дизайн-ревью (media_approval)
}

// EventBus — канал для обмена событиями между агентами и транспортным слоем.
// Буферизированный канал предотвращает блокировку агентов при медленном SSE.
type EventBus struct {
	ch chan AgentEvent
}

// NewEventBus создаёт шину событий с указанным размером буфера.
func NewEventBus(bufferSize int) *EventBus {
	if bufferSize < 1 {
		bufferSize = 128
	}
	return &EventBus{
		ch: make(chan AgentEvent, bufferSize),
	}
}

// Publish отправляет событие в шину. Неблокирующий: если буфер заполнен, событие отбрасывается.
func (bus *EventBus) Publish(event AgentEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	select {
	case bus.ch <- event:
	default:
		// буфер заполнен — отбрасываем (лучше потерять событие, чем заблокировать агента)
	}
}

// Subscribe возвращает read-only канал для получения событий (для SSE handler).
func (bus *EventBus) Subscribe() <-chan AgentEvent {
	return bus.ch
}

// Close закрывает шину. Все подписчики получат закрытие канала.
func (bus *EventBus) Close() {
	close(bus.ch)
}

// PublishStatus — удобный хелпер для публикации статусного события.
func (bus *EventBus) PublishStatus(agent AgentRole, state TaskState, message string, progress int) {
	bus.Publish(AgentEvent{
		Kind:      EventStatus,
		Agent:     agent,
		State:     state,
		Message:   message,
		Progress:  progress,
		Timestamp: time.Now(),
	})
}

// PublishFSMTransition — публикует событие перехода FSM.
func (bus *EventBus) PublishFSMTransition(from, to TaskState, reason string) {
	bus.Publish(AgentEvent{
		Kind:      EventFSM,
		State:     to,
		Message:   reason,
		Timestamp: time.Now(),
	})
}

// PublishFile — публикует сгенерированный файл.
func (bus *EventBus) PublishFile(agent AgentRole, filename, content string) {
	bus.Publish(AgentEvent{
		Kind:      EventFile,
		Agent:     agent,
		Filename:  filename,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// PublishError — публикует ошибку агента.
func (bus *EventBus) PublishError(agent AgentRole, err error) {
	bus.Publish(AgentEvent{
		Kind:      EventError,
		Agent:     agent,
		Message:   err.Error(),
		Timestamp: time.Now(),
	})
}

// PublishDone — публикует завершение пайплайна.
func (bus *EventBus) PublishDone(message string) {
	bus.Publish(AgentEvent{
		Kind:      EventDone,
		Message:   message,
		Timestamp: time.Now(),
	})
}

// PublishMediaApproval — публикует запрос на утверждение медиа-ассетов (дизайн-ревью).
func (bus *EventBus) PublishMediaApproval(agent AgentRole, assets []MediaAsset, sessionID string) {
	bus.Publish(AgentEvent{
		Kind:        EventMediaApproval,
		Agent:       agent,
		Message:     "⏸️ Ожидание утверждения медиа-ассетов...",
		MediaAssets: assets,
		SessionID:   sessionID,
		Timestamp:   time.Now(),
	})
}

// PublishReflection — публикует скрытый лог рассуждений агента (Thought Chain).
func (bus *EventBus) PublishReflection(agent AgentRole, thoughtChain string) {
	bus.Publish(AgentEvent{
		Kind:       EventReflection,
		Agent:      agent,
		Message:    thoughtChain,
		AgentPhase: AgentStateReflecting,
		Timestamp:  time.Now(),
	})
}
