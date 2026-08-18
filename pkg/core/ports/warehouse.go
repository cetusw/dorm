package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type WarehouseUseCase interface {
	GetWarehouse(ctx context.Context, currentUserID uuid.UUID) (*dto.WarehouseResponse, error)
	GetItemHistory(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID) (*dto.WarehouseItemHistoryResponse, error)
	CreateItem(ctx context.Context, currentUserID uuid.UUID, request dto.CreateWarehouseItemRequest) (*dto.WarehouseItem, error)
	UpdateItem(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID, request dto.UpdateWarehouseItemRequest) (*dto.WarehouseItem, error)
	DeleteItem(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID) error
	AddItems(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID, request dto.WarehouseMovementRequest) (*dto.WarehouseItem, error)
	WriteOffItems(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID, request dto.WarehouseMovementRequest) (*dto.WarehouseItem, error)
	UpdateMovement(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID, movementID uuid.UUID, request dto.UpdateWarehouseMovementRequest) (*dto.WarehouseItem, error)
	DeleteMovement(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID, movementID uuid.UUID) (*dto.WarehouseItem, error)
}
