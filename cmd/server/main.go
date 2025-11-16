package main

import (
	"context"
	"github.com/zxchelik/avito-test-task/internal/httpserver"
	"os/signal"
	"syscall"
)

func main() {
	server, err := httpserver.NewServer()
	if err != nil {
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := server.Run(ctx); err != nil {
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), server.Cfg.ShutdownTimeout)
	defer cancel()
	server.Shutdown(shutdownCtx)
}
