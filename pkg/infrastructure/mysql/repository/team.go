package repository

import (
	"bytes"
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
	const query = `SELECT id, group_id, leader_id, color, rotation_position FROM team WHERE id = ?`

	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var tID, gID, lID []byte
	var color sql.NullString
	var rotationPosition int

	if err := row.Scan(&tID, &gID, &lID, &color, &rotationPosition); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindTeamByID: %w", err)
	}

	teamID, _ := uuid.FromBytes(tID)
	groupID, _ := uuid.FromBytes(gID)
	leaderID, _ := uuid.FromBytes(lID)

	return structure.RestoreTeam(teamID, groupID, leaderID, color.String, rotationPosition), nil
}

// TODO: вынести в DTO, как в user
func (r *TeamRepository) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*structure.Team, error) {
	const query = `SELECT id, group_id, leader_id, color, rotation_position FROM team WHERE group_id = ? ORDER BY rotation_position, id`

	gIDBytes, _ := groupID.MarshalBinary()
	rows, err := r.db.QueryContext(ctx, query, gIDBytes)
	if err != nil {
		return nil, fmt.Errorf("FindByGroupID query error: %w", err)
	}
	defer rows.Close()

	var teams []*structure.Team
	for rows.Next() {
		var tID, gID, lID []byte
		var color sql.NullString
		var rotationPosition int

		if err := rows.Scan(&tID, &gID, &lID, &color, &rotationPosition); err != nil {
			return nil, fmt.Errorf("FindByGroupID scan error: %w", err)
		}

		teamID, _ := uuid.FromBytes(tID)
		grpID, _ := uuid.FromBytes(gID)
		leaderID, _ := uuid.FromBytes(lID)
		teams = append(teams, structure.RestoreTeam(teamID, grpID, leaderID, color.String, rotationPosition))
	}
	return teams, nil
}

func (r *TeamRepository) Save(ctx context.Context, team *structure.Team) error {
	const query = `
		INSERT INTO team (id, group_id, leader_id, color, rotation_position)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			group_id = VALUES(group_id),
			leader_id = VALUES(leader_id),
			color = VALUES(color),
			rotation_position = VALUES(rotation_position)
	`

	teamIDBytes, _ := team.ID().MarshalBinary()
	groupIDBytes, _ := team.GroupID().MarshalBinary()
	leaderID, _ := team.LeaderID().MarshalBinary()

	_, err := r.db.ExecContext(
		ctx,
		query,
		teamIDBytes,
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

// CreateWithLeader serializes rotation assignment and keeps the leader's membership
// in the new team in the same transaction.
func (r *TeamRepository) CreateWithLeader(ctx context.Context, team *structure.Team) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create team transaction: %w", err)
	}
	defer tx.Rollback()

	groupID, _ := team.GroupID().MarshalBinary()
	teamID, _ := team.ID().MarshalBinary()
	leaderID, _ := team.LeaderID().MarshalBinary()

	var lockedGroup []byte
	if err := tx.QueryRowContext(ctx, "SELECT id FROM `group` WHERE id = ? FOR UPDATE", groupID).Scan(&lockedGroup); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("group not found")
		}
		return fmt.Errorf("lock group: %w", err)
	}

	var currentTeamID []byte
	if err := tx.QueryRowContext(ctx, `SELECT team_id FROM user WHERE id = ? AND deleted_at IS NULL FOR UPDATE`, leaderID).Scan(&currentTeamID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("lock team leader: %w", err)
	}

	var leaderTeamID []byte
	err = tx.QueryRowContext(ctx, `SELECT id FROM team WHERE leader_id = ? FOR UPDATE`, leaderID).Scan(&leaderTeamID)
	if err == nil {
		return fmt.Errorf("resident is already a team leader")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check current team leader: %w", err)
	}

	var maxPosition int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(rotation_position), 0) FROM team WHERE group_id = ? FOR UPDATE`, groupID).Scan(&maxPosition); err != nil {
		return fmt.Errorf("load rotation position: %w", err)
	}
	position := maxPosition + 1
	if _, err := tx.ExecContext(ctx, `INSERT INTO team (id, group_id, leader_id, color, rotation_position) VALUES (?, ?, ?, ?, ?)`, teamID, groupID, leaderID, team.Color(), position); err != nil {
		return fmt.Errorf("create team: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE user SET team_id = ? WHERE id = ?`, teamID, leaderID); err != nil {
		return fmt.Errorf("move team leader: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create team transaction: %w", err)
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

func (r *TeamRepository) ReplaceLeaderAndRemoveMember(ctx context.Context, teamID, leaderID, replacementLeaderID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace team leader transaction: %w", err)
	}
	defer tx.Rollback()

	teamIDBytes, err := teamID.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal team id: %w", err)
	}
	leaderIDBytes, err := leaderID.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal leader id: %w", err)
	}
	replacementIDBytes, err := replacementLeaderID.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal replacement leader id: %w", err)
	}
	if leaderID == replacementLeaderID {
		return structure.ErrInvalidReplacementLeader
	}

	var currentLeaderID []byte
	if err := tx.QueryRowContext(ctx, `SELECT leader_id FROM team WHERE id = ? FOR UPDATE`, teamIDBytes).Scan(&currentLeaderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("team not found")
		}
		return fmt.Errorf("lock team: %w", err)
	}
	if len(currentLeaderID) == 0 || !bytes.Equal(currentLeaderID, leaderIDBytes) {
		return structure.ErrInvalidReplacementLeader
	}

	rows, err := tx.QueryContext(ctx, `SELECT id FROM user WHERE team_id = ? AND deleted_at IS NULL FOR UPDATE`, teamIDBytes)
	if err != nil {
		return fmt.Errorf("lock team members: %w", err)
	}
	memberCount := 0
	for rows.Next() {
		memberCount++
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close locked team members: %w", err)
	}
	if memberCount <= 1 {
		return structure.ErrCannotRemoveOnlyLeader
	}

	for _, userID := range [][]byte{leaderIDBytes, replacementIDBytes} {
		var memberID []byte
		if err := tx.QueryRowContext(ctx, `SELECT id FROM user WHERE id = ? AND team_id = ? AND deleted_at IS NULL FOR UPDATE`, userID, teamIDBytes).Scan(&memberID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return structure.ErrInvalidReplacementLeader
			}
			return fmt.Errorf("lock team member: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE team SET leader_id = ? WHERE id = ?`, replacementIDBytes, teamIDBytes); err != nil {
		return fmt.Errorf("replace team leader: %w", err)
	}
	result, err := tx.ExecContext(ctx, `UPDATE user SET team_id = NULL WHERE id = ? AND team_id = ?`, leaderIDBytes, teamIDBytes)
	if err != nil {
		return fmt.Errorf("remove former team leader: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count removed former team leader: %w", err)
	}
	if affected != 1 {
		return structure.ErrInvalidReplacementLeader
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit replace team leader transaction: %w", err)
	}
	return nil
}
