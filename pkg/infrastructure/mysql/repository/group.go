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
	const query = "SELECT id, leader_id, name, spreadsheet_id, dormitory_id, next_duty_team FROM `group` WHERE id = ?"

	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var gID, leaderIDBytes []byte
	var name string
	var sheetID sql.NullString
	var dormID int64
	var nextDutyTeam sql.NullInt64

	if err := row.Scan(&gID, &leaderIDBytes, &name, &sheetID, &dormID, &nextDutyTeam); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindGroupByID: %w", err)
	}

	groupID, _ := uuid.FromBytes(gID)
	return structure.RestoreGroup(groupID, uuidPtrFromBytes(leaderIDBytes), name, sheetID.String, dormID, intPtrFromNull(nextDutyTeam)), nil
}

func (r *GroupRepository) FindAll(ctx context.Context) ([]*structure.Group, error) {
	const query = "SELECT id, leader_id, name, spreadsheet_id, dormitory_id, next_duty_team FROM `group` ORDER BY dormitory_id, name"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*structure.Group
	for rows.Next() {
		var gID, leaderIDBytes []byte
		var name string
		var sheetID sql.NullString
		var dormID int64
		var nextDutyTeam sql.NullInt64
		if err := rows.Scan(&gID, &leaderIDBytes, &name, &sheetID, &dormID, &nextDutyTeam); err != nil {
			return nil, err
		}
		uid, _ := uuid.FromBytes(gID)
		groups = append(groups, structure.RestoreGroup(uid, uuidPtrFromBytes(leaderIDBytes), name, sheetID.String, dormID, intPtrFromNull(nextDutyTeam)))
	}
	return groups, nil
}

func (r *GroupRepository) FindByDormitoryID(ctx context.Context, dormitoryID int64) ([]*structure.Group, error) {
	const query = "SELECT id, leader_id, name, spreadsheet_id, dormitory_id, next_duty_team FROM `group` WHERE dormitory_id = ? ORDER BY name"

	rows, err := r.db.QueryContext(ctx, query, dormitoryID)
	if err != nil {
		return nil, fmt.Errorf("FindGroupsByDormitoryID: %w", err)
	}
	defer rows.Close()

	var groups []*structure.Group
	for rows.Next() {
		var gID, leaderIDBytes []byte
		var name string
		var sheetID sql.NullString
		var dormID int64
		var nextDutyTeam sql.NullInt64
		if err := rows.Scan(&gID, &leaderIDBytes, &name, &sheetID, &dormID, &nextDutyTeam); err != nil {
			return nil, fmt.Errorf("FindGroupsByDormitoryID scan: %w", err)
		}
		uid, _ := uuid.FromBytes(gID)
		groups = append(groups, structure.RestoreGroup(uid, uuidPtrFromBytes(leaderIDBytes), name, sheetID.String, dormID, intPtrFromNull(nextDutyTeam)))
	}
	return groups, rows.Err()
}

func (r *GroupRepository) Save(ctx context.Context, group *structure.Group) error {
	const query = `
		INSERT INTO ` + "`group`" + ` (id, leader_id, name, spreadsheet_id, dormitory_id, next_duty_team)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			leader_id = VALUES(leader_id),
			name = VALUES(name),
			spreadsheet_id = VALUES(spreadsheet_id),
			dormitory_id = VALUES(dormitory_id),
			next_duty_team = VALUES(next_duty_team)
	`
	idBytes, _ := group.ID().MarshalBinary()
	var leaderID interface{} = nil
	if group.LeaderID() != nil {
		leaderID, _ = group.LeaderID().MarshalBinary()
	}
	var spreadsheetID interface{} = nil
	if group.SpreadsheetID() != "" {
		spreadsheetID = group.SpreadsheetID()
	}
	var nextDutyTeam interface{} = nil
	if group.NextDutyTeam() != nil {
		nextDutyTeam = *group.NextDutyTeam()
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		idBytes,
		leaderID,
		group.Name(),
		spreadsheetID,
		group.DormitoryID(),
		nextDutyTeam,
	)
	if err != nil {
		return fmt.Errorf("SaveGroup: %w", err)
	}
	return nil
}

func intPtrFromNull(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	intValue := int(value.Int64)
	return &intValue
}

func uuidPtrFromBytes(value []byte) *uuid.UUID {
	if len(value) == 0 {
		return nil
	}
	id, _ := uuid.FromBytes(value)
	return &id
}

func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = "DELETE FROM `group` WHERE id = ?"
	idBytes, _ := id.MarshalBinary()
	_, err := r.db.ExecContext(ctx, query, idBytes)
	if err != nil {
		return fmt.Errorf("DeleteGroup: %w", err)
	}
	return nil
}
