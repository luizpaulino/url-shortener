package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luizpaulino/url-shortener/internal/httpserver"
	"github.com/luizpaulino/url-shortener/pkg/log"
)

func main() {
	logger := log.New() // JSON slog

	cfg := httpserver.Config{
		Port:           envOr("PORT", "8080"),
		DynamoEndpoint: os.Getenv("DYNAMO_ENDPOINT"),
		TableName:      envOr("URLS_TABLE", "urls"),
	}

	srv := httpserver.NewServer(cfg, logger)

	// graceful start
	go func() {
		logger.Info("server_starting", slog.String("port", cfg.Port))
		if err := http.ListenAndServe(":"+cfg.Port, srv.Router); err != nil && err != http.ErrServerClosed {
			logger.Error("server_listen_error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	logger.Info("server_shutting_down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server_shutdown_error", slog.String("error", err.Error()))
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
