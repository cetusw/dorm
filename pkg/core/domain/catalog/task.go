package catalog

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrInvalidTaskDefinition = errors.New("area and title are required")

type TaskDefinition struct {
	id        uuid.UUID
	areaID    int
	title     string
	cost      int
	frequency int
	isActive  bool
}

func NewTaskDefinition(areaID int, title string, cost, frequency int) (*TaskDefinition, error) {
	if areaID == 0 || title == "" {
		return nil, ErrInvalidTaskDefinition
	}
	return RestoreTaskDefinition(uuid.New(), areaID, title, cost, frequency), nil
}

func RestoreTaskDefinition(id uuid.UUID, areaID int, title string, cost, frequency int) *TaskDefinition {
	return RestoreTaskDefinitionWithActive(id, areaID, title, cost, frequency, true)
}

func RestoreTaskDefinitionWithActive(id uuid.UUID, areaID int, title string, cost, frequency int, isActive bool) *TaskDefinition {
	return &TaskDefinition{
		id:        id,
		areaID:    areaID,
		title:     title,
		cost:      cost,
		frequency: frequency,
		isActive:  isActive,
	}
}

func (t *TaskDefinition) ID() uuid.UUID  { return t.id }
func (t *TaskDefinition) Title() string  { return t.title }
func (t *TaskDefinition) Cost() int      { return t.cost }
func (t *TaskDefinition) AreaID() int    { return t.areaID }
func (t *TaskDefinition) Frequency() int { return t.frequency }
func (t *TaskDefinition) IsActive() bool { return t.isActive }

type TaskDefinitionRepository interface {
	GetAllTaskDefinitions(ctx context.Context) ([]*TaskDefinition, error)
	GetActiveTaskDefinitions(ctx context.Context) ([]*TaskDefinition, error)
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*TaskDefinition, error)
	FindCommon(ctx context.Context) ([]*TaskDefinition, error)
	FindByID(ctx context.Context, id uuid.UUID) (*TaskDefinition, error)
	Save(ctx context.Context, task *TaskDefinition) error
	UpdateGroupTaskActivity(ctx context.Context, groupID uuid.UUID, activeTaskIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}
