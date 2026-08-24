package repository

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	const query = `SELECT id, group_id, leader_id, color, rotation_position FROM team WHERE id = ? AND deleted_at IS NULL`

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
	const query = `SELECT id, group_id, leader_id, color, rotation_position FROM team WHERE group_id = ? AND deleted_at IS NULL ORDER BY rotation_position, id`

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
	err = tx.QueryRowContext(ctx, `SELECT id FROM team WHERE leader_id = ? AND deleted_at IS NULL FOR UPDATE`, leaderID).Scan(&leaderTeamID)
	if err == nil {
		return fmt.Errorf("resident is already a team leader")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check current team leader: %w", err)
	}

	var maxPosition int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(rotation_position), 0) FROM team WHERE group_id = ? AND deleted_at IS NULL FOR UPDATE`, groupID).Scan(&maxPosition); err != nil {
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
	var lockedGroupID []byte
	if err := tx.QueryRowContext(ctx, "SELECT id FROM `group` WHERE id = ? FOR UPDATE", groupIDBytes).Scan(&lockedGroupID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("TeamRepository.UpdateRotationPositions lock group: %w", err)
	}
	shiftValue := len(orderedTeamIDs)

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE team SET rotation_position = rotation_position + ? WHERE group_id = ? AND deleted_at IS NULL`,
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
			`UPDATE team SET rotation_position = ? WHERE id = ? AND group_id = ? AND deleted_at IS NULL`,
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

func (r *TeamRepository) SoftDelete(ctx context.Context, id uuid.UUID, at time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin soft delete team transaction: %w", err)
	}
	defer tx.Rollback()

	idBytes, _ := id.MarshalBinary()
	var groupID []byte
	if err := tx.QueryRowContext(ctx, `SELECT group_id FROM team WHERE id = ? AND deleted_at IS NULL`, idBytes).Scan(&groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("team not found")
		}
		return fmt.Errorf("lock team: %w", err)
	}
	// Every rotation mutation locks this group row first. Lock it before checking
	// the calendar period so creation, reorder and deletion cannot interleave.
	var lockedGroupID []byte
	if err := tx.QueryRowContext(ctx, "SELECT id FROM `group` WHERE id = ? FOR UPDATE", groupID).Scan(&lockedGroupID); err != nil {
		return fmt.Errorf("lock team group: %w", err)
	}
	var lockedTeamID []byte
	if err := tx.QueryRowContext(ctx, `SELECT id FROM team WHERE id = ? AND group_id = ? AND deleted_at IS NULL FOR UPDATE`, idBytes, groupID).Scan(&lockedTeamID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("team not found")
		}
		return fmt.Errorf("lock active team: %w", err)
	}

	var currentDutyID []byte
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM duty
		WHERE team_id = ? AND start_date <= ? AND end_date > ?
		LIMIT 1 FOR UPDATE`, idBytes, at, at).Scan(&currentDutyID)
	if err == nil {
		return structure.ErrCannotDeleteCurrentDutyTeam
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check current duty: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE team SET deleted_at = ?, rotation_position = NULL WHERE id = ?`, at, idBytes); err != nil {
		return fmt.Errorf("soft delete team: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE user SET team_id = NULL WHERE team_id = ?`, idBytes); err != nil {
		return fmt.Errorf("clear deleted team members: %w", err)
	}

	var activeCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM team WHERE group_id = ? AND deleted_at IS NULL`, groupID).Scan(&activeCount); err != nil {
		return fmt.Errorf("count active teams: %w", err)
	}
	if activeCount > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE team SET rotation_position = rotation_position + ? WHERE group_id = ? AND deleted_at IS NULL`, activeCount, groupID); err != nil {
			return fmt.Errorf("shift active team rotation: %w", err)
		}
		rows, err := tx.QueryContext(ctx, `SELECT id FROM team WHERE group_id = ? AND deleted_at IS NULL ORDER BY rotation_position, id`, groupID)
		if err != nil {
			return fmt.Errorf("load active teams for rotation: %w", err)
		}
		teamIDs := make([][]byte, 0, activeCount)
		for rows.Next() {
			var teamID []byte
			if err := rows.Scan(&teamID); err != nil {
				_ = rows.Close()
				return fmt.Errorf("scan active team for rotation: %w", err)
			}
			teamIDs = append(teamIDs, teamID)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("close active team rotation rows: %w", err)
		}
		for position, teamID := range teamIDs {
			if _, err := tx.ExecContext(ctx, `UPDATE team SET rotation_position = ? WHERE id = ?`, position+1, teamID); err != nil {
				return fmt.Errorf("normalize active team rotation: %w", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit soft delete team transaction: %w", err)
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
	if err := tx.QueryRowContext(ctx, `SELECT leader_id FROM team WHERE id = ? AND deleted_at IS NULL FOR UPDATE`, teamIDBytes).Scan(&currentLeaderID); err != nil {
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
