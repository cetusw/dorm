package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ActiveDuty struct {
	ID      uuid.UUID
	GroupID uuid.UUID
	TeamID  uuid.UUID
	StartAt time.Time
	EndAt   time.Time
}

type DutyFinishReminderState struct {
	FreeTaskCount              int
	UsersBelowAssignedGoalIDs  []uuid.UUID
	UsersWithIncompleteTaskIDs []uuid.UUID
}

type ActiveDutyQuery interface {
	FindAllActive(ctx context.Context, at time.Time) ([]ActiveDuty, error)
}

type DutyTeamMembersQuery interface {
	FindUserIDsByTeamID(ctx context.Context, teamID uuid.UUID) ([]uuid.UUID, error)
}

// DutyParticipantsQuery is deliberately separate so older adapters/tests can
// retain the legacy team lookup while production uses the duty snapshot.
type DutyParticipantsQuery interface {
	FindActiveParticipantIDsByDutyID(ctx context.Context, dutyID uuid.UUID) ([]uuid.UUID, error)
}

type DutyFinishReminderQuery interface {
	GetByDutyID(ctx context.Context, dutyID uuid.UUID, teamID uuid.UUID) (DutyFinishReminderState, error)
}
