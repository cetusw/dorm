package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type DormitoryUseCase interface {
	GetDormitoriesResponse(ctx context.Context) (dto.DormitoryListResponse, error)
	GetDormitoryDetails(ctx context.Context, id int64) (*dto.DormitoryDetails, error)
	CanManageDormitories(ctx context.Context, userID uuid.UUID) (bool, error)
	GetUserOptionsResponse(ctx context.Context) (dto.UserOptionsResponse, error)
	CreateDormitoryDetails(ctx context.Context, req dto.CreateDormitoryRequest) (*dto.DormitoryDetails, error)
	UpdateDormitoryDetails(ctx context.Context, id int64, req dto.UpdateDormitoryRequest) (*dto.DormitoryDetails, error)
	DeleteDormitory(ctx context.Context, id int64) error
	GetGroupsResponse(ctx context.Context, dormitoryID int64) (dto.GroupListResponse, error)
	GetGroupDetails(ctx context.Context, id uuid.UUID) (*dto.GroupDetails, error)
	GetDormitoryUserOptionsResponse(ctx context.Context, dormitoryID int64) (dto.UserOptionsResponse, error)
	CreateGroupDetails(ctx context.Context, req dto.CreateGroupRequest) (*dto.GroupDetails, error)
	UpdateGroupDetails(ctx context.Context, id uuid.UUID, req dto.UpdateGroupRequest) (*dto.GroupDetails, error)
	DeleteGroup(ctx context.Context, id uuid.UUID) error
}
