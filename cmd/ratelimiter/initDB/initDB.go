package initDB

import (
	"context"
	"errors"
	"ivanjabrony/cloud-test/internal/ratelimit/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func InitDatabase(cfg *config.Config) (*pgxpool.Pool, func(), error) {
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(cfg.DB.GetConnStr())
	if err != nil {
		return nil, nil, err
	}

	config.MaxConns = cfg.DB.MaxConns
	config.MinConns = cfg.DB.MinConns
	config.MaxConnLifetime = cfg.DB.MaxConnLifetime

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, nil, err
	}

	return db, db.Close, nil
}

func RunMigrations(db *pgxpool.Pool, dbName string, sourceMigration string) error {
	conn, err := db.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	sqlDB := stdlib.OpenDBFromPool(db)
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		sourceMigration,
		dbName,
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func DownMigrations(db *pgxpool.Pool, dbName string, sourceMigration string) error {
	sqlDB := stdlib.OpenDBFromPool(db)
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		sourceMigration,
		dbName,
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
