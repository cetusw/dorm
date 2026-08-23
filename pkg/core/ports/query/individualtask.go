package ports

import (
	"context"

	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type IndividualTaskQueryService interface {
	Get(context.Context, uuid.UUID) (*dto.IndividualTaskItem, error)
	ListMine(context.Context, uuid.UUID) ([]dto.IndividualTaskItem, error)
	ListResident(context.Context, uuid.UUID, []int64) ([]dto.IndividualTaskItem, error)
	ListReview(context.Context, []int64) ([]dto.IndividualTaskItem, error)
	SearchResidents(context.Context, []int64, string) ([]dto.IndividualTaskResidentOption, error)
	ListAreas(context.Context, int64) ([]dto.IndividualTaskArea, error)
}
