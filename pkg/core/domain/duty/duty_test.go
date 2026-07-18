package duty

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewDuty(t *testing.T) {
	teamID := uuid.New()
	start := time.Now()
	end := start.Add(7 * 24 * time.Hour)

	d := NewDuty(teamID, start, end)

	assert.NotNil(t, d)
	assert.NotEqual(t, uuid.Nil, d.ID())
	assert.Equal(t, teamID, d.TeamID())
	assert.Empty(t, d.Tasks())
}

func TestDuty_AddTask(t *testing.T) {
	d := NewDuty(uuid.New(), time.Now(), time.Now().Add(24*time.Hour))
	taskID := uuid.New()
	taskDefID := uuid.New()

	d.AddTask(taskID, taskDefID)

	tasks := d.Tasks()
	assert.Len(t, tasks, 1)
	assert.Equal(t, taskID, tasks[0].ID())
	assert.Equal(t, d.ID(), tasks[0].DutyID())
	assert.Equal(t, taskDefID, tasks[0].TaskDefID())
	assert.Nil(t, tasks[0].AssigneeID())
	assert.Nil(t, tasks[0].CompletionDate())
}

func TestDutyTask_CanCompleteUnassignedTask(t *testing.T) {
	task := RestoreDutyTask(RestoreDutyTaskParams{ID: uuid.New()})

	err := task.CanComplete(uuid.New())

	assert.ErrorIs(t, err, ErrTaskNotAssigned)
}

func TestDutyTask_CanCancelCompletion(t *testing.T) {
	userID := uuid.New()
	completedAt := time.Now()
	task := RestoreDutyTask(RestoreDutyTaskParams{
		ID:             uuid.New(),
		AssigneeID:     &userID,
		CompletionDate: &completedAt,
	})

	err := task.CanCancelCompletion(userID)

	assert.NoError(t, err)
}

func TestRestoreDuty(t *testing.T) {
	id := uuid.New()
	teamID := uuid.New()
	taskID := uuid.New()

	tasks := []*DutyTask{
		RestoreDutyTask(RestoreDutyTaskParams{
			ID: taskID, DutyID: id, TaskDefID: uuid.New(),
		}),
	}

	d := RestoreDuty(id, teamID, time.Now(), time.Now(), tasks)

	assert.Equal(t, id, d.ID())
	assert.Len(t, d.Tasks(), 1)
	assert.Equal(t, taskID, d.Tasks()[0].ID())
}
