package user

import (
	"encoding/json"
	"errors"
	"github.com/zxchelik/avito-test-task/internal/httpserver/handlers/shared"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zxchelik/avito-test-task/internal/model"
	srvuser "github.com/zxchelik/avito-test-task/internal/service/user"
)

type Handler struct {
	svc *srvuser.Service
	log *slog.Logger
}

func New(svc *srvuser.Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/users/setIsActive", h.handleUsersSetIsActive)
	r.Get("/users/getReview", h.handleUsersGetReview)
}

// POST /users/setIsActive
func (h *Handler) handleUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "invalid JSON")
		return
	}
	if req.UserID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "user_id is required")
		return
	}

	user, err := h.svc.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			shared.WriteError(w, http.StatusNotFound, shared.ErrorCodeNotFound, "user not found")
			return
		}
		shared.WriteInternalError(w, err, h.log)
		return
	}

	resp := SetIsActiveResponse{
		User: toUserDTO(user),
	}
	shared.WriteJSON(w, http.StatusOK, resp)
}

// GET /users/getReview?user_id=...
func (h *Handler) handleUsersGetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "user_id is required")
		return
	}

	prs, err := h.svc.ListUserReviews(r.Context(), userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// Если сервис вернёт ErrNotFound для пользователя.
			shared.WriteError(w, http.StatusNotFound, shared.ErrorCodeNotFound, "user not found")
			return
		}
		shared.WriteInternalError(w, err, h.log)
		return
	}

	out := make([]PullRequestShortDTO, 0, len(prs))
	for _, pr := range prs {
		out = append(out, toPullRequestShortDTO(pr))
	}

	resp := UserReviewsResponse{
		UserID:       userID,
		PullRequests: out,
	}
	shared.WriteJSON(w, http.StatusOK, resp)
}
