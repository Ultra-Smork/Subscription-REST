package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	app "github.com/Ultra-Smork/Subscription-service/internals"
	"github.com/Ultra-Smork/Subscription-service/internals/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}

	slog.Info("application stopped gracefully")
}
