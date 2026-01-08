package ports

import (
	"context"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type CleaningUseCase interface {
	AssignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	UnassignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	OpenTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	StartNewWeek(ctx context.Context) error

	IsUserOnDuty(ctx context.Context, userID uuid.UUID) (bool, error)

	// TODO: возможно, стоит переписать telegram на использование одного метода GetTeamTasks и внутри фильтровать так, как ему нужно
	GetTeamTasks(ctx context.Context, teamID uuid.UUID) ([]dto.TaskViewModel, error)
	GetLatestDuties(ctx context.Context) ([]dto.DutyViewModel, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*duty.UserStats, error)
	GetTaskCandidates(ctx context.Context, userID uuid.UUID) ([]dto.TaskViewModel, error)
	GetAllAssignedTasks(ctx context.Context, userID uuid.UUID) ([]dto.TaskViewModel, error)
	GetUncompletedAssignedTasks(ctx context.Context, userID uuid.UUID) ([]dto.TaskViewModel, error)
}
