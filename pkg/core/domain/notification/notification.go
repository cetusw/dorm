package notification

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxNotificationTitleLength            = 255
	MaxNotificationBodyLength             = 4096
	MaxNotificationTargetURLLength        = 1024
	MaxNotificationDeduplicationKeyLength = 255
)

var (
	ErrInvalidNotificationID        = errors.New("invalid notification id")
	ErrInvalidNotificationUserID    = errors.New("invalid notification user id")
	ErrInvalidNotificationType      = errors.New("invalid notification type")
	ErrEmptyNotificationTitle       = errors.New("notification title cannot be empty")
	ErrEmptyNotificationBody        = errors.New("notification body cannot be empty")
	ErrEmptyNotificationTargetURL   = errors.New("notification target url cannot be empty")
	ErrEmptyDeduplicationKey        = errors.New("notification deduplication key cannot be empty")
	ErrNotificationTitleTooLong     = errors.New("notification title is too long")
	ErrNotificationBodyTooLong      = errors.New("notification body is too long")
	ErrNotificationTargetURLTooLong = errors.New("notification target url is too long")
	ErrDeduplicationKeyTooLong      = errors.New("notification deduplication key is too long")
	ErrInvalidNotificationTargetURL = errors.New("invalid notification target url")
)

type Notification struct {
	id               uuid.UUID
	userID           uuid.UUID
	notificationType NotificationType
	title            string
	body             string
	targetURL        string
	deduplicationKey string
	createdAt        time.Time
	readAt           *time.Time
}

type RestoreNotificationParams struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	Type             NotificationType
	Title            string
	Body             string
	TargetURL        string
	DeduplicationKey string
	CreatedAt        time.Time
	ReadAt           *time.Time
}

func NewNotification(
	id uuid.UUID,
	userID uuid.UUID,
	notificationType NotificationType,
	title string,
	body string,
	targetURL string,
	deduplicationKey string,
	createdAt time.Time,
) (*Notification, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidNotificationID
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidNotificationUserID
	}
	if !notificationType.IsValid() {
		return nil, ErrInvalidNotificationType
	}

	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	targetURL = strings.TrimSpace(targetURL)
	deduplicationKey = strings.TrimSpace(deduplicationKey)

	switch {
	case title == "":
		return nil, ErrEmptyNotificationTitle
	case body == "":
		return nil, ErrEmptyNotificationBody
	case targetURL == "":
		return nil, ErrEmptyNotificationTargetURL
	case deduplicationKey == "":
		return nil, ErrEmptyDeduplicationKey
	case len(title) > MaxNotificationTitleLength:
		return nil, ErrNotificationTitleTooLong
	case len(body) > MaxNotificationBodyLength:
		return nil, ErrNotificationBodyTooLong
	case len(targetURL) > MaxNotificationTargetURLLength:
		return nil, ErrNotificationTargetURLTooLong
	case len(deduplicationKey) > MaxNotificationDeduplicationKeyLength:
		return nil, ErrDeduplicationKeyTooLong
	}

	if err := ValidateNotificationTargetURL(targetURL); err != nil {
		return nil, err
	}

	return &Notification{
		id:               id,
		userID:           userID,
		notificationType: notificationType,
		title:            title,
		body:             body,
		targetURL:        targetURL,
		deduplicationKey: deduplicationKey,
		createdAt:        createdAt,
		readAt:           nil,
	}, nil
}

func RestoreNotification(params RestoreNotificationParams) *Notification {
	return &Notification{
		id:               params.ID,
		userID:           params.UserID,
		notificationType: params.Type,
		title:            params.Title,
		body:             params.Body,
		targetURL:        params.TargetURL,
		deduplicationKey: params.DeduplicationKey,
		createdAt:        params.CreatedAt,
		readAt:           copyNotificationTime(params.ReadAt),
	}
}

func (n *Notification) ID() uuid.UUID            { return n.id }
func (n *Notification) UserID() uuid.UUID        { return n.userID }
func (n *Notification) Type() NotificationType   { return n.notificationType }
func (n *Notification) Title() string            { return n.title }
func (n *Notification) Body() string             { return n.body }
func (n *Notification) TargetURL() string        { return n.targetURL }
func (n *Notification) DeduplicationKey() string { return n.deduplicationKey }
func (n *Notification) CreatedAt() time.Time     { return n.createdAt }
func (n *Notification) ReadAt() *time.Time       { return copyNotificationTime(n.readAt) }

func ValidateNotificationTargetURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return ErrInvalidNotificationTargetURL
	}

	if parsed.IsAbs() || parsed.Host != "" {
		return ErrInvalidNotificationTargetURL
	}

	if strings.HasPrefix(value, "//") {
		return ErrInvalidNotificationTargetURL
	}

	if value == "/app" {
		return nil
	}

	if !strings.HasPrefix(value, "/app/") {
		return ErrInvalidNotificationTargetURL
	}

	return nil
}

func BuildDutyNotificationDeduplicationKey(
	notificationType NotificationType,
	dutyID uuid.UUID,
	userID uuid.UUID,
) (string, error) {
	if !notificationType.IsValid() {
		return "", ErrInvalidNotificationType
	}
	if dutyID == uuid.Nil {
		return "", fmt.Errorf("invalid duty id")
	}
	if userID == uuid.Nil {
		return "", fmt.Errorf("invalid user id")
	}

	key := fmt.Sprintf("%s:%s:%s", notificationType, dutyID.String(), userID.String())
	if len(key) > MaxNotificationDeduplicationKeyLength {
		return "", ErrDeduplicationKeyTooLong
	}

	return key, nil
}

func copyNotificationTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	copied := *value
	return &copied
}
