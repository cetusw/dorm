package catalog

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRestoreArea(t *testing.T) {
	id := 1
	name := "Kitchen"
	floor := 2
	groupID := uuid.New()

	area := RestoreArea(id, name, floor, &groupID)

	assert.Equal(t, id, area.ID())
	assert.Equal(t, name, area.Name())
	assert.Equal(t, floor, area.Floor())
}

func TestRestoreTaskDefinition(t *testing.T) {
	id := uuid.New()
	areaID := 1
	title := "Clean Floor"
	cost := 5
	recurrenceInterval := 4
	startSequence := 1

	task := RestoreTaskDefinition(id, areaID, title, cost, recurrenceInterval, startSequence)

	assert.Equal(t, id, task.ID())
	assert.Equal(t, areaID, task.AreaID())
	assert.Equal(t, title, task.Title())
	assert.Equal(t, cost, task.Cost())
	assert.Equal(t, recurrenceInterval, task.RecurrenceInterval())
	assert.Equal(t, startSequence, task.StartSequence())
}

func TestNewTaskDefinition(t *testing.T) {
	task, err := NewTaskDefinition(1, "Clean Floor", 5, 4, 1)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, task.ID())
	assert.Equal(t, 1, task.AreaID())
	assert.Equal(t, "Clean Floor", task.Title())
	assert.Equal(t, 5, task.Cost())
	assert.Equal(t, 4, task.RecurrenceInterval())
	assert.Equal(t, 1, task.StartSequence())
}

func TestNewTaskDefinition_Validation(t *testing.T) {
	_, err := NewTaskDefinition(0, "Clean Floor", 5, 1, 1)
	assert.ErrorIs(t, err, ErrInvalidTaskDefinition)

	_, err = NewTaskDefinition(1, "", 5, 1, 1)
	assert.ErrorIs(t, err, ErrInvalidTaskDefinition)
}
