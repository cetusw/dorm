package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"dorm/pkg/infrastructure/config"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func NewConnection(cfg *config.AppConfig) (*sql.DB, error) {
	return newConnection(cfg, cfg.GetDBConnectionString())
}

// NewMigrationConnection permits long DDL reads without changing backend limits.
func NewMigrationConnection(cfg *config.AppConfig) (*sql.DB, error) {
	options, err := mysqlDriver.ParseDSN(cfg.GetDBConnectionString())
	if err != nil {
		return nil, fmt.Errorf("parse migration connection options: %w", err)
	}
	options.ReadTimeout = 10 * time.Minute
	return newConnection(cfg, options.FormatDSN())
}

func newConnection(cfg *config.AppConfig, dsn string) (*sql.DB, error) {

	db, err := sql.Open(cfg.DBDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open connection error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(fmt.Errorf("ping connection error: %w", err), db.Close())
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	return db, nil
}
