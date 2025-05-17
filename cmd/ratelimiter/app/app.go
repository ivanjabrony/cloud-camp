package app

import (
	"errors"
	"ivanjabrony/cloud-test/cmd/ratelimiter/initDB"
	"ivanjabrony/cloud-test/internal/logger"
	"ivanjabrony/cloud-test/internal/ratelimit/config"
	"ivanjabrony/cloud-test/internal/ratelimit/controller/router"
	"log/slog"
)

type Application struct {
	cfg    *config.Config
	l      *logger.MyLogger
	router *router.RLRouter
}

func NewApplication(cfg *config.Config, logger *logger.MyLogger) (*Application, func(), error) {
	if cfg == nil || logger == nil {
		return nil, nil, errors.New("cfg and logger must be non nil")
	}

	pool, closeDB, err := initDB.InitDatabase(cfg)
	if err != nil {
		return nil, nil, err
	}

	err = initDB.RunMigrations(pool, cfg.DB.GetConnStr(), "./migrations")
	if err != nil {
		logger.Error("Error while migrating database", slog.Any("error", err))
		return nil, nil, errors.New("couldn't apply database migrations")
	}

	handlers, err := InitializeHandlers(pool, cfg, logger)
	if err != nil {
		return nil, closeDB, err
	}

	router := router.NewRouter(cfg, logger, handlers.Config, handlers.Ratelimit)

	app := Application{
		cfg:    cfg,
		l:      logger,
		router: router,
	}

	return &app, closeDB, nil
}

func (app *Application) Run() error {
	app.l.Info("Running application")

	app.l.Info("Starting HTTP", slog.Int("port", app.cfg.Port))
	go func() {
		if err := app.router.Run(); err != nil {
			app.l.Error("Error while running router", slog.Any("error", err))
			return
		}
	}()

	return nil
}

func (app *Application) Stop() {
	//TODO, send cancel to all context with timeout so all requests can finish
}
