package pull_request

import (
	"encoding/json"
	"errors"
	"github.com/zxchelik/avito-test-task/internal/httpserver/handlers/shared"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zxchelik/avito-test-task/internal/model"
	modelpr "github.com/zxchelik/avito-test-task/internal/model/pull_request"
	modelra "github.com/zxchelik/avito-test-task/internal/model/reviewer_assignment"
	modeluser "github.com/zxchelik/avito-test-task/internal/model/user"
	srvpr "github.com/zxchelik/avito-test-task/internal/service/pull_request"
)

type Handler struct {
	svc *srvpr.Service
	log *slog.Logger
}

func New(
	svc *srvpr.Service,
	log *slog.Logger,
) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/pullRequest/create", h.handlePullRequestCreate)
	r.Post("/pullRequest/merge", h.handlePullRequestMerge)
	r.Post("/pullRequest/reassign", h.handlePullRequestReassign)
}

// POST /pullRequest/create
func (h *Handler) handlePullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var req PullRequestCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "invalid JSON")
		return
	}
	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "pull_request_id, pull_request_name and author_id are required")
		return
	}

	prModel := &modelpr.PullRequest{
		ID:       req.PullRequestID,
		Title:    req.PullRequestName,
		AuthorID: req.AuthorID,
		Status:   modelpr.PROpen,
	}

	created, reviewers, err := h.svc.Create(r.Context(), prModel)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrNotFound):
			shared.WriteError(w, http.StatusNotFound, shared.ErrorCodeNotFound, "author or team not found")
			return
		case errors.Is(err, model.ErrAlreadyExists):
			shared.WriteError(w, http.StatusConflict, shared.ErrorCodePRExists, "PR id already exists")
			return
		case errors.Is(err, modeluser.ErrUserInactive):
			shared.WriteError(w, http.StatusConflict, shared.ErrorCodeNoCandidate, "author is inactive")
			return
		case errors.Is(err, modelra.ErrNoReviewerCandidatesLeft):
			shared.WriteError(w, http.StatusConflict, shared.ErrorCodeNoCandidate, "no active reviewer candidates in team")
			return
		default:
			shared.WriteInternalError(w, err, h.log)
			return
		}
	}

	resp := PullRequestCreateResponse{
		PR: toPullRequestDTO(created, reviewers),
	}
	shared.WriteJSON(w, http.StatusCreated, resp)
}

// POST /pullRequest/merge
func (h *Handler) handlePullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var req PullRequestMergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "invalid JSON")
		return
	}
	if req.PullRequestID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "pull_request_id is required")
		return
	}

	pr, err := h.svc.Merge(r.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			shared.WriteError(w, http.StatusNotFound, shared.ErrorCodeNotFound, "PR not found")
			return
		}
		shared.WriteInternalError(w, err, h.log)
		return
	}

	resp := PullRequestMergeResponse{
		PR: toPullRequestDTO(pr, nil),
	}
	shared.WriteJSON(w, http.StatusOK, resp)
}

// POST /pullRequest/reassign
func (h *Handler) handlePullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var req PullRequestReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "invalid JSON")
		return
	}
	if req.PullRequestID == "" || req.OldUserID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrorCodeInternal, "pull_request_id and old_user_id are required")
		return
	}

	newReviewer, err := h.svc.Reassign(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrNotFound):
			shared.WriteError(w, http.StatusNotFound, shared.ErrorCodeNotFound, "PR or user not found")
			return
		case errors.Is(err, modelra.ErrReviewerNotFoundInPR):
			shared.WriteError(w, http.StatusConflict, shared.ErrorCodeNotAssigned, "reviewer is not assigned to this PR")
			return
		case errors.Is(err, modelra.ErrNoReviewerCandidatesLeft):
			shared.WriteError(w, http.StatusConflict, shared.ErrorCodeNoCandidate, "no active replacement candidate in team")
			return
		case errors.Is(err, modelpr.ErrPRAlreadyMerged):
			shared.WriteError(w, http.StatusConflict, shared.ErrorCodePRMerged, "cannot reassign on merged PR")
			return
		default:
			shared.WriteError(w, http.StatusConflict, shared.ErrorCodePRMerged, "cannot reassign on merged PR")
			return
		}
	}

	pr, err := h.svc.GetByID(r.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			shared.WriteError(w, http.StatusNotFound, shared.ErrorCodeNotFound, "PR not found")
			return
		}
		shared.WriteInternalError(w, err, h.log)
		return
	}

	resp := PullRequestReassignResponse{
		PR:         toPullRequestDTO(pr, nil),
		ReplacedBy: newReviewer.ID,
	}
	shared.WriteJSON(w, http.StatusOK, resp)
}
