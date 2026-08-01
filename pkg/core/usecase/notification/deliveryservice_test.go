package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "dorm/pkg/core/domain/notification"
	"dorm/pkg/core/ports"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type notificationRepositoryStub struct {
	createCalls int
	created     bool
	err         error
	last        *domain.Notification
}

func (s *notificationRepositoryStub) CreateIfAbsent(_ context.Context, notification *domain.Notification) (bool, error) {
	s.createCalls++
	s.last = notification
	return s.created, s.err
}

type pushSenderStub struct {
	calls             int
	errBySubscription map[uuid.UUID]error
}

func (s *pushSenderStub) Send(_ context.Context, subscription *domain.PushSubscription, _ ports.PushMessage) error {
	s.calls++
	if s.errBySubscription == nil {
		return nil
	}
	return s.errBySubscription[subscription.ID()]
}

type pushSubscriptionsStub struct {
	subscriptions []*domain.PushSubscription
	findErr       error
	deletedIDs    []uuid.UUID
}

func (s *pushSubscriptionsStub) Save(context.Context, *domain.PushSubscription) error { return nil }
func (s *pushSubscriptionsStub) FindByEndpointHash(context.Context, []byte) (*domain.PushSubscription, error) {
	return nil, nil
}
func (s *pushSubscriptionsStub) FindByUserID(context.Context, uuid.UUID) ([]*domain.PushSubscription, error) {
	return s.subscriptions, s.findErr
}
func (s *pushSubscriptionsStub) DeleteByID(_ context.Context, id uuid.UUID) error {
	s.deletedIDs = append(s.deletedIDs, id)
	return nil
}
func (s *pushSubscriptionsStub) DeleteByEndpointHash(context.Context, uuid.UUID, []byte) error {
	return nil
}

func TestDeliveryServiceNotifyUserCreatesAndSends(t *testing.T) {
	repo := &notificationRepositoryStub{created: true}
	subscription, _ := domain.NewPushSubscription(uuid.New(), uuid.New(), "https://example.test/push", "p256dh", "auth", nil, time.Now())
	subscriptions := &pushSubscriptionsStub{subscriptions: []*domain.PushSubscription{subscription}}
	sender := &pushSenderStub{}
	service := NewDeliveryService(repo, subscriptions, sender, true)
	service.now = func() time.Time { return time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC) }
	service.newID = func() uuid.UUID { return uuid.MustParse("11111111-1111-1111-1111-111111111111") }

	err := service.NotifyUser(context.Background(), ports.NotifyUserCommand{
		UserID:           subscription.UserID(),
		Type:             domain.NotificationTasksReadyForReview,
		Title:            "Все задачи выполнены",
		Body:             "Участники команды завершили уборку. Задачи можно проверить.",
		TargetURL:        "/app/current-duty?tab=review",
		DeduplicationKey: "tasks_ready_for_review:duty:user",
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, repo.createCalls)
	assert.Equal(t, 1, sender.calls)
}

func TestDeliveryServiceNotifyUserSkipsDuplicate(t *testing.T) {
	repo := &notificationRepositoryStub{created: false}
	sender := &pushSenderStub{}
	service := NewDeliveryService(repo, &pushSubscriptionsStub{}, sender, true)

	err := service.NotifyUser(context.Background(), ports.NotifyUserCommand{
		UserID:           uuid.New(),
		Type:             domain.NotificationTasksReadyForReview,
		Title:            "Все задачи выполнены",
		Body:             "Участники команды завершили уборку. Задачи можно проверить.",
		TargetURL:        "/app/current-duty?tab=review",
		DeduplicationKey: "tasks_ready_for_review:duty:user",
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, sender.calls)
}

func TestDeliveryServiceNotifyUserDeletesExpiredSubscription(t *testing.T) {
	repo := &notificationRepositoryStub{created: true}
	subscription, _ := domain.NewPushSubscription(uuid.New(), uuid.New(), "https://example.test/push", "p256dh", "auth", nil, time.Now())
	subscriptions := &pushSubscriptionsStub{subscriptions: []*domain.PushSubscription{subscription}}
	sender := &pushSenderStub{
		errBySubscription: map[uuid.UUID]error{
			subscription.ID(): ports.ErrPushSubscriptionExpired,
		},
	}
	service := NewDeliveryService(repo, subscriptions, sender, true)

	err := service.NotifyUser(context.Background(), ports.NotifyUserCommand{
		UserID:           subscription.UserID(),
		Type:             domain.NotificationTasksReadyForReview,
		Title:            "Все задачи выполнены",
		Body:             "Участники команды завершили уборку. Задачи можно проверить.",
		TargetURL:        "/app/current-duty?tab=review",
		DeduplicationKey: "tasks_ready_for_review:duty:user",
	})

	assert.NoError(t, err)
	assert.Equal(t, []uuid.UUID{subscription.ID()}, subscriptions.deletedIDs)
}

func TestDeliveryServiceNotifyUserIgnoresTemporaryPushError(t *testing.T) {
	repo := &notificationRepositoryStub{created: true}
	subscription, _ := domain.NewPushSubscription(uuid.New(), uuid.New(), "https://example.test/push", "p256dh", "auth", nil, time.Now())
	subscriptions := &pushSubscriptionsStub{subscriptions: []*domain.PushSubscription{subscription}}
	sender := &pushSenderStub{
		errBySubscription: map[uuid.UUID]error{
			subscription.ID(): errors.New("temporary push failure"),
		},
	}
	service := NewDeliveryService(repo, subscriptions, sender, true)

	err := service.NotifyUser(context.Background(), ports.NotifyUserCommand{
		UserID:           subscription.UserID(),
		Type:             domain.NotificationTasksReadyForReview,
		Title:            "Все задачи выполнены",
		Body:             "Участники команды завершили уборку. Задачи можно проверить.",
		TargetURL:        "/app/current-duty?tab=review",
		DeduplicationKey: "tasks_ready_for_review:duty:user",
	})

	assert.NoError(t, err)
}
