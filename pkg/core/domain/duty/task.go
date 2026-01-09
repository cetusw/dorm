package duty

import (
	"time"

	"github.com/google/uuid"
)

type DutyTask struct {
	id               uuid.UUID
	taskDefID        uuid.UUID
	assigneeID       *uuid.UUID
	completionDate   *time.Time
	verificationDate *time.Time
}

func RestoreDutyTask(
	id uuid.UUID,
	taskDefID uuid.UUID,
	assigneeID *uuid.UUID,
	completionDate *time.Time,
	verificationDate *time.Time,
) *DutyTask {
	return &DutyTask{
		id:               id,
		taskDefID:        taskDefID,
		assigneeID:       assigneeID,
		verificationDate: verificationDate,
		completionDate:   completionDate,
	}
}

func (t *DutyTask) ID() uuid.UUID                { return t.id }
func (t *DutyTask) TaskDefID() uuid.UUID         { return t.taskDefID }
func (t *DutyTask) AssigneeID() *uuid.UUID       { return t.assigneeID }
func (t *DutyTask) CompletionDate() *time.Time   { return t.completionDate }
func (t *DutyTask) VerificationDate() *time.Time { return t.verificationDate }
