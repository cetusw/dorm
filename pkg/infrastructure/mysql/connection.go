package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"dorm/pkg/infrastructure/config"

	_ "github.com/go-sql-driver/mysql"
)

func NewConnection(cfg *config.AppConfig) (*sql.DB, error) {
	dsn := cfg.GetDBConnectionString()

	db, err := sql.Open(cfg.DBDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open connection error: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping connection error: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
