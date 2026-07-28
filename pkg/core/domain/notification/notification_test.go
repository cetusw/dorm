package notification

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewNotification(t *testing.T) {
	now := time.Now().UTC()

	n, err := NewNotification(
		uuid.New(),
		uuid.New(),
		NotificationTasksReadyForReview,
		" Все задачи выполнены ",
		" Участники команды завершили уборку. ",
		"/app/current-duty?tab=review",
		" tasks_ready_for_review:1:2 ",
		now,
	)

	assert.NoError(t, err)
	assert.Equal(t, "Все задачи выполнены", n.Title())
	assert.Equal(t, "Участники команды завершили уборку.", n.Body())
	assert.Equal(t, "/app/current-duty?tab=review", n.TargetURL())
	assert.Equal(t, "tasks_ready_for_review:1:2", n.DeduplicationKey())
	assert.Nil(t, n.ReadAt())
}

func TestNewNotificationValidation(t *testing.T) {
	now := time.Now().UTC()
	validID := uuid.New()
	validUserID := uuid.New()

	testCases := []struct {
		name             string
		id               uuid.UUID
		userID           uuid.UUID
		notificationType NotificationType
		title            string
		body             string
		targetURL        string
		deduplicationKey string
		expected         error
	}{
		{"invalid id", uuid.Nil, validUserID, NotificationTasksReadyForReview, "a", "b", "/app", "k", ErrInvalidNotificationID},
		{"invalid user id", validID, uuid.Nil, NotificationTasksReadyForReview, "a", "b", "/app", "k", ErrInvalidNotificationUserID},
		{"invalid type", validID, validUserID, NotificationType("bad"), "a", "b", "/app", "k", ErrInvalidNotificationType},
		{"empty title", validID, validUserID, NotificationTasksReadyForReview, " ", "b", "/app", "k", ErrEmptyNotificationTitle},
		{"empty body", validID, validUserID, NotificationTasksReadyForReview, "a", " ", "/app", "k", ErrEmptyNotificationBody},
		{"empty url", validID, validUserID, NotificationTasksReadyForReview, "a", "b", " ", "k", ErrEmptyNotificationTargetURL},
		{"empty dedup", validID, validUserID, NotificationTasksReadyForReview, "a", "b", "/app", " ", ErrEmptyDeduplicationKey},
		{"javascript url", validID, validUserID, NotificationTasksReadyForReview, "a", "b", "javascript:alert(1)", "k", ErrInvalidNotificationTargetURL},
		{"external url", validID, validUserID, NotificationTasksReadyForReview, "a", "b", "https://example.com", "k", ErrInvalidNotificationTargetURL},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewNotification(
				testCase.id,
				testCase.userID,
				testCase.notificationType,
				testCase.title,
				testCase.body,
				testCase.targetURL,
				testCase.deduplicationKey,
				now,
			)

			assert.ErrorIs(t, err, testCase.expected)
		})
	}
}

func TestBuildDutyNotificationDeduplicationKey(t *testing.T) {
	dutyID := uuid.New()
	userID := uuid.New()

	key, err := BuildDutyNotificationDeduplicationKey(NotificationTasksReadyForReview, dutyID, userID)

	assert.NoError(t, err)
	assert.Equal(t, "tasks_ready_for_review:"+dutyID.String()+":"+userID.String(), key)
}
