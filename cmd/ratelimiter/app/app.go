package app

import (
	"context"
	"errors"
	"ivanjabrony/cloud-test/cmd/ratelimiter/initDB"
	"ivanjabrony/cloud-test/internal/logger"
	"ivanjabrony/cloud-test/internal/ratelimit/config"
	"ivanjabrony/cloud-test/internal/ratelimit/controller/router"
	"ivanjabrony/cloud-test/internal/ratelimit/storage"
	"log/slog"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Application struct {
	cfg     *config.Config
	l       *logger.MyLogger
	router  *router.RLRouter
	storage *storage.BucketStorage
}

func NewApplication(cfg *config.Config, logger *logger.MyLogger) (*Application, func(), error) {
	if cfg == nil || logger == nil {
		return nil, nil, errors.New("cfg and logger must be non nil")
	}

	pool, closeDB, err := initDB.InitDatabase(cfg)
	if err != nil {
		return nil, nil, err
	}

	err = initDB.RunMigrations(pool, cfg.DB.GetConnStr(), "file://migrations")
	if err != nil {
		logger.Error("Error while migrating database", slog.Any("error", err))
		return nil, nil, errors.New("couldn't apply database migrations")
	}

	handlers, storage, err := InitBackend(pool, cfg, logger)
	if err != nil {
		return nil, closeDB, err
	}

	router := router.NewRouter(cfg, logger, handlers.Config, handlers.Ratelimit)

	app := Application{
		cfg:     cfg,
		l:       logger,
		router:  router,
		storage: storage,
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

func (app *Application) Stop(ctx context.Context) error {
	err := app.router.Stop(ctx)
	if err != nil {
		return err
	}
	app.storage.Stop(ctx)
	return nil
}
