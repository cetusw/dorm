package notification

import (
	"context"
	"fmt"
	"log"
	"time"

	domain "dorm/pkg/core/domain/notification"
	"dorm/pkg/core/ports"

	"github.com/google/uuid"
)

type DeliveryService struct {
	notifications ports.NotificationRepository
	subscriptions domain.PushSubscriptionRepository
	pushSender    ports.PushSender
	pushEnabled   bool
	now           func() time.Time
	newID         func() uuid.UUID
}

func NewDeliveryService(
	notifications ports.NotificationRepository,
	subscriptions domain.PushSubscriptionRepository,
	pushSender ports.PushSender,
	pushEnabled bool,
) *DeliveryService {
	return &DeliveryService{
		notifications: notifications,
		subscriptions: subscriptions,
		pushSender:    pushSender,
		pushEnabled:   pushEnabled,
		now:           time.Now,
		newID:         uuid.New,
	}
}

func (s *DeliveryService) NotifyUser(ctx context.Context, command ports.NotifyUserCommand) error {
	notification, err := domain.NewNotification(
		s.newID(),
		command.UserID,
		command.Type,
		command.Title,
		command.Body,
		command.TargetURL,
		command.DeduplicationKey,
		s.now(),
	)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}

	created, err := s.notifications.CreateIfAbsent(ctx, notification)
	if err != nil {
		return fmt.Errorf("store notification: %w", err)
	}
	if !created {
		return nil
	}
	if !s.pushEnabled {
		return nil
	}

	subscriptions, err := s.subscriptions.FindByUserID(ctx, notification.UserID())
	if err != nil {
		return fmt.Errorf("load push subscriptions: %w", err)
	}
	if len(subscriptions) == 0 {
		return nil
	}

	message := ports.PushMessage{
		Title: notification.Title(),
		Body:  notification.Body(),
		URL:   notification.TargetURL(),
		Tag:   notification.DeduplicationKey(),
		Icon:  "/app/icons/app-icon-192.png",
		Badge: "/app/icons/app-icon-192.png",
	}

	for _, subscription := range subscriptions {
		err := s.pushSender.Send(ctx, subscription, message)
		if err == nil {
			continue
		}

		if err == ports.ErrPushSubscriptionExpired {
			if deleteErr := s.subscriptions.DeleteByID(ctx, subscription.ID()); deleteErr != nil {
				log.Printf("delete expired push subscription %s: %v", subscription.ID(), deleteErr)
			}
			continue
		}

		log.Printf("send push notification to subscription %s: %v", subscription.ID(), err)
	}

	return nil
}
