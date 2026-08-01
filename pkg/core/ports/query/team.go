package ports

import (
	"context"
	"dorm/pkg/core/domain/structure"
)

type TeamQueryService interface {
	FindTeamsByDormitoryID(ctx context.Context, dormID int64) ([]*structure.Team, error)
}
