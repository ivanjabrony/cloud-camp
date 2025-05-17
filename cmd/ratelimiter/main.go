package main

import (
	"context"
	"ivanjabrony/cloud-test/cmd/ratelimiter/app"
	"ivanjabrony/cloud-test/internal/logger"
	"ivanjabrony/cloud-test/internal/ratelimit/config"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var configPath string
	switch len(os.Args) {
	case 1:
		configPath = "config.json"
	case 2:
		configPath = os.Args[1]
	default:
		log.Fatal("Wrong amount of passed variables, the only possible passed variable is a configuration.json path\n")
	}

	// load config and create logger
	cfg := config.MustLoadConfig(configPath)
	l := logger.New(cfg.Env, cfg.LogFormat)
	l.Info("config loaded successfully", slog.String("env", cfg.Env))

	app, deferFn, err := app.NewApplication(cfg, l)
	if err != nil {
		log.Fatalf("failed to create app instance: %v", err)
	}
	defer deferFn()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run application in separate goroutine
	appErr := make(chan error, 1)
	go func() {
		l.Info("Starting application")
		if err := app.Run(); err != nil {
			appErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-quit:
		l.Info("Received signal, initiating shutdown",
			slog.String("signal", sig.String()))

		// Start graceful shutdown with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer shutdownCancel()

		if err := app.Stop(shutdownCtx); err != nil {
			l.Error("Graceful shutdown failed",
				slog.Any("error", err),
				slog.Duration("timeout", cfg.ShutdownTimeout))
			os.Exit(1)
		}

		l.Info("Application stopped gracefully")

	case err := <-appErr:
		l.Error("Application runtime error", slog.Any("error", err))
		cancel() // Trigger cleanup
		os.Exit(1)
	}

}
