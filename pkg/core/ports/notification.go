package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type NotificationUseCase interface {
	Subscribe(ctx context.Context, userID uuid.UUID, request dto.SubscribeToPushRequest) error
	Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) error
}
