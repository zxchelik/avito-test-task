package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/metrics"
	"log/slog"
	"net/http"

	srvpr "github.com/zxchelik/avito-test-task/internal/service/pull_request"
	srvteam "github.com/zxchelik/avito-test-task/internal/service/team"
	srvuser "github.com/zxchelik/avito-test-task/internal/service/user"

	prhandlers "github.com/zxchelik/avito-test-task/internal/httpserver/handlers/pull_request"
	teamhandlers "github.com/zxchelik/avito-test-task/internal/httpserver/handlers/team"
	userhandlers "github.com/zxchelik/avito-test-task/internal/httpserver/handlers/user"
)

type Handler struct {
	teamSvc *srvteam.Service
	userSvc *srvuser.Service
	prSvc   *srvpr.Service
	log     *slog.Logger
}

func NewHandler(
	teamSvc *srvteam.Service,
	userSvc *srvuser.Service,
	prSvc *srvpr.Service,
	log *slog.Logger,
) *Handler {
	return &Handler{
		teamSvc: teamSvc,
		userSvc: userSvc,
		prSvc:   prSvc,
		log:     log,
	}
}

// Router возвращает chi.Router, соответствующий openapi.yml.
func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(metrics.Collector(metrics.CollectorOpts{
		Host:  false,
		Proto: true,
		Skip: func(r *http.Request) bool {
			if r.Method == http.MethodOptions {
				return true // OPTIONS не считаем
			}
			return false
		},
	}))

	r.Handle("/metrics", metrics.Handler())

	// Team endpoints
	teamHandler := teamhandlers.New(h.teamSvc, h.log)
	teamHandler.Register(r)

	// User endpoints
	userHandler := userhandlers.New(h.userSvc, h.log)
	userHandler.Register(r)

	// PullRequest endpoints
	prHandler := prhandlers.New(h.prSvc, h.log)
	prHandler.Register(r)

	return r
}
