package ports

import (
	"context"
	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type TaskCatalogUseCase interface {
	ListTasks(ctx context.Context) ([]dto.TaskCatalogItem, error)
	ListAreas(ctx context.Context) ([]*catalog.Area, error)
	GetTask(ctx context.Context, id uuid.UUID) (*dto.TaskCatalogItem, error)
	CreateTask(ctx context.Context, req dto.UpsertTaskCatalogRequest) error
	UpdateTask(ctx context.Context, id uuid.UUID, req dto.UpsertTaskCatalogRequest) error
	DeleteTask(ctx context.Context, id uuid.UUID) error
}

