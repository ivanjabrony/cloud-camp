package service

import (
	"context"
	"errors"
	"ivanjabrony/cloud-test/internal/logger"
	"ivanjabrony/cloud-test/internal/ratelimit"
	"ivanjabrony/cloud-test/internal/ratelimit/config"
	"ivanjabrony/cloud-test/internal/ratelimit/dto"
	"log/slog"
)

// BucketStorage is a interface for preferably in-memory storage for buckets. Basic implementation uses map[string]*ratelimit.TokenBucket with mutex,
// but you can implement your own, for example sync.Map or Redis storage
type BucketStorage interface {
	Store(ctx context.Context, key string, bucket *ratelimit.TokenBucket)
	Load(ctx context.Context, key string) (bucket *ratelimit.TokenBucket, ok bool)
}

type ConfigurationRepository interface {
	CreateOrUpdate(ctx context.Context, config *dto.UserConfig) (*dto.UserConfig, error)
	GetByIp(ctx context.Context, ip string) (*dto.UserConfig, error)
	GetAll(ctx context.Context) ([]*dto.UserConfig, error)
}

type RateLimitService struct {
	cfg           *config.Config
	logger        *logger.MyLogger
	cfgRepository ConfigurationRepository
	bucketStorage BucketStorage
}

func NewService(cfg *config.Config, logger *logger.MyLogger, cfgRepository ConfigurationRepository, bucketStorage BucketStorage) (*RateLimitService, error) {
	rl := &RateLimitService{cfg, logger, cfgRepository, bucketStorage}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RepositoryTimeout)
	defer cancel()

	configs, err := cfgRepository.GetAll(ctx)
	if err != nil {
		logger.Error("Error in initial service configuration", slog.Any("error", err))
		return nil, err
	}

	for _, config := range configs {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.BucketConfigureTimeout)
		defer cancel()

		err := rl.configureBucket(ctx, config)
		if err != nil {
			logger.Error("Error in initial configuration of buckets from repository", slog.Any("error", err))
			return nil, err
		}
	}

	return rl, nil
}

func (rl *RateLimitService) configureBucket(ctx context.Context, config *dto.UserConfig) error {
	tb, ok := rl.bucketStorage.Load(ctx, config.Ip)
	if !ok {
		tb := ratelimit.NewTokenBucket(config.Capacity, config.RatePerSec)
		rl.bucketStorage.Store(ctx, config.Ip, tb)
		return nil
	}

	tb.UpdateConfig(config.Capacity, config.RatePerSec)

	return nil
}

func (rs *RateLimitService) CreateOrUpdateConfig(ctx context.Context, userConfig *dto.UserConfig) error {
	ctx, cancel := context.WithTimeout(ctx, rs.cfg.RepositoryTimeout)
	defer cancel()

	config, err := rs.cfgRepository.CreateOrUpdate(ctx, userConfig)
	if err != nil {
		rs.logger.Error("Couldn't save or update configuration in repository", slog.Any("error", err))
		return errors.New("couldn't save or update configuration")
	}

	err = rs.configureBucket(ctx, config)
	if err != nil {
		rs.logger.Error("Couldn't save or update bucket configuration", slog.Any("error", err))
		return errors.New("couldn't save or update configuration")
	}
	return nil
}
