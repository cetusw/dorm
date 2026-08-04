package duty

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTaskNotFound           = errors.New("duty task not found")
	ErrTaskAlreadyIncluded    = errors.New("duty task already included")
	ErrTaskAssigned           = errors.New("duty task already assigned")
	ErrTaskNotAssigned        = errors.New("duty task not assigned")
	ErrTaskOwnedByAnotherUser = errors.New("duty task belongs to another user")
	ErrTaskAlreadyCompleted   = errors.New("duty task already completed")
	ErrTaskNotCompleted       = errors.New("duty task not completed")
	ErrTaskAlreadyVerified    = errors.New("duty task already verified")
	ErrTaskStateConflict      = errors.New("duty task state conflict")
	ErrAssigneeRequired       = errors.New("assignee is required")
	ErrReviewerRequired       = errors.New("reviewer is required")
	ErrTaskAccessDenied       = errors.New("duty task access denied")
	ErrDutyActionsUnavailable = errors.New("duty actions unavailable")
	ErrDutyPeriodOverlap      = errors.New("duty period overlap")
)

type Duty struct {
	id             uuid.UUID
	teamID         uuid.UUID
	start          time.Time
	end            time.Time
	sequenceNumber int
	tasks          map[uuid.UUID]*DutyTask
}

func NewDuty(teamID uuid.UUID, start, end time.Time, sequenceNumber int) *Duty {
	return &Duty{
		id:             uuid.New(),
		teamID:         teamID,
		start:          start,
		end:            end,
		sequenceNumber: sequenceNumber,
		tasks:          make(map[uuid.UUID]*DutyTask),
	}
}

func RestoreDuty(id, teamID uuid.UUID, start, end time.Time, sequenceNumber int, tasks []*DutyTask) *Duty {
	d := &Duty{
		id:             id,
		teamID:         teamID,
		start:          start,
		end:            end,
		sequenceNumber: sequenceNumber,
		tasks:          make(map[uuid.UUID]*DutyTask),
	}
	for _, t := range tasks {
		d.tasks[t.id] = t
	}
	return d
}

func (d *Duty) AddTask(taskID, taskDefID uuid.UUID) {
	d.tasks[taskID] = &DutyTask{
		id:        taskID,
		dutyID:    d.id,
		taskDefID: taskDefID,
	}
}

func (d *Duty) ID() uuid.UUID     { return d.id }
func (d *Duty) TeamID() uuid.UUID { return d.teamID }
func (d *Duty) SequenceNumber() int {
	return d.sequenceNumber
}
func (d *Duty) Tasks() []*DutyTask {
	list := make([]*DutyTask, 0, len(d.tasks))
	for _, t := range d.tasks {
		list = append(list, t)
	}
	return list
}
func (d *Duty) Start() time.Time { return d.start }
func (d *Duty) End() time.Time   { return d.end }
func (d *Duty) StartDate() time.Time {
	return d.start
}
func (d *Duty) EndDate() time.Time {
	return d.end
}
func (d *Duty) IsActiveAt(at time.Time) bool {
	return !at.Before(d.start) && at.Before(d.end)
}
func (d *Duty) BelongsToTeam(teamID uuid.UUID) bool {
	return d.teamID == teamID
}

type DutyRepository interface {
	CreateWithTasks(ctx context.Context, currentDuty *Duty, tasks []*DutyTask) error
	FindCurrentByTeamID(ctx context.Context, teamID uuid.UUID) (*Duty, error)
	FindActiveByTeamID(ctx context.Context, teamID uuid.UUID, at time.Time) (*Duty, error)
	FindLatestByTeamID(ctx context.Context, teamID uuid.UUID) (*Duty, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Duty, error)
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*Duty, error)
	FindLatestByGroupID(ctx context.Context, groupID uuid.UUID) (*Duty, error)
	FindHistoryByGroupID(ctx context.Context, groupID uuid.UUID) ([]DutyHistoryEntry, error)
	ReassignTeamAndResetTasks(ctx context.Context, dutyID uuid.UUID, teamID uuid.UUID) error
	CountDistinctStartDates(ctx context.Context) (int, error)
	FindLastByTaskDefID(ctx context.Context, taskDefID uuid.UUID) (*Duty, error)
	FindAllLatest(ctx context.Context) ([]*Duty, error)
}

type DutyHistoryEntry struct {
	DutyID              uuid.UUID
	TeamID              uuid.UUID
	Start               time.Time
	End                 time.Time
	SequenceNumber      int
	TeamLeaderName      string
	TotalCostSum        int
	TakenCostSum        int
	TotalTasksCount     int
	TakenTasksCount     int
	CompletedTasksCount int
	VerifiedTasksCount  int
}
