package ports

import (
	"context"

	"dorm/pkg/core/domain/structure"
)

type DormitoryUseCase interface {
	GetDormitories(ctx context.Context) ([]*structure.Dormitory, error)
}
