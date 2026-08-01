package ports

import (
	"context"
	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type TaskCatalogUseCase interface {
	ListAreasByGroup(ctx context.Context, groupID uuid.UUID) ([]*catalog.Area, error)
	ListAreasByDormitory(ctx context.Context, dormitoryID int64) ([]*catalog.Area, error)
	GetAreasResponse(ctx context.Context, dormitoryID int64) (dto.AreaListResponse, error)
	GetAreaDetails(ctx context.Context, id int) (*dto.AreaDetails, error)
	GetTasksResponse(ctx context.Context, dormitoryID int64) (dto.TaskListResponse, error)
	GetTaskDetails(ctx context.Context, dormitoryID int64, id uuid.UUID) (*dto.TaskDetails, error)
	CreateArea(ctx context.Context, dormitoryID int64, req dto.CreateAreaRequest) (*dto.AreaDetails, error)
	UpdateArea(ctx context.Context, dormitoryID int64, id int, req dto.UpdateAreaRequest) (*dto.AreaDetails, error)
	DeleteArea(ctx context.Context, dormitoryID int64, id int) error
	CreateTaskDetails(ctx context.Context, dormitoryID int64, req dto.CreateTaskRequest) (*dto.TaskDetails, error)
	UpdateTaskDetails(ctx context.Context, dormitoryID int64, id uuid.UUID, req dto.UpdateTaskRequest) (*dto.TaskDetails, error)
	DeleteTaskDetails(ctx context.Context, dormitoryID int64, id uuid.UUID) error
}
