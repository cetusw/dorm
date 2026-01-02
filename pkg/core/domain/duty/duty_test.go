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

func TestDuty_TaskLifecycle(t *testing.T) {
	d := NewDuty(uuid.New(), time.Now(), time.Now().Add(24*time.Hour))
	taskID := uuid.New()
	taskDefID := uuid.New()
	userID := uuid.New()
	otherTaskID := uuid.New()

	d.AddTask(taskID, taskDefID)

	tasks := d.Tasks()
	assert.Len(t, tasks, 1)
	assert.Equal(t, taskID, tasks[0].ID())
	assert.Equal(t, taskDefID, tasks[0].TaskDefID())
	assert.Nil(t, tasks[0].AssigneeID())
	assert.False(t, tasks[0].IsCompleted())

	err := d.AssignTask(otherTaskID, userID)
	assert.ErrorIs(t, err, ErrTaskNotFound)

	err = d.AssignTask(taskID, userID)
	assert.NoError(t, err)

	tasks = d.Tasks()
	assert.Equal(t, &userID, tasks[0].AssigneeID())

	err = d.CompleteTask(otherTaskID)
	assert.ErrorIs(t, err, ErrTaskNotFound)

	err = d.CompleteTask(taskID)
	assert.NoError(t, err)

	tasks = d.Tasks()
	assert.True(t, tasks[0].IsCompleted())
	assert.NotNil(t, tasks[0].CompletionDate())
	assert.WithinDuration(t, time.Now(), *tasks[0].CompletionDate(), time.Second)
}

func TestDuty_CompleteUnassignedTask(t *testing.T) {
	d := NewDuty(uuid.New(), time.Now(), time.Now())
	taskID := uuid.New()
	d.AddTask(taskID, uuid.New())

	err := d.CompleteTask(taskID)
	assert.ErrorIs(t, err, ErrTaskNotAssigned)
}

func TestDuty_UnassignTask(t *testing.T) {
	d := NewDuty(uuid.New(), time.Now(), time.Now())
	taskID := uuid.New()
	userID := uuid.New()

	d.AddTask(taskID, uuid.New())
	_ = d.AssignTask(taskID, userID)
	_ = d.CompleteTask(taskID)

	err := d.UnassignTask(taskID)
	assert.NoError(t, err)

	tasks := d.Tasks()
	assert.Nil(t, tasks[0].AssigneeID())
	assert.False(t, tasks[0].IsCompleted())
	assert.Nil(t, tasks[0].CompletionDate())
}

func TestRestoreDuty(t *testing.T) {
	id := uuid.New()
	teamID := uuid.New()
	taskID := uuid.New()

	tasks := []*DutyTask{
		RestoreDutyTask(taskID, uuid.New(), nil, nil),
	}

	d := RestoreDuty(id, teamID, time.Now(), time.Now(), tasks)

	assert.Equal(t, id, d.ID())
	assert.Len(t, d.Tasks(), 1)
	assert.Equal(t, taskID, d.Tasks()[0].ID())
}
