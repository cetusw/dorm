package duty

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTaskNotFound    = errors.New("task not found in this duty")
	ErrTaskAssigned    = errors.New("task is already assigned")
	ErrTaskNotAssigned = errors.New("task is not assigned")
)

type Duty struct {
	id     uuid.UUID
	teamID uuid.UUID
	start  time.Time
	end    time.Time
	tasks  map[uuid.UUID]*DutyTask
}

func NewDuty(teamID uuid.UUID, start, end time.Time) *Duty {
	return &Duty{
		id:     uuid.New(),
		teamID: teamID,
		start:  start,
		end:    end,
		tasks:  make(map[uuid.UUID]*DutyTask),
	}
}

func RestoreDuty(id, teamID uuid.UUID, start, end time.Time, tasks []*DutyTask) *Duty {
	d := &Duty{
		id:     id,
		teamID: teamID,
		start:  start,
		end:    end,
		tasks:  make(map[uuid.UUID]*DutyTask),
	}
	for _, t := range tasks {
		d.tasks[t.id] = t
	}
	return d
}

func (d *Duty) AddTask(taskID, taskDefID uuid.UUID) {
	d.tasks[taskID] = &DutyTask{
		id:        taskID,
		taskDefID: taskDefID,
	}
}

func (d *Duty) AssignTask(taskID uuid.UUID, userID uuid.UUID) error {
	task, exists := d.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}
	if task.assigneeID != nil {
		return ErrTaskAssigned
	}
	task.assigneeID = &userID
	return nil
}

func (d *Duty) UnassignTask(taskID uuid.UUID) error {
	task, exists := d.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}
	task.assigneeID = nil
	task.completionDate = nil
	return nil
}

func (d *Duty) CompleteTask(taskID uuid.UUID) error {
	task, exists := d.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}
	if task.assigneeID == nil {
		return ErrTaskNotAssigned
	}
	now := time.Now()
	task.completionDate = &now
	return nil
}

func (d *Duty) OpenTask(taskID uuid.UUID) error {
	task, exists := d.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}
	if task.assigneeID == nil {
		return ErrTaskNotAssigned
	}
	task.completionDate = nil
	return nil
}

func (d *Duty) ID() uuid.UUID     { return d.id }
func (d *Duty) TeamID() uuid.UUID { return d.teamID }
func (d *Duty) Tasks() []*DutyTask {
	list := make([]*DutyTask, 0, len(d.tasks))
	for _, t := range d.tasks {
		list = append(list, t)
	}
	return list
}
func (d *Duty) Start() time.Time { return d.start }
func (d *Duty) End() time.Time   { return d.end }

type Repository interface {
	Save(ctx context.Context, duty *Duty) error
	FindCurrentByTeamID(ctx context.Context, teamID uuid.UUID) (*Duty, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Duty, error)
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*Duty, error)
	CountDistinctStartDates(ctx context.Context) (int, error)
	FindLastByTaskDefID(ctx context.Context, taskDefID uuid.UUID) (*Duty, error)
	FindAllLatest(ctx context.Context) ([]*Duty, error)
}
