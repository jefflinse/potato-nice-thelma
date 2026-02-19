package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jefflinse/potato-nice-thelma/internal/cataas"
	"github.com/jefflinse/potato-nice-thelma/internal/config"
	"github.com/jefflinse/potato-nice-thelma/internal/meme"
	"github.com/jefflinse/potato-nice-thelma/internal/potato"
	"github.com/jefflinse/potato-nice-thelma/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	memeGen, err := meme.NewGenerator()
	if err != nil {
		slog.Error("failed to create meme generator", "error", err)
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	potatoClient := potato.NewRedditClient(httpClient)
	cataasClient := cataas.NewClient(httpClient)

	srv := server.NewServer(potatoClient, cataasClient, memeGen, httpClient)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort("", cfg.Port),
		Handler: srv,
	}

	go func() {
		slog.Info("starting server", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
