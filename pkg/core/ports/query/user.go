package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"
)

type UserQueryService interface {
	GetUsersDetailedList(ctx context.Context) ([]dto.UserListItem, error)
}
