package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/user"
)

type UserUseCase interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) error
	UpdateUser(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) error
	SoftDeleteUser(ctx context.Context, id uuid.UUID) error
	RegisterUser(ctx context.Context, telegramID int64, fullName string) error
	AuthenticateResident(ctx context.Context, login string, password string) (*user.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*user.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.CurrentUserResponse, error)
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*dto.ProfileViewModel, error)
	GetUsersList(ctx context.Context) ([]dto.UserListItem, error)
	GetResidentsResponse(ctx context.Context, dormitoryID int64) (dto.ResidentListResponse, error)
	GetResidentDetails(ctx context.Context, id uuid.UUID) (*dto.ResidentDetails, error)
	CreateResident(ctx context.Context, req dto.CreateResidentRequest) (*dto.ResidentDetails, error)
	UpdateResident(ctx context.Context, id uuid.UUID, req dto.UpdateResidentRequest) (*dto.ResidentDetails, error)
	DeleteResident(ctx context.Context, id uuid.UUID) error
}
