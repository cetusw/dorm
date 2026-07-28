package notification

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidPushSubscriptionID = errors.New("invalid push subscription id")
	ErrInvalidPushUserID         = errors.New("invalid push subscription user id")
	ErrEmptyPushEndpoint         = errors.New("push endpoint cannot be empty")
	ErrEmptyPushP256DH           = errors.New("push p256dh cannot be empty")
	ErrEmptyPushAuthSecret       = errors.New("push auth secret cannot be empty")
	ErrInvalidPushEndpoint       = errors.New("invalid push endpoint")
	ErrInvalidPushEndpointHash   = errors.New("invalid push endpoint hash")
)

type PushSubscription struct {
	id           uuid.UUID
	userID       uuid.UUID
	endpoint     string
	endpointHash []byte
	p256dh       string
	authSecret   string
	userAgent    *string
	createdAt    time.Time
	updatedAt    time.Time
}

type RestorePushSubscriptionParams struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Endpoint     string
	EndpointHash []byte
	P256DH       string
	AuthSecret   string
	UserAgent    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewPushSubscription(
	id uuid.UUID,
	userID uuid.UUID,
	endpoint string,
	p256dh string,
	authSecret string,
	userAgent *string,
	now time.Time,
) (*PushSubscription, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidPushSubscriptionID
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidPushUserID
	}

	endpoint = strings.TrimSpace(endpoint)
	p256dh = strings.TrimSpace(p256dh)
	authSecret = strings.TrimSpace(authSecret)

	if endpoint == "" {
		return nil, ErrEmptyPushEndpoint
	}
	if p256dh == "" {
		return nil, ErrEmptyPushP256DH
	}
	if authSecret == "" {
		return nil, ErrEmptyPushAuthSecret
	}

	parsedEndpoint, err := url.ParseRequestURI(endpoint)
	if err != nil || parsedEndpoint == nil || parsedEndpoint.Scheme != "https" {
		return nil, ErrInvalidPushEndpoint
	}

	var normalizedUserAgent *string
	if userAgent != nil {
		trimmedUserAgent := strings.TrimSpace(*userAgent)
		if trimmedUserAgent != "" {
			normalizedUserAgent = &trimmedUserAgent
		}
	}

	hash := HashPushEndpoint(endpoint)

	return &PushSubscription{
		id:           id,
		userID:       userID,
		endpoint:     endpoint,
		endpointHash: append([]byte(nil), hash[:]...),
		p256dh:       p256dh,
		authSecret:   authSecret,
		userAgent:    normalizedUserAgent,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func RestorePushSubscription(params RestorePushSubscriptionParams) *PushSubscription {
	return &PushSubscription{
		id:           params.ID,
		userID:       params.UserID,
		endpoint:     params.Endpoint,
		endpointHash: append([]byte(nil), params.EndpointHash...),
		p256dh:       params.P256DH,
		authSecret:   params.AuthSecret,
		userAgent:    copyOptionalString(params.UserAgent),
		createdAt:    params.CreatedAt,
		updatedAt:    params.UpdatedAt,
	}
}

func HashPushEndpoint(endpoint string) [sha256.Size]byte {
	return sha256.Sum256([]byte(endpoint))
}

func (s *PushSubscription) ID() uuid.UUID      { return s.id }
func (s *PushSubscription) UserID() uuid.UUID  { return s.userID }
func (s *PushSubscription) Endpoint() string   { return s.endpoint }
func (s *PushSubscription) P256DH() string     { return s.p256dh }
func (s *PushSubscription) AuthSecret() string { return s.authSecret }

func (s *PushSubscription) EndpointHash() []byte {
	return append([]byte(nil), s.endpointHash...)
}

func (s *PushSubscription) UserAgent() *string {
	return copyOptionalString(s.userAgent)
}

func (s *PushSubscription) CreatedAt() time.Time { return s.createdAt }
func (s *PushSubscription) UpdatedAt() time.Time { return s.updatedAt }

type PushSubscriptionRepository interface {
	Save(ctx context.Context, subscription *PushSubscription) error
	FindByEndpointHash(ctx context.Context, endpointHash []byte) (*PushSubscription, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*PushSubscription, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteByEndpointHash(ctx context.Context, userID uuid.UUID, endpointHash []byte) error
}

func copyOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
