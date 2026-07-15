package ports

import (
	"context"
	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type TaskCatalogUseCase interface {
	ListTaskGroupsByGroup(ctx context.Context, groupID uuid.UUID) ([]dto.TaskCatalogGroup, error)
	ListAreas(ctx context.Context) ([]*catalog.Area, error)
	ListAreasByGroup(ctx context.Context, groupID uuid.UUID) ([]*catalog.Area, error)
	ListAreasByDormitory(ctx context.Context, dormitoryID int64) ([]*catalog.Area, error)
	GetAreasResponse(ctx context.Context, dormitoryID int64) (dto.AreaListResponse, error)
	GetAreaDetails(ctx context.Context, id int) (*dto.AreaDetails, error)
	ListCommonTaskGroups(ctx context.Context) ([]dto.TaskCatalogGroup, error)
	GetTask(ctx context.Context, id uuid.UUID) (*dto.TaskCatalogItem, error)
	CreateArea(ctx context.Context, dormitoryID int64, req dto.CreateAreaRequest) (*dto.AreaDetails, error)
	UpdateArea(ctx context.Context, dormitoryID int64, id int, req dto.UpdateAreaRequest) (*dto.AreaDetails, error)
	DeleteArea(ctx context.Context, dormitoryID int64, id int) error
	CreateTask(ctx context.Context, req dto.UpsertTaskCatalogRequest) error
	CreateTaskInGroup(ctx context.Context, groupID uuid.UUID, req dto.UpsertTaskCatalogRequest) error
	CreateCommonTask(ctx context.Context, req dto.UpsertTaskCatalogRequest) error
	UpdateTask(ctx context.Context, id uuid.UUID, req dto.UpsertTaskCatalogRequest) error
	UpdateTaskInGroup(ctx context.Context, groupID uuid.UUID, id uuid.UUID, req dto.UpsertTaskCatalogRequest) error
	UpdateCommonTask(ctx context.Context, id uuid.UUID, req dto.UpsertTaskCatalogRequest) error
	DeleteTask(ctx context.Context, id uuid.UUID) error
	DeleteTaskInGroup(ctx context.Context, groupID uuid.UUID, id uuid.UUID) error
	DeleteCommonTask(ctx context.Context, id uuid.UUID) error
}
