package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/structure"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) FindByID(ctx context.Context, id uuid.UUID) (*structure.Team, error) {
	const query = `SELECT id, name, group_id, leader_id, color, rotation_position FROM team WHERE id = ?`

	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var tID, gID, lID []byte
	var name string
	var color sql.NullString
	var rotationPosition int

	if err := row.Scan(&tID, &name, &gID, &lID, &color, &rotationPosition); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindTeamByID: %w", err)
	}

	teamID, _ := uuid.FromBytes(tID)
	groupID, _ := uuid.FromBytes(gID)
	var leaderID *uuid.UUID
	if len(lID) > 0 {
		uid, _ := uuid.FromBytes(lID)
		leaderID = &uid
	}

	return structure.RestoreTeam(teamID, name, groupID, leaderID, color.String, rotationPosition), nil
}

// TODO: вынести в DTO, как в user
func (r *TeamRepository) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*structure.Team, error) {
	const query = `SELECT id, name, group_id, leader_id, color, rotation_position FROM team WHERE group_id = ? ORDER BY rotation_position, name`

	gIDBytes, _ := groupID.MarshalBinary()
	rows, err := r.db.QueryContext(ctx, query, gIDBytes)
	if err != nil {
		return nil, fmt.Errorf("FindByGroupID query error: %w", err)
	}
	defer rows.Close()

	var teams []*structure.Team
	for rows.Next() {
		var tID, gID, lID []byte
		var name string
		var color sql.NullString
		var rotationPosition int

		if err := rows.Scan(&tID, &name, &gID, &lID, &color, &rotationPosition); err != nil {
			return nil, fmt.Errorf("FindByGroupID scan error: %w", err)
		}

		teamID, _ := uuid.FromBytes(tID)
		grpID, _ := uuid.FromBytes(gID)
		var leaderID *uuid.UUID
		if len(lID) > 0 {
			uid, _ := uuid.FromBytes(lID)
			leaderID = &uid
		}
		teams = append(teams, structure.RestoreTeam(teamID, name, grpID, leaderID, color.String, rotationPosition))
	}
	return teams, nil
}

func (r *TeamRepository) Save(ctx context.Context, team *structure.Team) error {
	const query = `
		INSERT INTO team (id, name, group_id, leader_id, color, rotation_position)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			group_id = VALUES(group_id),
			leader_id = VALUES(leader_id),
			color = VALUES(color),
			rotation_position = VALUES(rotation_position)
	`

	teamIDBytes, _ := team.ID().MarshalBinary()
	groupIDBytes, _ := team.GroupID().MarshalBinary()
	var leaderID interface{} = nil
	if team.LeaderID() != nil {
		leaderIDBytes, _ := team.LeaderID().MarshalBinary()
		leaderID = leaderIDBytes
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		teamIDBytes,
		team.Name(),
		groupIDBytes,
		leaderID,
		team.Color(),
		team.RotationPosition(),
	)
	if err != nil {
		return fmt.Errorf("TeamRepository.Save: %w", err)
	}
	return nil
}

func (r *TeamRepository) UpdateRotationPositions(ctx context.Context, groupID uuid.UUID, orderedTeamIDs []uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("TeamRepository.UpdateRotationPositions begin: %w", err)
	}

	groupIDBytes, _ := groupID.MarshalBinary()
	shiftValue := len(orderedTeamIDs)

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE team SET rotation_position = rotation_position + ? WHERE group_id = ?`,
		shiftValue,
		groupIDBytes,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("TeamRepository.UpdateRotationPositions shift: %w", err)
	}

	for index, teamID := range orderedTeamIDs {
		teamIDBytes, _ := teamID.MarshalBinary()
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE team SET rotation_position = ? WHERE id = ? AND group_id = ?`,
			index+1,
			teamIDBytes,
			groupIDBytes,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("TeamRepository.UpdateRotationPositions apply: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("TeamRepository.UpdateRotationPositions commit: %w", err)
	}

	return nil
}

func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM team WHERE id = ?`
	idBytes, _ := id.MarshalBinary()
	_, err := r.db.ExecContext(ctx, query, idBytes)
	if err != nil {
		return fmt.Errorf("TeamRepository.Delete: %w", err)
	}
	return nil
}
