package ports

import (
	"context"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/duty"
)

type CleaningUseCase interface {
	AssignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	UnassignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error

	StartNewWeek(ctx context.Context) error

	GetTeamActiveDuty(ctx context.Context, teamID uuid.UUID) (*duty.Duty, error)
}
