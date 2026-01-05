package ports

import (
	"context"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/user"
)

type UserUseCase interface {
	RegisterUser(ctx context.Context, telegramID int64, fullName string) error
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*user.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}
