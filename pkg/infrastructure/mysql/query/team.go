package query

import (
	"context"
	"database/sql"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/ports/dto"
	"fmt"

	"github.com/google/uuid"
)

type TeamQueryService struct {
	db *sql.DB
}

func NewTeamQueryService(db *sql.DB) *TeamQueryService {
	return &TeamQueryService{db: db}
}

func (q *TeamQueryService) FindTeamsByDormitoryID(ctx context.Context, dormID int64) ([]*structure.Team, error) {
	const query = `
		SELECT t.id, t.name, t.group_id, t.leader_id, t.color, t.team_order 
		FROM team t
		JOIN ` + "`group` g" + ` ON t.group_id = g.id
		WHERE g.dormitory_id = ?
	`

	rows, err := q.db.QueryContext(ctx, query, dormID)
	if err != nil {
		return nil, fmt.Errorf("FindByGroupID query error: %w", err)
	}
	defer rows.Close()

	var teams []*structure.Team
	for rows.Next() {
		var tID, gID, lID []byte
		var name string
		var color string
		var order int

		if err := rows.Scan(&tID, &name, &gID, &lID, &color, &order); err != nil {
			return nil, fmt.Errorf("FindByGroupID scan error: %w", err)
		}

		teamID, _ := uuid.FromBytes(tID)
		grpID, _ := uuid.FromBytes(gID)
		var leaderID *uuid.UUID
		if len(lID) > 0 {
			uid, _ := uuid.FromBytes(lID)
			leaderID = &uid
		}
		teams = append(teams, structure.RestoreTeam(teamID, name, grpID, leaderID, color, order))
	}
	return teams, nil
}

func (q *TeamQueryService) GetTeamsDetailedList(ctx context.Context, dormID int64) ([]dto.TeamListItem, error) {
	const query = `
		SELECT
			t.id,
			t.name,
			t.color,
			t.team_order,
			t.group_id,
			g.name,
			g.dormitory_id,
			d.name,
			COUNT(u.id)
		FROM team t
		JOIN ` + "`group` g" + ` ON t.group_id = g.id
		JOIN dormitory d ON g.dormitory_id = d.id
		LEFT JOIN user u ON u.team_id = t.id AND u.deleted_at IS NULL
		WHERE g.dormitory_id = ?
		GROUP BY t.id, t.name, t.color, t.team_order, t.group_id, g.name, g.dormitory_id, d.name
	`

	rows, err := q.db.QueryContext(ctx, query, dormID)
	if err != nil {
		return nil, fmt.Errorf("GetTeamsDetailedList query error: %w", err)
	}
	defer rows.Close()

	var teams []dto.TeamListItem
	for rows.Next() {
		var item dto.TeamListItem
		var teamIDBytes, groupIDBytes []byte
		if err := rows.Scan(
			&teamIDBytes,
			&item.Name,
			&item.Color,
			&item.Order,
			&groupIDBytes,
			&item.GroupName,
			&item.DormitoryID,
			&item.DormitoryName,
			&item.MembersCount,
		); err != nil {
			return nil, fmt.Errorf("GetTeamsDetailedList scan error: %w", err)
		}

		teamID, _ := uuid.FromBytes(teamIDBytes)
		groupID, _ := uuid.FromBytes(groupIDBytes)
		item.ID = teamID
		item.GroupID = groupID
		teams = append(teams, item)
	}
	return teams, rows.Err()
}
