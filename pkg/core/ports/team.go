package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type TeamUseCase interface {
	GetTeamsResponseByGroup(ctx context.Context, groupID uuid.UUID) (dto.TeamListResponse, error)
	GetTeamDetails(ctx context.Context, id uuid.UUID) (*dto.TeamDetails, error)
	CreateResidentTeam(ctx context.Context, req dto.CreateResidentTeamRequest) (*dto.TeamDetails, error)
	UpdateResidentTeam(ctx context.Context, id uuid.UUID, req dto.UpdateResidentTeamRequest) (*dto.TeamDetails, error)
	DeleteTeam(ctx context.Context, id uuid.UUID) error
	GetTeamMemberOptionsResponse(ctx context.Context, groupID uuid.UUID, teamID *uuid.UUID) (dto.TeamMemberOptionsResponse, error)
	MoveUserToTeam(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
}
