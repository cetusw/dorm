package repository

import (
	"database/sql"
	"dorm/pkg/dorm/application/model"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Find(teamID uuid.UUID) (*model.Team, error) {
	const sqlQuery = `
		SELECT id, group_id, leader_id, color, team_order
		FROM team 
		WHERE id = UUID_TO_BIN(?)`

	team := &model.Team{}
	err := r.db.QueryRow(sqlQuery, teamID).Scan(
		&team.ID,
		&team.GroupID,
		&team.TeamLeaderID,
		&team.Color,
		&team.Order,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}

func (r *TeamRepository) FindAll() ([]model.Team, error) {
	const sqlQuery = `
		SELECT id, group_id, leader_id, color, team_order 
		FROM team`

	rows, err := r.db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %w", err)
	}
	defer rows.Close()

	var teams []model.Team
	for rows.Next() {
		var team model.Team
		if err := rows.Scan(
			&team.ID,
			&team.GroupID,
			&team.TeamLeaderID,
			&team.Color,
			&team.Order,
		); err != nil {
			return nil, fmt.Errorf("failed to scan team row: %w", err)
		}

		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team rows: %w", err)
	}

	return teams, nil
}

func (r *TeamRepository) FindTeamByGroupIDAndOrder(groupID uuid.UUID, order int) (*model.Team, error) {
	const sqlQuery = `
		SELECT id, group_id, leader_id, color, team_order
		FROM team 
		WHERE group_id = UUID_TO_BIN(?) 
		  AND team_order = ?`

	team := &model.Team{}
	err := r.db.QueryRow(sqlQuery, groupID, order).Scan(
		&team.ID,
		&team.GroupID,
		&team.TeamLeaderID,
		&team.Color,
		&team.Order,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}

func (r *TeamRepository) FindTeamsByGroupID(groupID uuid.UUID) ([]model.Team, error) {
	const sqlQuery = `
		SELECT id, group_id, leader_id, color, team_order 
		FROM team 
		WHERE group_id = UUID_TO_BIN(?) 
		ORDER BY id`

	rows, err := r.db.Query(sqlQuery, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams by team group ID %s: %w", groupID, err)
	}
	defer rows.Close()

	var teams []model.Team
	for rows.Next() {
		var team model.Team
		if err := rows.Scan(
			&team.ID,
			&team.GroupID,
			&team.TeamLeaderID,
			&team.Color,
			&team.Order,
		); err != nil {
			return nil, fmt.Errorf("failed to scan team row for team group ID %s: %w", groupID, err)
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating teams by team group ID rows: %w", err)
	}

	return teams, nil
}

func (r *TeamRepository) FindGroupTeamByOrder(groupID uuid.UUID, order int) (*model.Team, error) {
	const sqlQuery = `
		SELECT id, group_id, leader_id, color, team_order
		FROM team 
		WHERE group_id = UUID_TO_BIN(?) AND team_order = ?
		ORDER BY team_order DESC 
		LIMIT 1`

	team := &model.Team{}
	err := r.db.QueryRow(sqlQuery, groupID, order).Scan(
		&team.ID,
		&team.GroupID,
		&team.TeamLeaderID,
		&team.Color,
		&team.Order,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get group team by order: %w", err)
	}
	return team, nil
}

func (r *TeamRepository) Store(team *model.Team) error {
	const sqlQuery = `
		INSERT INTO team (id, group_id, leader_id, color, team_order) 
		VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?)`

	_, err := r.db.Exec(sqlQuery,
		team.GroupID,
		team.TeamLeaderID,
		team.Color,
		team.Order,
	)
	if err != nil {
		return fmt.Errorf("failed to save team: %w", err)
	}
	return nil
}
