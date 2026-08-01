package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"
)

type UserQueryService interface {
	GetResidentsByDormitoryID(ctx context.Context, dormitoryID int64) ([]dto.ResidentListItem, error)
}
