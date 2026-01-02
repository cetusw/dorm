package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/structure"
)

type StructureRepository struct {
	db *sql.DB
}

func NewStructureRepository(db *sql.DB) *StructureRepository {
	return &StructureRepository{db: db}
}

func (r *StructureRepository) FindTeamByID(ctx context.Context, id uuid.UUID) (*structure.Team, error) {
	const query = `SELECT id, group_id, leader_id, color, team_order FROM team WHERE id = ?`

	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var tID, gID, lID []byte
	var color string
	var order int

	if err := row.Scan(&tID, &gID, &lID, &color, &order); err != nil {
		return nil, fmt.Errorf("FindTeamByID: %w", err)
	}

	teamID, _ := uuid.FromBytes(tID)
	groupID, _ := uuid.FromBytes(gID)
	var leaderID *uuid.UUID
	if len(lID) > 0 {
		uid, _ := uuid.FromBytes(lID)
		leaderID = &uid
	}

	return structure.RestoreTeam(teamID, groupID, leaderID, color, order), nil
}

func (r *StructureRepository) FindGroupByID(ctx context.Context, id uuid.UUID) (*structure.Group, error) {
	const query = "SELECT id, name, spreadsheet_id, dormitory_id FROM `group` WHERE id = ?"

	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var gID []byte
	var name, sheetID string
	var dormID int64

	if err := row.Scan(&gID, &name, &sheetID, &dormID); err != nil {
		return nil, fmt.Errorf("FindGroupByID: %w", err)
	}

	groupID, _ := uuid.FromBytes(gID)
	return structure.RestoreGroup(groupID, name, sheetID, dormID), nil
}

func (r *StructureRepository) FindAllGroups(ctx context.Context) ([]*structure.Group, error) {
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

// --- DORMITORY ---
