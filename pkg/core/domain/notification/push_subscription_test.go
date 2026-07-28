package notification

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewPushSubscription(t *testing.T) {
	now := time.Now().UTC()
	userAgent := "  Safari on iPhone  "

	subscription, err := NewPushSubscription(
		uuid.New(),
		uuid.New(),
		"  https://example.test/push/subscription-1  ",
		"  p256dh-value  ",
		"  auth-secret  ",
		&userAgent,
		now,
	)

	assert.NoError(t, err)
	assert.NotNil(t, subscription)
	assert.Equal(t, "https://example.test/push/subscription-1", subscription.Endpoint())
	assert.Equal(t, "p256dh-value", subscription.P256DH())
	assert.Equal(t, "auth-secret", subscription.AuthSecret())
	assert.Equal(t, now, subscription.CreatedAt())
	assert.Equal(t, now, subscription.UpdatedAt())
	expectedHash := HashPushEndpoint("https://example.test/push/subscription-1")
	assert.Equal(t, expectedHash[:], subscription.EndpointHash())

	if assert.NotNil(t, subscription.UserAgent()) {
		assert.Equal(t, "Safari on iPhone", *subscription.UserAgent())
	}
}

func TestNewPushSubscriptionValidation(t *testing.T) {
	now := time.Now().UTC()
	validID := uuid.New()
	validUserID := uuid.New()

	testCases := []struct {
		name       string
		id         uuid.UUID
		userID     uuid.UUID
		endpoint   string
		p256dh     string
		authSecret string
		expected   error
	}{
		{name: "invalid id", id: uuid.Nil, userID: validUserID, endpoint: "https://example.test", p256dh: "key", authSecret: "secret", expected: ErrInvalidPushSubscriptionID},
		{name: "invalid user id", id: validID, userID: uuid.Nil, endpoint: "https://example.test", p256dh: "key", authSecret: "secret", expected: ErrInvalidPushUserID},
		{name: "empty endpoint", id: validID, userID: validUserID, endpoint: " ", p256dh: "key", authSecret: "secret", expected: ErrEmptyPushEndpoint},
		{name: "empty p256dh", id: validID, userID: validUserID, endpoint: "https://example.test", p256dh: " ", authSecret: "secret", expected: ErrEmptyPushP256DH},
		{name: "empty auth secret", id: validID, userID: validUserID, endpoint: "https://example.test", p256dh: "key", authSecret: " ", expected: ErrEmptyPushAuthSecret},
		{name: "invalid endpoint", id: validID, userID: validUserID, endpoint: "://bad", p256dh: "key", authSecret: "secret", expected: ErrInvalidPushEndpoint},
		{name: "non https endpoint", id: validID, userID: validUserID, endpoint: "http://example.test", p256dh: "key", authSecret: "secret", expected: ErrInvalidPushEndpoint},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewPushSubscription(
				testCase.id,
				testCase.userID,
				testCase.endpoint,
				testCase.p256dh,
				testCase.authSecret,
				nil,
				now,
			)

			assert.ErrorIs(t, err, testCase.expected)
		})
	}
}

func TestHashPushEndpoint(t *testing.T) {
	endpoint := "https://example.test/push/subscription-1"

	hash := HashPushEndpoint(endpoint)

	assert.Equal(t, sha256.Sum256([]byte(endpoint)), hash)
}

func TestRestorePushSubscriptionCopiesData(t *testing.T) {
	endpointHash := []byte("01234567890123456789012345678901")
	userAgent := "Test browser"

	subscription := RestorePushSubscription(RestorePushSubscriptionParams{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Endpoint:     "https://example.test/push/subscription-1",
		EndpointHash: endpointHash,
		P256DH:       "p256dh",
		AuthSecret:   "auth",
		UserAgent:    &userAgent,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	})

	endpointHash[0] = 'x'
	userAgent = "Changed browser"

	assert.Equal(t, byte('0'), subscription.EndpointHash()[0])
	if assert.NotNil(t, subscription.UserAgent()) {
		assert.Equal(t, "Test browser", *subscription.UserAgent())
	}
}
