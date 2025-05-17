package main

import (
	"context"
	"fmt"
	"ivanjabrony/cloud-test/internal/balancer"
	"ivanjabrony/cloud-test/internal/balancer/config"
	"ivanjabrony/cloud-test/internal/logger"
	"log"
	"log/slog"
	"net/http"
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

	cfg := config.MustLoadConfig(configPath)

	logger := logger.New(cfg.Env, cfg.LogFormat)
	global := balancer.NewPool(logger, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: http.HandlerFunc(balancer.LoadBalancer(logger, cfg, global)),
	}

	go balancer.HealthCheckRoutine(ctx, logger, cfg, global)

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("Load Balancer started at :%d", slog.Int("port", cfg.Port))
		serverErr <- server.ListenAndServe()
	}()

	select {
	case sig := <-sigChan:
		logger.Info("Received signal, shutting down...", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error", slog.Any("error", err))
		}

		logger.Info("Server stopped")

	case err := <-serverErr:
		logger.Error("Server failed", slog.Any("error", err))
		os.Exit(1)
	}
}
