package query

import (
	"context"
	"database/sql"
	"dorm/pkg/core/domain/structure"
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
		SELECT t.id, t.group_id, t.leader_id, t.color, t.rotation_position
		FROM team t
		JOIN ` + "`group` g" + ` ON t.group_id = g.id
		WHERE g.dormitory_id = ? AND t.deleted_at IS NULL
		ORDER BY t.rotation_position, t.id
	`

	rows, err := q.db.QueryContext(ctx, query, dormID)
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
