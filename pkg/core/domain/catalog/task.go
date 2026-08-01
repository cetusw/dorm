package catalog

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidTaskDefinition = errors.New("area and title are required")

func IsValidRecurrenceInterval(value int) bool {
	switch value {
	case 0, 1, 2, 4, 12:
		return true
	default:
		return false
	}
}

type TaskDefinition struct {
	id                 uuid.UUID
	areaID             int
	title              string
	cost               int
	recurrenceInterval int
	startSequence      int
}

func NewTaskDefinition(areaID int, title string, cost, recurrenceInterval, startSequence int) (*TaskDefinition, error) {
	if areaID == 0 || title == "" {
		return nil, ErrInvalidTaskDefinition
	}
	return RestoreTaskDefinition(uuid.New(), areaID, title, cost, recurrenceInterval, startSequence), nil
}

func RestoreTaskDefinition(id uuid.UUID, areaID int, title string, cost, recurrenceInterval, startSequence int) *TaskDefinition {
	return &TaskDefinition{
		id:                 id,
		areaID:             areaID,
		title:              title,
		cost:               cost,
		recurrenceInterval: recurrenceInterval,
		startSequence:      startSequence,
	}
}

func (t *TaskDefinition) ID() uuid.UUID           { return t.id }
func (t *TaskDefinition) Title() string           { return t.title }
func (t *TaskDefinition) Cost() int               { return t.cost }
func (t *TaskDefinition) AreaID() int             { return t.areaID }
func (t *TaskDefinition) RecurrenceInterval() int { return t.recurrenceInterval }
func (t *TaskDefinition) StartSequence() int      { return t.startSequence }
func (t *TaskDefinition) IsScheduledFor(dutySequence int) bool {
	if t.recurrenceInterval <= 0 {
		return false
	}
	if dutySequence < t.startSequence {
		return false
	}
	return (dutySequence-t.startSequence)%t.recurrenceInterval == 0
}

type TaskDefinitionRepository interface {
	GetAllTaskDefinitions(ctx context.Context) ([]*TaskDefinition, error)
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*TaskDefinition, error)
	FindCommon(ctx context.Context) ([]*TaskDefinition, error)
	FindByID(ctx context.Context, id uuid.UUID) (*TaskDefinition, error)
	FindLastCompletionDates(ctx context.Context, taskIDs []uuid.UUID) (map[uuid.UUID]*time.Time, error)
	Save(ctx context.Context, task *TaskDefinition) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type DutyTaskOverride struct {
	taskID            uuid.UUID
	includeInNextDuty bool
}

func NewDutyTaskOverride(taskID uuid.UUID, includeInNextDuty bool) *DutyTaskOverride {
	return &DutyTaskOverride{taskID: taskID, includeInNextDuty: includeInNextDuty}
}

func (o *DutyTaskOverride) TaskID() uuid.UUID       { return o.taskID }
func (o *DutyTaskOverride) IncludeInNextDuty() bool { return o.includeInNextDuty }

type DutyTaskOverrideRepository interface {
	FindAll(ctx context.Context) ([]*DutyTaskOverride, error)
	ReplaceForTasks(ctx context.Context, taskIDs []uuid.UUID, overrides []*DutyTaskOverride) error
	DeleteByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) error
}
