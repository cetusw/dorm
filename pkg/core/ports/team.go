package ports

import (
	"context"

	"dorm/pkg/core/domain/structure"
)

type TeamUseCase interface {
	GetTeamsByDormitory(ctx context.Context, dormID int64) ([]*structure.Team, error)
}
