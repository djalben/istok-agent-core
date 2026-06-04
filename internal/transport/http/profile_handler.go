package http

import (
	"net/http"
	"time"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"github.com/djalben/istok-agent-core/internal/domain"
)

// ProfileHandler — GET /api/v1/user/profile (за AuthMiddleware).
type ProfileHandler struct {
	auth     *usecases.AuthService
	projects *usecases.ProjectService
}

func NewProfileHandler(auth *usecases.AuthService, projects *usecases.ProjectService) *ProfileHandler {
	return &ProfileHandler{auth: auth, projects: projects}
}

// profileResponse соответствует UserProfileSchema на фронтенде.
type profileResponse struct {
	ID          string              `json:"id"`
	Email       string              `json:"email"`
	DisplayName string              `json:"display_name"`
	Username    *string             `json:"username"`
	AvatarURL   *string             `json:"avatar_url"`
	Bio         *string             `json:"bio"`
	Location    *string             `json:"location"`
	Website     *string             `json:"website"`
	CreatedAt   time.Time           `json:"created_at"`
	Stats       domain.ProfileStats `json:"stats"`
	Activity    []int               `json:"activity"`
}

func (h *ProfileHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")

		return
	}
	user, err := h.auth.GetByID(r.Context(), userID)
	if err != nil {
		writeDomainError(w, err)

		return
	}
	stats, err := h.projects.Stats(r.Context(), userID)
	if err != nil {
		writeDomainError(w, err)

		return
	}
	_ = writeJSON(w, http.StatusOK, profileResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		Stats:       stats,
		Activity:    []int{}, // нет таблицы активности — пустой граф
	})
}
