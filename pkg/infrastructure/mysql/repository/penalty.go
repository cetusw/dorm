package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	penaltydomain "dorm/pkg/core/domain/penalty"

	"github.com/google/uuid"
)

type PenaltyRepository struct {
	db *sql.DB
}

func NewPenaltyRepository(db *sql.DB) *PenaltyRepository {
	return &PenaltyRepository{db: db}
}

func (r *PenaltyRepository) Create(ctx context.Context, penalty *penaltydomain.Penalty) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin penalty create transaction: %w", err)
	}

	if err := lockPenaltyResident(ctx, tx, penalty.UserID()); err != nil {
		_ = tx.Rollback()
		return err
	}

	balance, err := penaltyBalance(ctx, tx, penalty.UserID())
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if balance < 0 {
		_ = tx.Rollback()
		return penaltydomain.ErrNegativePenaltyHistory
	}

	if err := insertPenaltyEntry(ctx, tx, penalty); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit penalty create transaction: %w", err)
	}

	return nil
}

func (r *PenaltyRepository) CreateResolve(ctx context.Context, penalty *penaltydomain.Penalty) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin penalty resolve transaction: %w", err)
	}

	if err := lockPenaltyResident(ctx, tx, penalty.UserID()); err != nil {
		_ = tx.Rollback()
		return err
	}

	balance, err := penaltyBalance(ctx, tx, penalty.UserID())
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if balance < 0 || balance < penalty.Weight() {
		_ = tx.Rollback()
		return penaltydomain.ErrInsufficientPenaltyBalance
	}

	if err := insertPenaltyEntry(ctx, tx, penalty); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit penalty resolve transaction: %w", err)
	}

	return nil
}

func (r *PenaltyRepository) FindByID(ctx context.Context, id uuid.UUID) (*penaltydomain.Penalty, error) {
	const query = `
		SELECT id, user_id, type, weight, reason, created_at
		FROM penalty_entry
		WHERE id = ?
	`

	idBytes, err := marshalUUID(id, "penalty id")
	if err != nil {
		return nil, err
	}

	var entryIDBytes []byte
	var userIDBytes []byte
	var entryType string
	var weight float64
	var reason string
	var createdAt sql.NullTime

	if err := r.db.QueryRowContext(ctx, query, idBytes).Scan(
		&entryIDBytes,
		&userIDBytes,
		&entryType,
		&weight,
		&reason,
		&createdAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find penalty entry: %w", err)
	}

	return restorePenaltyEntry(entryIDBytes, userIDBytes, entryType, weight, reason, createdAt.Time)
}

func (r *PenaltyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin penalty delete transaction: %w", err)
	}

	entry, err := loadPenaltyEntry(ctx, tx, id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := lockPenaltyResident(ctx, tx, entry.UserID()); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := validatePenaltyBalanceAfterDelete(ctx, tx, entry); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := deletePenaltyEntry(ctx, tx, entry.ID()); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit penalty delete transaction: %w", err)
	}

	return nil
}

func (r *PenaltyRepository) Update(ctx context.Context, penalty *penaltydomain.Penalty) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin penalty update transaction: %w", err)
	}

	currentEntry, err := loadPenaltyEntry(ctx, tx, penalty.ID())
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := lockPenaltyResident(ctx, tx, currentEntry.UserID()); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := validatePenaltyBalanceAfterReplace(ctx, tx, currentEntry, penalty); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := updatePenaltyEntry(ctx, tx, penalty); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit penalty update transaction: %w", err)
	}

	return nil
}

type penaltyEntryWriter interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type penaltyEntryReader interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func insertPenaltyEntry(ctx context.Context, writer penaltyEntryWriter, penalty *penaltydomain.Penalty) error {
	const query = `
		INSERT INTO penalty_entry (id, user_id, type, weight, reason, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	penaltyID, err := marshalUUID(penalty.ID(), "penalty id")
	if err != nil {
		return err
	}
	userID, err := marshalUUID(penalty.UserID(), "penalty user id")
	if err != nil {
		return err
	}

	if _, err := writer.ExecContext(
		ctx,
		query,
		penaltyID,
		userID,
		string(penalty.Type()),
		fmt.Sprintf("%.1f", penalty.Weight()),
		penalty.Reason(),
		penalty.CreatedAt(),
	); err != nil {
		return fmt.Errorf("insert penalty entry: %w", err)
	}

	return nil
}

func loadPenaltyEntry(ctx context.Context, reader penaltyEntryReader, id uuid.UUID) (*penaltydomain.Penalty, error) {
	const query = `
		SELECT id, user_id, type, weight, reason, created_at
		FROM penalty_entry
		WHERE id = ?
	`

	idBytes, err := marshalUUID(id, "penalty id")
	if err != nil {
		return nil, err
	}

	var entryIDBytes []byte
	var userIDBytes []byte
	var entryType string
	var weight float64
	var reason string
	var createdAt time.Time

	if err := reader.QueryRowContext(ctx, query, idBytes).Scan(
		&entryIDBytes,
		&userIDBytes,
		&entryType,
		&weight,
		&reason,
		&createdAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, penaltydomain.ErrPenaltyEntryNotFound
		}
		return nil, fmt.Errorf("load penalty entry: %w", err)
	}

	return restorePenaltyEntry(entryIDBytes, userIDBytes, entryType, weight, reason, createdAt)
}

func restorePenaltyEntry(
	entryIDBytes []byte,
	userIDBytes []byte,
	entryType string,
	weight float64,
	reason string,
	createdAt time.Time,
) (*penaltydomain.Penalty, error) {
	entryID, err := uuid.FromBytes(entryIDBytes)
	if err != nil {
		return nil, fmt.Errorf("decode penalty entry id: %w", err)
	}
	userID, err := uuid.FromBytes(userIDBytes)
	if err != nil {
		return nil, fmt.Errorf("decode penalty entry user id: %w", err)
	}

	entry, err := penaltydomain.RestorePenalty(
		entryID,
		userID,
		penaltydomain.EntryType(entryType),
		weight,
		reason,
		createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("restore penalty entry: %w", err)
	}

	return entry, nil
}

func lockPenaltyResident(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	const query = `
		SELECT id
		FROM user
		WHERE id = ?
		FOR UPDATE
	`

	userIDBytes, err := marshalUUID(userID, "penalty user id")
	if err != nil {
		return err
	}

	var lockedID []byte
	if err := tx.QueryRowContext(ctx, query, userIDBytes).Scan(&lockedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return penaltydomain.ErrResidentNotFound
		}
		return fmt.Errorf("lock penalty resident: %w", err)
	}

	return nil
}

func penaltyBalance(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(
			CASE
				WHEN type = 'ISSUE' THEN weight
				WHEN type = 'RESOLVE' THEN -weight
				ELSE 0
			END
		), 0)
		FROM penalty_entry
		WHERE user_id = ?
	`

	userIDBytes, err := marshalUUID(userID, "penalty user id")
	if err != nil {
		return 0, err
	}

	var balance float64
	if err := tx.QueryRowContext(ctx, query, userIDBytes).Scan(&balance); err != nil {
		return 0, fmt.Errorf("calculate penalty balance: %w", err)
	}

	return balance, nil
}

func validatePenaltyBalanceAfterDelete(ctx context.Context, tx *sql.Tx, target *penaltydomain.Penalty) error {
	return validatePenaltyBalanceAfterReplace(ctx, tx, target, nil)
}

func validatePenaltyBalanceAfterReplace(
	ctx context.Context,
	tx *sql.Tx,
	target *penaltydomain.Penalty,
	replacement *penaltydomain.Penalty,
) error {
	const query = `
		SELECT id, type, weight
		FROM penalty_entry
		WHERE user_id = ?
		ORDER BY created_at ASC, id ASC
		FOR UPDATE
	`

	userIDBytes, err := marshalUUID(target.UserID(), "penalty user id")
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, query, userIDBytes)
	if err != nil {
		return fmt.Errorf("query penalty history for validation: %w", err)
	}
	defer rows.Close()

	balance := 0.0
	found := false
	for rows.Next() {
		var entryIDBytes []byte
		var entryType string
		var weight float64
		if err := rows.Scan(&entryIDBytes, &entryType, &weight); err != nil {
			return fmt.Errorf("scan penalty history for validation: %w", err)
		}

		entryID, err := uuid.FromBytes(entryIDBytes)
		if err != nil {
			return fmt.Errorf("decode penalty history id for validation: %w", err)
		}

		if entryID == target.ID() {
			found = true
			if replacement == nil {
				continue
			}

			entryType = string(replacement.Type())
			weight = replacement.Weight()
		}

		if entryType == string(penaltydomain.EntryTypeResolve) {
			balance -= weight
		} else {
			balance += weight
		}

	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate penalty history for validation: %w", err)
	}

	if !found {
		return penaltydomain.ErrPenaltyEntryNotFound
	}
	if balance < 0 {
		return penaltydomain.ErrNegativePenaltyHistory
	}

	return nil
}

func updatePenaltyEntry(ctx context.Context, tx *sql.Tx, penalty *penaltydomain.Penalty) error {
	const query = `
		UPDATE penalty_entry
		SET weight = ?, reason = ?
		WHERE id = ?
	`

	idBytes, err := marshalUUID(penalty.ID(), "penalty id")
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		query,
		fmt.Sprintf("%.1f", penalty.Weight()),
		penalty.Reason(),
		idBytes,
	); err != nil {
		return fmt.Errorf("update penalty entry: %w", err)
	}

	return nil
}

func deletePenaltyEntry(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	const query = `DELETE FROM penalty_entry WHERE id = ?`

	idBytes, err := marshalUUID(id, "penalty id")
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, query, idBytes); err != nil {
		return fmt.Errorf("delete penalty entry: %w", err)
	}

	return nil
}
