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
	const query = `SELECT id, group_id, leader_id, color, team_order FROM team WHERE id = ?`

	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var tID, gID, lID []byte
	var color string
	var order int

	if err := row.Scan(&tID, &gID, &lID, &color, &order); err != nil {
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

	return structure.RestoreTeam(teamID, groupID, leaderID, color, order), nil
}

func (r *TeamRepository) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*structure.Team, error) {
	const query = `SELECT id, group_id, leader_id, color, team_order FROM team WHERE group_id = ?`

	gIDBytes, _ := groupID.MarshalBinary()
	rows, err := r.db.QueryContext(ctx, query, gIDBytes)
	if err != nil {
		return nil, fmt.Errorf("FindByGroupID query error: %w", err)
	}
	defer rows.Close()

	var teams []*structure.Team
	for rows.Next() {
		var tID, gID, lID []byte
		var color string
		var order int

		if err := rows.Scan(&tID, &gID, &lID, &color, &order); err != nil {
			return nil, fmt.Errorf("FindByGroupID scan error: %w", err)
		}

		teamID, _ := uuid.FromBytes(tID)
		grpID, _ := uuid.FromBytes(gID)
		var leaderID *uuid.UUID
		if len(lID) > 0 {
			uid, _ := uuid.FromBytes(lID)
			leaderID = &uid
		}
		teams = append(teams, structure.RestoreTeam(teamID, grpID, leaderID, color, order))
	}
	return teams, nil
}
