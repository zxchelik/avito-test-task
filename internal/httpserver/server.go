package httpserver

import (
	"context"
	"errors"
	"github.com/zxchelik/avito-test-task/internal/application"
	"github.com/zxchelik/avito-test-task/internal/httpserver/handlers"
	"github.com/zxchelik/avito-test-task/internal/infrastructure/pg"
	prRep "github.com/zxchelik/avito-test-task/internal/repository/pull_request"
	raRep "github.com/zxchelik/avito-test-task/internal/repository/reviewer_assignment"
	teamRep "github.com/zxchelik/avito-test-task/internal/repository/team"
	userRep "github.com/zxchelik/avito-test-task/internal/repository/user"
	prSvc "github.com/zxchelik/avito-test-task/internal/service/pull_request"
	teamSvc "github.com/zxchelik/avito-test-task/internal/service/team"
	userSvc "github.com/zxchelik/avito-test-task/internal/service/user"

	"github.com/zxchelik/avito-test-task/pkg/logger"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	Http *http.Server
	Log  *slog.Logger
	Cfg  *application.Config
}

func NewServer() (*Server, error) {
	cfg := application.MustLoad()

	log := logger.New(cfg.Env)

	db, err := application.NewDB(context.Background(), &cfg.Postgres, log)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	txManager := pg.NewTxManager(db.Pool)

	// Репозитории
	userRepo := userRep.NewPGRepository(db.Pool)
	teamRepo := teamRep.NewPGRepository(db.Pool)
	prRepo := prRep.NewPGRepository(db.Pool)
	raRepo := raRep.NewPGRepository(db.Pool)

	// Сервисы
	userService := userSvc.NewService(userRepo, prRepo, raRepo)
	teamService := teamSvc.NewService(teamRepo, userRepo, txManager)
	prService := prSvc.NewService(prRepo, userRepo, raRepo, txManager)

	handler := handlers.NewHandler(teamService, userService, prService, log)

	return &Server{
		Http: &http.Server{
			Addr:         cfg.Server.Address(),
			Handler:      handler.Router(),
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		Log: log,
		Cfg: cfg,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.Log.Info("starting server", slog.String("env", string(s.Cfg.Env)), slog.String("address", s.Cfg.Address()))
		if err := s.Http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		s.Log.Info("shutdown signal received", slog.String("signal", "SIGINT/SIGTERM"))
		return nil
	case err := <-errCh:
		if err != nil {
			s.Log.Error("server error", slog.String("error", err.Error()))
			return err
		}
		return nil
	}
}

func (s *Server) Shutdown(ctx context.Context) {
	s.Log.Info("shutting down server", slog.Int64("graceful_timeout_seconds", int64(s.Cfg.ShutdownTimeout/time.Second)))
	if err := s.Http.Shutdown(ctx); err != nil {
		s.Log.Error("graceful shutdown failed", slog.String("error", err.Error()))
	} else {
		s.Log.Info("server stopped gracefully")
	}
}
