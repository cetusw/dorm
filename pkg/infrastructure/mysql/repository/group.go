package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/structure"
)

type GroupRepository struct {
	db *sql.DB
}

func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) FindByID(ctx context.Context, id uuid.UUID) (*structure.Group, error) {
	const query = "SELECT id, name, spreadsheet_id, dormitory_id FROM `group` WHERE id = ?"

	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var gID []byte
	var name, sheetID string
	var dormID int64

	if err := row.Scan(&gID, &name, &sheetID, &dormID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindGroupByID: %w", err)
	}

	groupID, _ := uuid.FromBytes(gID)
	return structure.RestoreGroup(groupID, name, sheetID, dormID), nil
}

func (r *GroupRepository) FindAll(ctx context.Context) ([]*structure.Group, error) {
	const query = "SELECT id, name, spreadsheet_id, dormitory_id FROM `group`"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*structure.Group
	for rows.Next() {
		var gID []byte
		var name, sheetID string
		var dormID int64
		if err := rows.Scan(&gID, &name, &sheetID, &dormID); err != nil {
			return nil, err
		}
		uid, _ := uuid.FromBytes(gID)
		groups = append(groups, structure.RestoreGroup(uid, name, sheetID, dormID))
	}
	return groups, nil
}
