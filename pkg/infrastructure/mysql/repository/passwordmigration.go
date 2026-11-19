package repository

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"

	useruc "dorm/pkg/core/usecase/user"
)

func MigratePasswords(ctx context.Context, db *sql.DB, plan map[string]string, dryRun bool) (int, error) {
	if err := checkActiveLogins(ctx, db); err != nil {
		return 0, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, errors.New("cannot begin credential migration")
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, "SELECT id,login,password_hash FROM user WHERE deleted_at IS NULL ORDER BY id FOR UPDATE")
	if err != nil {
		return 0, errors.New("cannot lock active credentials")
	}
	type update struct {
		id   []byte
		hash string
	}
	var updates []update
	seen := make(map[string]bool)
	stored := make(map[string]bool)
	for rows.Next() {
		var id []byte
		var hash, login string
		if err := rows.Scan(&id, &login, &hash); err != nil {
			rows.Close()
			return 0, errors.New("cannot read active credentials")
		}
		login = strings.ToLower(strings.TrimSpace(login))
		if stored[login] {
			rows.Close()
			return 0, errors.New("conflicting active logins require manual resolution")
		}
		stored[login] = true
		password, included := plan[login]
		if included {
			seen[login] = true
		}
		if useruc.ValidPasswordHash(hash) {
			if included && !useruc.VerifyPassword(password, hash) {
				rows.Close()
				return 0, errors.New("password mismatch for migrated user")
			}
			continue
		}
		old, decodeErr := hex.DecodeString(hash)
		expected := sha256.Sum256([]byte(password))
		if !included || decodeErr != nil || len(old) != 32 || subtle.ConstantTimeCompare(old, expected[:]) != 1 {
			rows.Close()
			return 0, errors.New("incomplete credential input or password mismatch")
		}
		newHash, err := useruc.HashPassword(password)
		if err != nil {
			rows.Close()
			return 0, errors.New("cannot generate password hash")
		}
		updates = append(updates, update{id: id, hash: newHash})
	}
	readErr := rows.Err()
	closeErr := rows.Close()
	if readErr != nil || closeErr != nil {
		return 0, errors.New("cannot finish reading active credentials")
	}
	if len(seen) != len(plan) {
		return 0, errors.New("credential input includes unknown or inactive user")
	}
	if dryRun {
		return len(updates), nil
	}
	for _, u := range updates {
		result, err := tx.ExecContext(ctx, "UPDATE user SET password_hash=? WHERE id=? AND deleted_at IS NULL", u.hash, u.id)
		if err != nil {
			return 0, errors.New("credential update failed; transaction rolled back")
		}
		count, err := result.RowsAffected()
		if err != nil || count != 1 {
			return 0, errors.New("unexpected credential update count")
		}
	}
	if err := checkActivePasswordHashes(ctx, tx); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, errors.New("credential commit outcome unknown; run dry-run again before retrying")
	}
	return len(updates), nil
}
