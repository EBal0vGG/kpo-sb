package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hse/pkg-config"
	"github.com/hse/pkg-logger"

	httpapi "github.com/hse/file-analysis/internal/httpapi"
)

func main() {
	cfg, err := config.Load("file-analysis")
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.ServiceName)
	handler, worker := httpapi.New(log, cfg)

	// Start worker in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Start(ctx)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	go func() {
		log.Info("analysis server starting", slog.String("addr", cfg.HTTP.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("analysis server failed", slog.Any("err", err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	// Stop worker
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownGrace)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("analysis server shutdown error", slog.Any("err", err))
	}
	log.Info("analysis server stopped")
}
