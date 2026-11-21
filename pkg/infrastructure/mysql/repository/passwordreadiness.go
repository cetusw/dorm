package repository

import (
	"context"
	"database/sql"
	"errors"

	useruc "dorm/pkg/core/usecase/user"
)

func CheckPasswordReadiness(ctx context.Context, db *sql.DB) error {
	if err := checkActiveLogins(ctx, db); err != nil {
		return err
	}
	return checkActivePasswordHashes(ctx, db)
}

func checkActiveLogins(ctx context.Context, db *sql.DB) error {
	var duplicates int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM (SELECT login FROM user WHERE deleted_at IS NULL GROUP BY login HAVING COUNT(*)>1) conflicts").Scan(&duplicates)
	if err != nil {
		return errors.New("cannot check active login uniqueness")
	}
	if duplicates != 0 {
		return errors.New("duplicate active logins; credential readiness check failed")
	}
	return nil
}

func checkActivePasswordHashes(ctx context.Context, db interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}) error {
	rows, err := db.QueryContext(ctx, "SELECT password_hash FROM user WHERE deleted_at IS NULL")
	if err != nil {
		return errors.New("cannot check credential readiness")
	}
	defer rows.Close()
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return errors.New("cannot read credential readiness")
		}
		if !useruc.ValidPasswordHash(hash) {
			return errors.New("active credentials contain an invalid password hash; backend startup blocked")
		}
	}
	if rows.Err() != nil {
		return errors.New("cannot finish credential readiness check")
	}
	return nil
}
