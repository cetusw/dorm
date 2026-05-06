package ports

import (
	"context"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/ports/dto"
)

type TeamQueryService interface {
	FindTeamsByDormitoryID(ctx context.Context, dormID int64) ([]*structure.Team, error)
	GetTeamsDetailedList(ctx context.Context, dormID int64) ([]dto.TeamListItem, error)
}
