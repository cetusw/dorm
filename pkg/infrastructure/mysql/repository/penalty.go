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

func (r *PenaltyRepository) FindByID(ctx context.Context, id uuid.UUID) (*penaltydomain.Penalty, error) {
	const query = `
		SELECT id, user_id, weight, reason, issued_on, created_at, resolved_at
		FROM penalty
		WHERE id = ?
	`

	idBytes, err := id.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal penalty id: %w", err)
	}

	row := r.db.QueryRowContext(ctx, query, idBytes)

	var penaltyIDBytes []byte
	var userIDBytes []byte
	var weight float64
	var reason string
	var issuedOn time.Time
	var createdAt time.Time
	var resolvedAt sql.NullTime

	if err := row.Scan(
		&penaltyIDBytes,
		&userIDBytes,
		&weight,
		&reason,
		&issuedOn,
		&createdAt,
		&resolvedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find penalty by id: %w", err)
	}

	penaltyID, err := uuid.FromBytes(penaltyIDBytes)
	if err != nil {
		return nil, fmt.Errorf("decode penalty id: %w", err)
	}
	userID, err := uuid.FromBytes(userIDBytes)
	if err != nil {
		return nil, fmt.Errorf("decode penalty user id: %w", err)
	}

	var resolvedAtPtr *time.Time
	if resolvedAt.Valid {
		resolvedAtValue := resolvedAt.Time
		resolvedAtPtr = &resolvedAtValue
	}

	entity, err := penaltydomain.RestorePenalty(
		penaltyID,
		userID,
		weight,
		reason,
		issuedOn,
		createdAt,
		resolvedAtPtr,
	)
	if err != nil {
		return nil, fmt.Errorf("restore penalty: %w", err)
	}

	return entity, nil
}

func (r *PenaltyRepository) Save(ctx context.Context, penalty *penaltydomain.Penalty) error {
	const query = `
		INSERT INTO penalty (id, user_id, weight, reason, issued_on, created_at, resolved_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			resolved_at = VALUES(resolved_at)
	`

	penaltyID, err := penalty.ID().MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal penalty id: %w", err)
	}
	userID, err := penalty.UserID().MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal penalty user id: %w", err)
	}

	var resolvedAt interface{}
	resolvedAtValue := penalty.ResolvedAt()
	if resolvedAtValue != nil {
		resolvedAt = *resolvedAtValue
	}

	if _, err := r.db.ExecContext(
		ctx,
		query,
		penaltyID,
		userID,
		fmt.Sprintf("%.1f", penalty.Weight()),
		penalty.Reason(),
		penalty.IssuedOn(),
		penalty.CreatedAt(),
		resolvedAt,
	); err != nil {
		return fmt.Errorf("save penalty: %w", err)
	}

	return nil
}
