package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	TopicTaskCompleted       = "task.completed"
	TopicTaskUncompleted     = "task.uncompleted"
	TopicTaskAssigned        = "task.assigned"
	TopicTasksReadyForReview = "tasks.ready_for_review"
	TopicWeekStarted         = "week.started"
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

type TasksReadyForReviewEvent struct {
	DutyID     uuid.UUID
	TeamID     uuid.UUID
	TeamHeadID uuid.UUID
	OccurredAt time.Time
}

type StartedDuty struct {
	DutyID uuid.UUID
	TeamID uuid.UUID
}

type WeekStartedEvent struct {
	Duties []StartedDuty
}
