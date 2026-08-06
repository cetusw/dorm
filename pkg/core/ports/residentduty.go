package ports

import (
	"context"

	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type ResidentDutyUseCase interface {
	GetCurrentDuty(ctx context.Context, userID uuid.UUID, groupID *uuid.UUID, dormitoryID *int64) (*dto.ResidentCurrentDutyResponse, error)
	GetDutyHistory(ctx context.Context, userID uuid.UUID, groupID *uuid.UUID) (*dto.ResidentDutyHistoryResponse, error)
	GetDutyDetails(ctx context.Context, userID uuid.UUID, dutyID uuid.UUID) (*dto.ResidentDutyDetailsResponse, error)
	TakeTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error
	ReturnTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error
	CompleteTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error
	OpenTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error
	VerifyTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error
	ReopenTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error
}
