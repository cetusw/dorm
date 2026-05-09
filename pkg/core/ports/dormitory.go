package ports

import (
	"context"

	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type DormitoryUseCase interface {
	GetDormitories(ctx context.Context) ([]*structure.Dormitory, error)
	GetDormitoryByID(ctx context.Context, id int64) (*structure.Dormitory, error)
	GetDormitoriesList(ctx context.Context) ([]dto.DormitoryListItem, error)
	GetUserOptions(ctx context.Context) ([]dto.UserOption, error)
	CreateDormitory(ctx context.Context, req dto.UpsertDormitoryRequest) error
	UpdateDormitory(ctx context.Context, id int64, req dto.UpsertDormitoryRequest) error
	DeleteDormitory(ctx context.Context, id int64) error
	GetGroupsList(ctx context.Context, dormitoryID int64) ([]dto.GroupListItem, error)
	GetGroupByID(ctx context.Context, id uuid.UUID) (*structure.Group, error)
	GetDormitoryUserOptions(ctx context.Context, dormitoryID int64) ([]dto.UserOption, error)
	CreateGroup(ctx context.Context, dormitoryID int64, req dto.UpsertGroupRequest) error
	UpdateGroup(ctx context.Context, id uuid.UUID, req dto.UpsertGroupRequest) error
	DeleteGroup(ctx context.Context, id uuid.UUID) error
}
