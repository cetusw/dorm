package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"

	"dorm/pkg/core/domain/structure"
	"github.com/google/uuid"
)

type TeamUseCase interface {
	GetGroups(ctx context.Context) ([]*structure.Group, error)
	GetTeamsByDormitory(ctx context.Context, dormID int64) ([]*structure.Team, error)
	GetTeamByID(ctx context.Context, id uuid.UUID) (*dto.TeamListItem, error)
	GetTeamsList(ctx context.Context, dormID int64) ([]dto.TeamListItem, error)
	GetTeamsListByGroup(ctx context.Context, groupID uuid.UUID) ([]dto.TeamListItem, error)
	GetTeamsResponseByGroup(ctx context.Context, groupID uuid.UUID) (dto.TeamListResponse, error)
	GetTeamDetails(ctx context.Context, id uuid.UUID) (*dto.TeamDetails, error)
	CreateTeam(ctx context.Context, req dto.CreateTeamRequest) error
	CreateTeamInGroup(ctx context.Context, groupID uuid.UUID, req dto.CreateTeamRequest) error
	CreateResidentTeam(ctx context.Context, req dto.CreateResidentTeamRequest) (*dto.TeamDetails, error)
	UpdateTeam(ctx context.Context, id uuid.UUID, req dto.UpdateTeamRequest) error
	UpdateTeamInGroup(ctx context.Context, groupID uuid.UUID, id uuid.UUID, req dto.UpdateTeamRequest) error
	UpdateResidentTeam(ctx context.Context, id uuid.UUID, req dto.UpdateResidentTeamRequest) (*dto.TeamDetails, error)
	DeleteTeam(ctx context.Context, id uuid.UUID) error
	GetTeamMembersForEdit(ctx context.Context, teamID uuid.UUID) ([]dto.TeamMemberItem, error)
	GetTeamMembersForNewTeam(ctx context.Context, groupID uuid.UUID) ([]dto.TeamMemberItem, error)
	GetTeamMemberOptionsResponse(ctx context.Context, groupID uuid.UUID, teamID *uuid.UUID) (dto.TeamMemberOptionsResponse, error)
	MoveUserToTeam(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
	RemoveUserFromTeam(ctx context.Context, userID uuid.UUID) error
}
