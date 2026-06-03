package domain

import "time"

// User — сущность пользователя (Layer 1: Auth & DB).
// PasswordHash никогда не сериализуется в JSON.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// ProfileStats — агрегированная статистика для страницы профиля.
// Соответствует контракту фронтенда (ProfileStatsSchema).
type ProfileStats struct {
	TotalProjects     int `json:"total_projects"`
	PublishedProjects int `json:"published_projects"`
	TotalGenerations  int `json:"total_generations"`
	DaysActive        int `json:"days_active"`
	CurrentStreak     int `json:"current_streak"`
}
