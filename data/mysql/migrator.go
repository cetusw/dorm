package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	migrationsPath = "file://data/mysql/migrations"
)

func ApplyMigrations(db *sql.DB) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("create mysql migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func RollbackMigration(db *sql.DB) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("create mysql migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback migration: %w", err)
	}

	return nil
}

func MigrationVersion(db *sql.DB) (version uint, dirty bool, err error) {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("create mysql migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return 0, false, fmt.Errorf("create migrate instance: %w", err)
	}

	version, dirty, err = m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, fmt.Errorf("get migration version: %w", err)
	}

	return version, dirty, nil
}

func ForceMigrationVersion(db *sql.DB, version int) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("create mysql migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Force(version); err != nil {
		return fmt.Errorf("force migration version: %w", err)
	}

	return nil
}
