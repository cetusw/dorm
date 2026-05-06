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
}

func NewTaskDefinition(areaID int, title string, cost, frequency int) (*TaskDefinition, error) {
	if areaID == 0 || title == "" {
		return nil, ErrInvalidTaskDefinition
	}
	return RestoreTaskDefinition(uuid.New(), areaID, title, cost, frequency), nil
}

func RestoreTaskDefinition(id uuid.UUID, areaID int, title string, cost, frequency int) *TaskDefinition {
	return &TaskDefinition{
		id:        id,
		areaID:    areaID,
		title:     title,
		cost:      cost,
		frequency: frequency,
	}
}

func (t *TaskDefinition) ID() uuid.UUID  { return t.id }
func (t *TaskDefinition) Title() string  { return t.title }
func (t *TaskDefinition) Cost() int      { return t.cost }
func (t *TaskDefinition) AreaID() int    { return t.areaID }
func (t *TaskDefinition) Frequency() int { return t.frequency }

type TaskDefinitionRepository interface {
	GetAllTaskDefinitions(ctx context.Context) ([]*TaskDefinition, error)
	FindByID(ctx context.Context, id uuid.UUID) (*TaskDefinition, error)
	Save(ctx context.Context, task *TaskDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
}
