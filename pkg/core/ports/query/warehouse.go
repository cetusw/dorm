package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type WarehouseQueryService interface {
	ListItems(ctx context.Context, dormitoryID int64) ([]dto.WarehouseItem, error)
	GetItemHistory(ctx context.Context, dormitoryID int64, itemID uuid.UUID) (*dto.WarehouseItemHistoryResponse, error)
}
