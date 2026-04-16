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
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*user.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*dto.ProfileViewModel, error)
	GetUsersList(ctx context.Context) ([]dto.UserListItem, error)
}
