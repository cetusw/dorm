package duty

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ParticipantType string

const (
	ParticipantTypeRegular   ParticipantType = "REGULAR"
	ParticipantTypeTemporary ParticipantType = "TEMPORARY"
)

var (
	ErrParticipantAlreadyActive   = errors.New("duty participant is already active")
	ErrParticipantAlreadyExcluded = errors.New("duty participant is already excluded")
	ErrParticipantNotFound        = errors.New("duty participant not found")
	ErrParticipantPeriodConflict  = errors.New("duty participant has overlapping duty")
	ErrLeaderReplacementRequired  = errors.New("active replacement leader is required")
)

type DutyParticipant struct {
	DutyID        uuid.UUID
	ParticipantID uuid.UUID
	Type          ParticipantType
	ExcludedAt    *time.Time
}

func (p DutyParticipant) Active() bool { return p.ExcludedAt == nil }

type ParticipantRepository interface {
	List(ctx context.Context, dutyID uuid.UUID, includeExcluded bool) ([]DutyParticipant, error)
	IsActive(ctx context.Context, dutyID, participantID uuid.UUID) (bool, error)
	Add(ctx context.Context, participant DutyParticipant, start, end time.Time) error
	Restore(ctx context.Context, dutyID, participantID uuid.UUID, start, end time.Time) error
	Exclude(ctx context.Context, dutyID, participantID uuid.UUID, replacementLeaderID *uuid.UUID, excludedAt time.Time) error
	ChangeLeader(ctx context.Context, dutyID, leaderID uuid.UUID) error
}
