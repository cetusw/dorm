package team

import (
	"context"
	ports "dorm/pkg/core/ports/query"

	"dorm/pkg/core/domain/structure"
)

type Service struct {
	queryService ports.TeamQueryService
}

func NewTeamService(
	queryService ports.TeamQueryService,
) *Service {
	return &Service{
		queryService: queryService,
	}
}

func (s *Service) GetTeamsByDormitory(ctx context.Context, dormID int64) ([]*structure.Team, error) {
	return s.queryService.FindTeamsByDormitoryID(ctx, dormID)
}
