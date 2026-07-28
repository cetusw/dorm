package notification

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	domain "dorm/pkg/core/domain/notification"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

const (
	MaxPushEndpointLength = 4096
	MaxPushKeyLength      = 255
)

type Service struct {
	repository domain.PushSubscriptionRepository
	now        func() time.Time
}

func NewNotificationService(repository domain.PushSubscriptionRepository) *Service {
	return &Service{
		repository: repository,
		now:        time.Now,
	}
}

func (s *Service) Subscribe(ctx context.Context, userID uuid.UUID, request dto.SubscribeToPushRequest) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user id is required")
	}

	endpoint, err := normalizeEndpoint(request.Endpoint)
	if err != nil {
		return err
	}

	p256dh, err := normalizePushKey(request.P256DH, "p256dh")
	if err != nil {
		return err
	}

	authSecret, err := normalizePushKey(request.AuthSecret, "auth")
	if err != nil {
		return err
	}

	subscription, err := domain.NewPushSubscription(
		uuid.New(),
		userID,
		endpoint,
		p256dh,
		authSecret,
		normalizeOptionalString(request.UserAgent),
		s.now(),
	)
	if err != nil {
		return fmt.Errorf("create push subscription: %w", err)
	}

	if err := s.repository.Save(ctx, subscription); err != nil {
		return fmt.Errorf("save push subscription: %w", err)
	}

	return nil
}

func (s *Service) Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user id is required")
	}

	normalizedEndpoint, err := normalizeEndpoint(endpoint)
	if err != nil {
		return err
	}

	hash := domain.HashPushEndpoint(normalizedEndpoint)
	if len(hash) != sha256.Size {
		return fmt.Errorf("invalid push endpoint hash")
	}

	if err := s.repository.DeleteByEndpointHash(ctx, userID, hash[:]); err != nil {
		return fmt.Errorf("delete push subscription: %w", err)
	}

	return nil
}

func normalizeEndpoint(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", domain.ErrEmptyPushEndpoint
	}
	if len(value) > MaxPushEndpointLength {
		return "", fmt.Errorf("push endpoint is too long")
	}
	return value, nil
}

func normalizePushKey(value string, fieldName string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		if fieldName == "p256dh" {
			return "", domain.ErrEmptyPushP256DH
		}
		return "", domain.ErrEmptyPushAuthSecret
	}
	if len(value) > MaxPushKeyLength {
		return "", fmt.Errorf("%s is too long", fieldName)
	}
	return value, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
