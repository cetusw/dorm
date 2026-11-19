package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"dorm/pkg/core/ports"

	"github.com/google/uuid"
)

type SessionRepository struct{ db *sql.DB }

func NewSessionRepository(db *sql.DB) *SessionRepository { return &SessionRepository{db: db} }

func lockSessionUser(ctx context.Context, tx *sql.Tx, id uuid.UUID) (string, error) {
	var hash string
	var deleted sql.NullTime
	err := tx.QueryRowContext(ctx, "SELECT password_hash,deleted_at FROM user WHERE id=? FOR UPDATE", id[:]).Scan(&hash, &deleted)
	if errors.Is(err, sql.ErrNoRows) || deleted.Valid {
		return "", ports.SessionRejection("user_inactive")
	}
	return hash, err
}

func (r *SessionRepository) Create(ctx context.Context, id []byte, uid uuid.UUID, expected string, expires time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	hash, err := lockSessionUser(ctx, tx, uid)
	if err != nil {
		return err
	}
	if hash != expected {
		return ports.SessionRejection("credentials_changed")
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO auth_session(id,user_id,expires_at) VALUES(?,?,?)", id, uid[:], expires); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SessionRepository) Validate(ctx context.Context, id []byte) (*ports.AuthSession, error) {
	var uid []byte
	var expires time.Time
	var revoked, deleted sql.NullTime
	err := r.db.QueryRowContext(ctx, "SELECT s.user_id,s.expires_at,s.revoked_at,u.deleted_at FROM auth_session s JOIN user u ON u.id=s.user_id WHERE s.id=?", id).Scan(&uid, &expires, &revoked, &deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ports.SessionRejection("session_not_found")
	}
	if err != nil {
		return nil, err
	}
	if deleted.Valid {
		return nil, ports.SessionRejection("user_inactive")
	}
	if revoked.Valid {
		return nil, ports.SessionRejection("session_revoked")
	}
	if !time.Now().Before(expires) {
		return nil, ports.SessionRejection("session_expired")
	}
	userID, err := uuid.FromBytes(uid)
	if err != nil {
		return nil, err
	}
	return &ports.AuthSession{UserID: userID, ExpiresAt: expires}, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id []byte) error {
	_, err := r.db.ExecContext(ctx, "UPDATE auth_session SET revoked_at=COALESCE(revoked_at,CURRENT_TIMESTAMP(6)) WHERE id=?", id)
	return err
}
