package notification

import (
	"context"
	"testing"
	"time"

	domain "dorm/pkg/core/domain/notification"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type pushSubscriptionRepositoryStub struct {
	savedSubscriptions  []*domain.PushSubscription
	deletedUserID       uuid.UUID
	deletedEndpointHash []byte
	saveErr             error
	deleteByEndpointErr error
}

func (s *pushSubscriptionRepositoryStub) Save(_ context.Context, subscription *domain.PushSubscription) error {
	s.savedSubscriptions = append(s.savedSubscriptions, subscription)
	return s.saveErr
}

func (s *pushSubscriptionRepositoryStub) FindByEndpointHash(_ context.Context, _ []byte) (*domain.PushSubscription, error) {
	return nil, nil
}

func (s *pushSubscriptionRepositoryStub) FindByUserID(_ context.Context, _ uuid.UUID) ([]*domain.PushSubscription, error) {
	return nil, nil
}

func (s *pushSubscriptionRepositoryStub) DeleteByID(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *pushSubscriptionRepositoryStub) DeleteByEndpointHash(_ context.Context, userID uuid.UUID, endpointHash []byte) error {
	s.deletedUserID = userID
	s.deletedEndpointHash = append([]byte(nil), endpointHash...)
	return s.deleteByEndpointErr
}

func TestServiceSubscribeSavesSubscription(t *testing.T) {
	repository := &pushSubscriptionRepositoryStub{}
	service := NewNotificationService(repository)
	now := time.Date(2026, time.July, 28, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	userAgent := " Test Browser "
	userID := uuid.New()

	err := service.Subscribe(context.Background(), userID, dto.SubscribeToPushRequest{
		Endpoint:   " https://example.test/push/subscription ",
		P256DH:     " test-p256dh ",
		AuthSecret: " test-auth ",
		UserAgent:  &userAgent,
	})

	assert.NoError(t, err)
	if assert.Len(t, repository.savedSubscriptions, 1) {
		subscription := repository.savedSubscriptions[0]
		assert.Equal(t, userID, subscription.UserID())
		assert.Equal(t, "https://example.test/push/subscription", subscription.Endpoint())
		assert.Equal(t, "test-p256dh", subscription.P256DH())
		assert.Equal(t, "test-auth", subscription.AuthSecret())
		assert.Equal(t, now, subscription.CreatedAt())
		assert.Equal(t, now, subscription.UpdatedAt())
		if assert.NotNil(t, subscription.UserAgent()) {
			assert.Equal(t, "Test Browser", *subscription.UserAgent())
		}
	}
}

func TestServiceSubscribeValidation(t *testing.T) {
	service := NewNotificationService(&pushSubscriptionRepositoryStub{})
	userID := uuid.New()

	testCases := []struct {
		name    string
		userID  uuid.UUID
		request dto.SubscribeToPushRequest
	}{
		{name: "missing user id", userID: uuid.Nil, request: dto.SubscribeToPushRequest{Endpoint: "https://example.test", P256DH: "a", AuthSecret: "b"}},
		{name: "empty endpoint", userID: userID, request: dto.SubscribeToPushRequest{Endpoint: " ", P256DH: "a", AuthSecret: "b"}},
		{name: "http endpoint", userID: userID, request: dto.SubscribeToPushRequest{Endpoint: "http://example.test", P256DH: "a", AuthSecret: "b"}},
		{name: "empty p256dh", userID: userID, request: dto.SubscribeToPushRequest{Endpoint: "https://example.test", P256DH: " ", AuthSecret: "b"}},
		{name: "empty auth", userID: userID, request: dto.SubscribeToPushRequest{Endpoint: "https://example.test", P256DH: "a", AuthSecret: " "}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := service.Subscribe(context.Background(), testCase.userID, testCase.request)
			assert.Error(t, err)
		})
	}
}

func TestServiceUnsubscribeDeletesOnlyForUserAndEndpoint(t *testing.T) {
	repository := &pushSubscriptionRepositoryStub{}
	service := NewNotificationService(repository)
	userID := uuid.New()
	endpoint := " https://example.test/push/subscription "

	err := service.Unsubscribe(context.Background(), userID, endpoint)

	assert.NoError(t, err)
	assert.Equal(t, userID, repository.deletedUserID)
	expectedHash := domain.HashPushEndpoint("https://example.test/push/subscription")
	assert.Equal(t, expectedHash[:], repository.deletedEndpointHash)
}

func TestServiceUnsubscribeIsIdempotent(t *testing.T) {
	service := NewNotificationService(&pushSubscriptionRepositoryStub{})

	err := service.Unsubscribe(context.Background(), uuid.New(), "https://example.test/push/subscription")

	assert.NoError(t, err)
}
