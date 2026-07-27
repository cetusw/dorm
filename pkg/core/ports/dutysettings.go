package ports

import (
	"context"

	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type DutySettingsUseCase interface {
	GetDutySettings(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID) (*dto.DutySettingsResponse, error)
	GetTeamDetails(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamID uuid.UUID) (*dto.TeamDetails, error)
	GetTeamMemberOptions(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamID *uuid.UUID) (dto.TeamMemberOptionsResponse, error)
	CreateTeam(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, req dto.CreateResidentTeamRequest) (*dto.TeamDetails, error)
	UpdateTeam(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamID uuid.UUID, req dto.UpdateResidentTeamRequest) (*dto.TeamDetails, error)
	DeleteTeam(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamID uuid.UUID) error
	ReorderTeams(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamIDs []uuid.UUID) error
	CreateArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, req dto.CreateAreaRequest) (*dto.AreaDetails, error)
	UpdateArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int, req dto.UpdateAreaRequest) (*dto.AreaDetails, error)
	DeleteArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int) error
	CreateTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int, req dto.CreateTaskRequest) (*dto.TaskDetails, error)
	UpdateTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, taskID uuid.UUID, req dto.UpdateTaskRequest) (*dto.TaskDetails, error)
	DeleteTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, taskID uuid.UUID) error
}
