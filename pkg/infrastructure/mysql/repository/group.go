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
	const query = "SELECT id, leader_id, name, dormitory_id FROM `group` WHERE id = ?"

	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var gID, leaderIDBytes []byte
	var name string
	var dormID int64

	if err := row.Scan(&gID, &leaderIDBytes, &name, &dormID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindGroupByID: %w", err)
	}

	groupID, _ := uuid.FromBytes(gID)
	return structure.RestoreGroup(groupID, uuidPtrFromBytes(leaderIDBytes), name, dormID), nil
}

func (r *GroupRepository) FindAll(ctx context.Context) ([]*structure.Group, error) {
	const query = "SELECT id, leader_id, name, dormitory_id FROM `group` ORDER BY dormitory_id, name"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*structure.Group
	for rows.Next() {
		var gID, leaderIDBytes []byte
		var name string
		var dormID int64
		if err := rows.Scan(&gID, &leaderIDBytes, &name, &dormID); err != nil {
			return nil, err
		}
		uid, _ := uuid.FromBytes(gID)
		groups = append(groups, structure.RestoreGroup(uid, uuidPtrFromBytes(leaderIDBytes), name, dormID))
	}
	return groups, nil
}

func (r *GroupRepository) FindByDormitoryID(ctx context.Context, dormitoryID int64) ([]*structure.Group, error) {
	const query = "SELECT id, leader_id, name, dormitory_id FROM `group` WHERE dormitory_id = ? ORDER BY name"

	rows, err := r.db.QueryContext(ctx, query, dormitoryID)
	if err != nil {
		return nil, fmt.Errorf("FindGroupsByDormitoryID: %w", err)
	}
	defer rows.Close()

	var groups []*structure.Group
	for rows.Next() {
		var gID, leaderIDBytes []byte
		var name string
		var dormID int64
		if err := rows.Scan(&gID, &leaderIDBytes, &name, &dormID); err != nil {
			return nil, fmt.Errorf("FindGroupsByDormitoryID scan: %w", err)
		}
		uid, _ := uuid.FromBytes(gID)
		groups = append(groups, structure.RestoreGroup(uid, uuidPtrFromBytes(leaderIDBytes), name, dormID))
	}
	return groups, rows.Err()
}

func (r *GroupRepository) Save(ctx context.Context, group *structure.Group) error {
	const query = `
		INSERT INTO ` + "`group`" + ` (id, leader_id, name, dormitory_id)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			leader_id = VALUES(leader_id),
			name = VALUES(name),
			dormitory_id = VALUES(dormitory_id)
	`
	idBytes, _ := group.ID().MarshalBinary()
	var leaderID interface{} = nil
	if group.LeaderID() != nil {
		leaderID, _ = group.LeaderID().MarshalBinary()
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		idBytes,
		leaderID,
		group.Name(),
		group.DormitoryID(),
	)
	if err != nil {
		return fmt.Errorf("SaveGroup: %w", err)
	}
	return nil
}

func uuidPtrFromBytes(value []byte) *uuid.UUID {
	if len(value) == 0 {
		return nil
	}
	id, _ := uuid.FromBytes(value)
	return &id
}

func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	idBytes, _ := id.MarshalBinary()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete group transaction: %w", err)
	}
	defer tx.Rollback()

	// fk_duty_team is RESTRICT to protect a Team's history from direct deletion.
	// Group deletion intentionally retains its previous cascade semantics, so Duty
	// roots are removed first and their dependent task/participant rows cascade.
	if _, err := tx.ExecContext(ctx, "DELETE FROM duty WHERE group_id = ?", idBytes); err != nil {
		return fmt.Errorf("delete group duties: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM `group` WHERE id = ?", idBytes); err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete group transaction: %w", err)
	}
	return nil
}
