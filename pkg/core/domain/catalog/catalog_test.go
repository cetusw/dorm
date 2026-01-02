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
