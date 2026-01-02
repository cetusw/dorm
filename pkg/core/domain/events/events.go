package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	TopicTaskCompleted = "task.completed"
	TopicTaskAssigned  = "task.assigned"
	TopicWeekStarted   = "week.started"
)

type TaskCompletedEvent struct {
	TaskID   uuid.UUID
	UserID   uuid.UUID
	Time     time.Time
	TaskCost int
}

type TaskAssignedEvent struct {
	TaskID     uuid.UUID
	AssigneeID uuid.UUID
}

type WeekStartedEvent struct {
	WeekNumber int
	StartDate  time.Time
}
