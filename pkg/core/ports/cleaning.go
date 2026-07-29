package ports

import (
	"context"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/ports/dto"
	"time"

	"github.com/google/uuid"
)

type CleaningUseCase interface {
	AssignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	UnassignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	OpenTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	CancelCompletion(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	VerifyTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	ReopenTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	StartNewWeek(ctx context.Context) error
	StartNewDutyForGroup(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, startDate, endDate time.Time) error
	StartNewDutiesForDormitory(ctx context.Context, dormitoryID int64, startDate, endDate time.Time) error

	IsUserOnDuty(ctx context.Context, userID uuid.UUID) (bool, error)

	// TODO: возможно, стоит переписать telegram на использование одного метода GetTeamTasks и внутри фильтровать так, как ему нужно
	GetTeamTasks(ctx context.Context, teamID uuid.UUID) ([]dto.TaskViewModel, error)
	GetLatestDuties(ctx context.Context) ([]dto.DutyViewModel, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*duty.UserStats, error)
	GetTaskCandidates(ctx context.Context, userID uuid.UUID) ([]dto.TaskViewModel, error)
	GetAllAssignedTasks(ctx context.Context, userID uuid.UUID) ([]dto.TaskViewModel, error)
	GetUncompletedAssignedTasks(ctx context.Context, userID uuid.UUID) ([]dto.TaskViewModel, error)
}
