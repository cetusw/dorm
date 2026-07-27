package duty

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DutyTaskStatus string

const (
	DutyTaskStatusFree      DutyTaskStatus = "free"
	DutyTaskStatusAssigned  DutyTaskStatus = "assigned"
	DutyTaskStatusCompleted DutyTaskStatus = "completed"
	DutyTaskStatusVerified  DutyTaskStatus = "verified"
)

type DutyTask struct {
	id               uuid.UUID
	dutyID           uuid.UUID
	taskDefID        uuid.UUID
	assigneeID       *uuid.UUID
	reviewerID       *uuid.UUID
	assignmentDate   *time.Time
	completionDate   *time.Time
	verificationDate *time.Time
}

type RestoreDutyTaskParams struct {
	ID               uuid.UUID
	DutyID           uuid.UUID
	TaskDefID        uuid.UUID
	AssigneeID       *uuid.UUID
	ReviewerID       *uuid.UUID
	AssignmentDate   *time.Time
	CompletionDate   *time.Time
	VerificationDate *time.Time
}

func RestoreDutyTask(params RestoreDutyTaskParams) *DutyTask {
	return &DutyTask{
		id:               params.ID,
		dutyID:           params.DutyID,
		taskDefID:        params.TaskDefID,
		assigneeID:       copyUUID(params.AssigneeID),
		reviewerID:       copyUUID(params.ReviewerID),
		assignmentDate:   copyTime(params.AssignmentDate),
		completionDate:   copyTime(params.CompletionDate),
		verificationDate: copyTime(params.VerificationDate),
	}
}

func NewDutyTask(dutyID, taskDefID uuid.UUID) *DutyTask {
	return &DutyTask{
		id:        uuid.New(),
		dutyID:    dutyID,
		taskDefID: taskDefID,
	}
}

func (t *DutyTask) ID() uuid.UUID        { return t.id }
func (t *DutyTask) DutyID() uuid.UUID    { return t.dutyID }
func (t *DutyTask) TaskDefID() uuid.UUID { return t.taskDefID }
func (t *DutyTask) AssigneeID() *uuid.UUID {
	return copyUUID(t.assigneeID)
}
func (t *DutyTask) ReviewerID() *uuid.UUID {
	return copyUUID(t.reviewerID)
}
func (t *DutyTask) AssignmentDate() *time.Time {
	return copyTime(t.assignmentDate)
}
func (t *DutyTask) CompletionDate() *time.Time {
	return copyTime(t.completionDate)
}
func (t *DutyTask) VerificationDate() *time.Time {
	return copyTime(t.verificationDate)
}

func (t *DutyTask) Status() DutyTaskStatus {
	switch {
	case t.verificationDate != nil:
		return DutyTaskStatusVerified
	case t.completionDate != nil:
		return DutyTaskStatusCompleted
	case t.assigneeID != nil:
		return DutyTaskStatusAssigned
	default:
		return DutyTaskStatusFree
	}
}

func (t *DutyTask) CanAssign(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrAssigneeRequired
	}
	if t.assigneeID != nil {
		return ErrTaskAssigned
	}
	return nil
}

func (t *DutyTask) CanUnassign(userID uuid.UUID) error {
	if err := t.requireAssignee(userID); err != nil {
		return err
	}
	if t.completionDate != nil {
		return ErrTaskAlreadyCompleted
	}
	return nil
}

func (t *DutyTask) CanComplete(userID uuid.UUID) error {
	if err := t.requireAssignee(userID); err != nil {
		return err
	}
	if t.completionDate != nil {
		return ErrTaskAlreadyCompleted
	}
	return nil
}

func (t *DutyTask) CanCancelCompletion(userID uuid.UUID) error {
	if err := t.requireAssignee(userID); err != nil {
		return err
	}
	return t.requirePendingVerification()
}

func (t *DutyTask) CanVerify(reviewerID uuid.UUID) error {
	if reviewerID == uuid.Nil {
		return ErrReviewerRequired
	}
	if t.assigneeID == nil {
		return ErrTaskNotAssigned
	}
	return t.requirePendingVerification()
}

func (t *DutyTask) CanReopen(reviewerID uuid.UUID) error {
	if reviewerID == uuid.Nil {
		return ErrReviewerRequired
	}
	if t.assigneeID == nil {
		return ErrTaskNotAssigned
	}
	return t.requirePendingVerification()
}

func (t *DutyTask) requireAssignee(userID uuid.UUID) error {
	if t.assigneeID == nil {
		return ErrTaskNotAssigned
	}
	if *t.assigneeID != userID {
		return ErrTaskOwnedByAnotherUser
	}
	return nil
}

func (t *DutyTask) requirePendingVerification() error {
	if t.completionDate == nil {
		return ErrTaskNotCompleted
	}
	if t.verificationDate != nil {
		return ErrTaskAlreadyVerified
	}
	return nil
}

func copyUUID(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

type DutyTaskRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*DutyTask, error)
	FindByDutyID(ctx context.Context, dutyID uuid.UUID) ([]*DutyTask, error)
	Create(ctx context.Context, task *DutyTask) error
	DeletePending(ctx context.Context, taskID uuid.UUID) error
	Assign(ctx context.Context, taskID uuid.UUID, assigneeID uuid.UUID, assignedAt time.Time) error
	Unassign(ctx context.Context, taskID uuid.UUID, assigneeID uuid.UUID) error
	Complete(ctx context.Context, taskID uuid.UUID, assigneeID uuid.UUID, completedAt time.Time) error
	CancelCompletion(ctx context.Context, taskID uuid.UUID, assigneeID uuid.UUID) error
	Verify(ctx context.Context, taskID uuid.UUID, reviewerID uuid.UUID, verifiedAt time.Time) error
	Reopen(ctx context.Context, taskID uuid.UUID, reviewerID uuid.UUID) error
}
