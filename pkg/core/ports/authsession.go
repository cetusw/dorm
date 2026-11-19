package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RequestIDKey struct{}

type SessionRejection string

func (r SessionRejection) Error() string { return string(r) }

type AuthSession struct {
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
}

type SessionUseCase interface {
	Create(context.Context, uuid.UUID, string) (*AuthSession, error)
	Validate(context.Context, string) (*AuthSession, error)
	Revoke(context.Context, string) error
}

type SessionRepository interface {
	Create(context.Context, []byte, uuid.UUID, string, time.Time) error
	Validate(context.Context, []byte) (*AuthSession, error)
	Revoke(context.Context, []byte) error
}
