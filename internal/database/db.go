package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"be-zor/internal/config"
)

func Open(cfg config.Config) (*bun.DB, error) {
	if cfg.DatabaseURL == "" {
		return nil, errors.New("database url is not configured")
	}

	if err := ensureDatabaseExists(cfg.DatabaseURL); err != nil {
		return nil, err
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.DatabaseURL)))
	sqldb.SetConnMaxIdleTime(5 * time.Minute)
	sqldb.SetConnMaxLifetime(30 * time.Minute)
	sqldb.SetMaxIdleConns(10)
	sqldb.SetMaxOpenConns(20)

	db := bun.NewDB(sqldb, pgdialect.New())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func ensureDatabaseExists(dsn string) error {
	databaseName, adminDSN, err := adminConnectionString(dsn)
	if err != nil {
		return err
	}

	if databaseName == "" {
		return nil
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(adminDSN)))
	defer sqldb.Close()

	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS "%s"`, strings.ReplaceAll(databaseName, `"`, `""`))
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("ensure database exists: %w", err)
	}

	return nil
}

func adminConnectionString(dsn string) (string, string, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", "", fmt.Errorf("parse database url: %w", err)
	}

	databaseName := strings.TrimPrefix(parsed.Path, "/")
	if databaseName == "" {
		return "", dsn, nil
	}

	adminURL := *parsed
	adminURL.Path = "/defaultdb"
	return databaseName, adminURL.String(), nil
}
