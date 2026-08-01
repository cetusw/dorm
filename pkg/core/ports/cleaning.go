package ports

import (
	"context"
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
}
