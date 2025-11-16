package team

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/zxchelik/avito-test-task/internal/httpserver/handlers/shared"
	"github.com/zxchelik/avito-test-task/internal/model"
	modelteam "github.com/zxchelik/avito-test-task/internal/model/team"
	modeluser "github.com/zxchelik/avito-test-task/internal/model/user"
	srvteam "github.com/zxchelik/avito-test-task/internal/service/team"
	"log/slog"
	"net/http"
)

type Handler struct {
	svc *srvteam.Service
	log *slog.Logger
}

func New(svc *srvteam.Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Register регистрирует маршруты команды.
func (h *Handler) Register(r chi.Router) {
	r.Post("/team/add", h.handleTeamAdd)
	r.Get("/team/get", h.handleTeamGet)
}

// POST /team/add
func (h *Handler) handleTeamAdd(w http.ResponseWriter, r *http.Request) {
	var req TeamAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "invalid JSON")
		return
	}
	if req.TeamName == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "team_name is required")
		return
	}

	team := &modelteam.Team{
		Name: req.TeamName,
	}
	members := make([]*modeluser.User, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, &modeluser.User{
			ID:       m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	createdTeam, createdMembers, err := h.svc.Add(r.Context(), team, members)
	if err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeTeamExists, "team_name already exists")
			return
		}
		shared.WriteInternalError(w, err, h.log)
		return
	}

	resp := TeamAddResponse{
		Team: toTeamDTO(createdTeam, createdMembers),
	}
	shared.WriteJSON(w, http.StatusCreated, resp)
}

// GET /team/get?team_name=...
func (h *Handler) handleTeamGet(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "team_name is required")
		return
	}

	team, members, err := h.svc.Get(r.Context(), teamName)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			shared.WriteError(w, http.StatusNotFound, shared.ErrorCodeNotFound, "team not found")
			return
		}
		shared.WriteInternalError(w, err, h.log)
		return
	}

	resp := toTeamDTO(team, members)
	shared.WriteJSON(w, http.StatusOK, resp)
}
