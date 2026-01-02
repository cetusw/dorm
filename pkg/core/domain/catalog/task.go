package catalog

import (
	"context"

	"github.com/google/uuid"
)

type TaskDefinition struct {
	id        uuid.UUID
	areaID    int
	title     string
	cost      int
	frequency int
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
}
