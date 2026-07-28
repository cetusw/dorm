package ports

import (
	"context"
	domain "dorm/pkg/core/domain/notification"
	"errors"

	"github.com/google/uuid"
)

var ErrPushSubscriptionExpired = errors.New("push subscription expired")

type NotifyUserCommand struct {
	UserID           uuid.UUID
	Type             domain.NotificationType
	Title            string
	Body             string
	TargetURL        string
	DeduplicationKey string
}

type UserNotificationUseCase interface {
	NotifyUser(ctx context.Context, command NotifyUserCommand) error
}

type NotificationRepository interface {
	CreateIfAbsent(ctx context.Context, notification *domain.Notification) (bool, error)
}

type PushMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
	Tag   string `json:"tag,omitempty"`
	Icon  string `json:"icon,omitempty"`
	Badge string `json:"badge,omitempty"`
}

type PushSender interface {
	Send(ctx context.Context, subscription *domain.PushSubscription, message PushMessage) error
}
