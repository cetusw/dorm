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
	freq := 7

	task := RestoreTaskDefinition(id, areaID, title, cost, freq)

	assert.Equal(t, id, task.ID())
	assert.Equal(t, areaID, task.AreaID())
	assert.Equal(t, title, task.Title())
	assert.Equal(t, cost, task.Cost())
	assert.Equal(t, freq, task.Frequency())
}

func TestNewTaskDefinition(t *testing.T) {
	task, err := NewTaskDefinition(1, "Clean Floor", 5, 7)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, task.ID())
	assert.Equal(t, 1, task.AreaID())
	assert.Equal(t, "Clean Floor", task.Title())
	assert.Equal(t, 5, task.Cost())
	assert.Equal(t, 7, task.Frequency())
}

func TestNewTaskDefinition_Validation(t *testing.T) {
	_, err := NewTaskDefinition(0, "Clean Floor", 5, 7)
	assert.ErrorIs(t, err, ErrInvalidTaskDefinition)

	_, err = NewTaskDefinition(1, "", 5, 7)
	assert.ErrorIs(t, err, ErrInvalidTaskDefinition)
}
