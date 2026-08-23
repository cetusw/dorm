package ports

import (
	"context"

	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type IndividualTaskUseCase interface {
	Create(context.Context, uuid.UUID, dto.IndividualTaskRequest) (*dto.IndividualTaskItem, error)
	Update(context.Context, uuid.UUID, uuid.UUID, dto.IndividualTaskRequest) (*dto.IndividualTaskItem, error)
	Delete(context.Context, uuid.UUID, uuid.UUID) error
	Complete(context.Context, uuid.UUID, uuid.UUID) (*dto.IndividualTaskItem, error)
	Open(context.Context, uuid.UUID, uuid.UUID) (*dto.IndividualTaskItem, error)
	Reject(context.Context, uuid.UUID, uuid.UUID) (*dto.IndividualTaskItem, error)
	Verify(context.Context, uuid.UUID, uuid.UUID) (*dto.IndividualTaskItem, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (*dto.IndividualTaskItem, error)
	ListMine(context.Context, uuid.UUID) (dto.IndividualTaskListResponse, error)
	ListResident(context.Context, uuid.UUID, uuid.UUID) (dto.IndividualTaskListResponse, error)
	ListReview(context.Context, uuid.UUID) (dto.IndividualTaskListResponse, error)
	SearchResidents(context.Context, uuid.UUID, string) (dto.IndividualTaskResidentsResponse, error)
	ListAreas(context.Context, uuid.UUID, int64) (dto.IndividualTaskAreasResponse, error)
}
