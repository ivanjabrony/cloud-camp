package repository

import (
	"context"
	"errors"
	"fmt"
	"ivanjabrony/cloud-test/internal/logger"
	"ivanjabrony/cloud-test/internal/ratelimit/dto"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type PgxIface interface {
	Begin(context.Context) (pgx.Tx, error)
	Close()
}

// ConfigRepository is a Postgres based repository client for storing clients configurations
type ConfigRepository struct {
	pool    PgxIface
	builder squirrel.StatementBuilderType
	logger  *logger.MyLogger
}

func NewConfigRepository(pool PgxIface, logger *logger.MyLogger) (*ConfigRepository, error) {
	if pool == nil {
		return nil, errors.New("nil values in ConfigRepository constructor")
	}

	return &ConfigRepository{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger,
	}, nil
}

func (repo ConfigRepository) CreateOrUpdate(ctx context.Context, config *dto.UserConfig) (*dto.UserConfig, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if rbErr := tx.Rollback(rollbackCtx); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
			}
		}
	}()

	query, args, err := repo.builder.
		Insert("user_configs").
		Columns("ip", "capacity", "rate_per_sec").
		Values(config.Ip, config.Capacity, config.RatePerSec).
		Suffix("ON CONFLICT (ip) DO UPDATE SET capacity = ?, rate_per_sec = ?", config.Capacity, config.RatePerSec).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	rows.Close()
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	return config, nil
}

func (repo ConfigRepository) GetByIp(ctx context.Context, ip string) (*dto.UserConfig, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			rollbackCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if rbErr := tx.Rollback(rollbackCtx); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
			}
		}
	}()

	query, args, err := repo.builder.
		Select("ip", "capacity", "rate_per_sec").
		From("user_configs").
		Where(squirrel.Eq{"ip": ip}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var config dto.UserConfig
	err = tx.QueryRow(ctx, query, args...).Scan(&config.Ip, &config.Capacity, &config.RatePerSec)
	if err != nil {
		return nil, fmt.Errorf("failed to load user configuration: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	return &config, nil
}

func (repo *ConfigRepository) GetAll(ctx context.Context) ([]*dto.UserConfig, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if rbErr := tx.Rollback(rollbackCtx); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				err = fmt.Errorf("commit failed: %w", commitErr)
			}
		}
	}()

	query, args, err := repo.builder.
		Select("ip", "capacity", "rate_per_sec").
		From("user_configs").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var configs []*dto.UserConfig
	for rows.Next() {
		var config dto.UserConfig
		if err := rows.Scan(
			&config.Ip,
			&config.Capacity,
			&config.RatePerSec,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return configs, nil
}
