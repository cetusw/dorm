package ports

import (
	"context"

	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type DutyUseCase interface {
	GetGroupDuties(ctx context.Context, groupID uuid.UUID) ([]dto.DutyListItem, error)
	GetDutyDetail(ctx context.Context, groupID uuid.UUID, dutyID uuid.UUID) (*dto.DutyDetailItem, error)
	GetFutureDutyTasks(ctx context.Context, groupID uuid.UUID) ([]dto.FutureDutyTaskGroup, error)
	GetCommonFutureDutyTasks(ctx context.Context) ([]dto.FutureDutyTaskGroup, error)
	UpdateGroupDutySettings(ctx context.Context, groupID uuid.UUID, nextDutyTeam *int, activeTaskIDs []uuid.UUID) error
}
