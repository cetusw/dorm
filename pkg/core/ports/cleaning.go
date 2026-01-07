package ports

import (
	"context"
	"dorm/pkg/core/domain/duty"

	"github.com/google/uuid"
)

type CleaningUseCase interface {
	AssignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	UnassignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	OpenTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	StartNewWeek(ctx context.Context) error

	IsUserOnDuty(ctx context.Context, userID uuid.UUID) (bool, error)

	GetTeamActiveDuty(ctx context.Context, teamID uuid.UUID) (*duty.Duty, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*duty.UserStats, error)
	GetTaskCandidates(ctx context.Context, userID uuid.UUID) ([]TaskViewModel, error)
	GetAllAssignedTasks(ctx context.Context, userID uuid.UUID) ([]TaskViewModel, error)
	GetUncompletedAssignedTasks(ctx context.Context, userID uuid.UUID) ([]TaskViewModel, error)
}
