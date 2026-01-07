package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	TopicTaskCompleted   = "task.completed"
	TopicTaskUncompleted = "task.uncompleted"
	TopicTaskAssigned    = "task.assigned"
	TopicWeekStarted     = "week.started"
)

type TaskCompletedEvent struct {
	TaskID uuid.UUID
	UserID uuid.UUID
	Time   time.Time
}

type TaskUncompletedEvent struct {
	TaskID uuid.UUID
	UserID uuid.UUID
	Time   time.Time
}

type TaskAssignedEvent struct {
	TaskID     uuid.UUID
	AssigneeID *uuid.UUID
}

type WeekStartedEvent struct {
	StartDate time.Time
	EndDate   time.Time
}
