package main

import (
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

	go func() {
		if err := app.Run(); err != nil {
			log.Fatalf("app encountered an error while running: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // TODO app.Stop() with grateful shutdown
	l.Info("Gracefully shutting down application")
}
