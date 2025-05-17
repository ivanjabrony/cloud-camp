package app

import (
	"errors"
	"ivanjabrony/cloud-test/internal/logger"
	"ivanjabrony/cloud-test/internal/ratelimit"
	"ivanjabrony/cloud-test/internal/ratelimit/config"
	"ivanjabrony/cloud-test/internal/ratelimit/controller/handler"
	"ivanjabrony/cloud-test/internal/ratelimit/repository"
	"ivanjabrony/cloud-test/internal/ratelimit/service"
	"ivanjabrony/cloud-test/internal/ratelimit/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitBackend(pool *pgxpool.Pool, cfg *config.Config, logger *logger.MyLogger) (*Handlers, *storage.BucketStorage, error) {
	if pool == nil || cfg == nil || logger == nil {
		return nil, nil, errors.New("nil values in init constructor")
	}

	storage, err := initStorages()
	if err != nil {
		return nil, nil, err
	}

	repository, err := initRepositories(pool, logger)
	if err != nil {
		return nil, nil, err
	}

	services, err := initServices(repository, storage, cfg, logger)
	if err != nil {
		return nil, nil, err
	}

	ratelimiter, err := initRatelimiter(storage, cfg, logger)
	if err != nil {
		return nil, nil, err
	}

	handlers, err := initHandlers(services, ratelimiter, cfg, logger)
	if err != nil {
		return nil, nil, err
	}

	return handlers, storage.BucketStorage, nil
}

type storages struct {
	BucketStorage *storage.BucketStorage
}

type repositories struct {
	configRepo *repository.ConfigRepository
}

type services struct {
	ratelimit *service.RateLimitService
}

type Handlers struct {
	Config    *handler.ConfigHandler
	Ratelimit *handler.RateLimitHandler
}

func initStorages() (*storages, error) {
	storage := storage.NewBucketStorage()
	return &storages{storage}, nil
}

func initRepositories(pool *pgxpool.Pool, logger *logger.MyLogger) (*repositories, error) {
	repo, err := repository.NewConfigRepository(pool, logger)
	if err != nil {
		return nil, err
	}

	return &repositories{repo}, nil
}

func initServices(repo *repositories, storage *storages, cfg *config.Config, logger *logger.MyLogger) (*services, error) {
	service, err := service.NewService(cfg, logger, repo.configRepo, storage.BucketStorage)
	if err != nil {
		return nil, err
	}

	return &services{service}, nil
}

func initRatelimiter(storage *storages, cfg *config.Config, logger *logger.MyLogger) (*ratelimit.RateLimiter, error) {
	ratelimiter, err := ratelimit.NewRateLimiter(storage.BucketStorage, cfg.UserConfig.Tokens, float64(cfg.UserConfig.RatePerSec))
	if err != nil {
		return nil, err
	}

	return ratelimiter, nil
}

func initHandlers(s *services, ratelimiter *ratelimit.RateLimiter, cfg *config.Config, logger *logger.MyLogger) (*Handlers, error) {
	configHandler, err := handler.NewConfigHandler(cfg, logger, s.ratelimit)
	if err != nil {
		return nil, err
	}
	ratelimitHandler, err := handler.NewRateLimitHandler(cfg, logger, ratelimiter)
	if err != nil {
		return nil, err
	}

	return &Handlers{configHandler, ratelimitHandler}, nil
}
