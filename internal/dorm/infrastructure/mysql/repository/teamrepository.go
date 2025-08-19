package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Store(team *model.Team) error {
	query := "INSERT INTO team (team_id, team_leader_id, color) VALUES (?, ?, ?)"
	_, err := r.db.Exec(query, team.TeamLeaderID, team.Color)
	if err != nil {
		return fmt.Errorf("failed to save team: %w", err)
	}
	return nil
}

func (r *TeamRepository) Find(teamID int) (*model.Team, error) {
	team := &model.Team{}
	query := "SELECT team_id, team_leader_id, color FROM team WHERE team_id = ?"

	err := r.db.QueryRow(query, teamID).Scan(&team.TeamID, &team.TeamLeaderID, &team.Color)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}

func (r *TeamRepository) FindAll() ([]model.Team, error) {
	query := "SELECT team_id, team_leader_id, color FROM team"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %w", err)
	}
	defer rows.Close()

	var teams []model.Team

	for rows.Next() {
		var team model.Team
		if err := rows.Scan(&team.TeamID, &team.TeamLeaderID, &team.Color); err != nil {
			return nil, fmt.Errorf("failed to scan team row: %w", err)
		}

		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team rows: %w", err)
	}

	return teams, nil
}
