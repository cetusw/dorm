package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/duty"
)

type DutyTaskRepository struct {
	db *sql.DB
}

var _ duty.DutyTaskRepository = (*DutyTaskRepository)(nil)

func NewDutyTaskRepository(db *sql.DB) *DutyTaskRepository {
	return &DutyTaskRepository{db: db}
}

func (r *DutyTaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*duty.DutyTask, error) {
	const query = `
		SELECT id, duty_id, task_id, assignee_id, reviewer_id,
			assignment_date, completion_date, verification_date
		FROM duty_task
		WHERE id = ?
	`

	idBytes, err := id.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("find duty task: marshal id: %w", err)
	}

	task, err := scanDutyTask(r.db.QueryRowContext(ctx, query, idBytes))
	if err != nil {
		return nil, mapDutyTaskReadError(err)
	}
	return task, nil
}

func (r *DutyTaskRepository) FindByDutyID(ctx context.Context, dutyID uuid.UUID) ([]*duty.DutyTask, error) {
	return findDutyTasksByDutyID(ctx, r.db, dutyID)
}

func (r *DutyTaskRepository) Create(ctx context.Context, task *duty.DutyTask) error {
	const query = `
		INSERT INTO duty_task (
			id, duty_id, task_id, assignee_id, reviewer_id,
			assignment_date, completion_date, verification_date
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	taskID, taskDefID, err := marshalDutyTaskInsertIDs(task)
	if err != nil {
		return err
	}

	dutyID, err := marshalUUID(task.DutyID(), "duty id")
	if err != nil {
		return err
	}

	assigneeID, reviewerID, err := nullableDutyTaskUserIDs(task)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		taskID,
		dutyID,
		taskDefID,
		assigneeID,
		reviewerID,
		nullTime(task.AssignmentDate()),
		nullTime(task.CompletionDate()),
		nullTime(task.VerificationDate()),
	)
	if err != nil {
		if isDuplicateDutyTaskError(err) {
			return duty.ErrTaskAlreadyIncluded
		}
		return fmt.Errorf("create duty task: %w", err)
	}

	return nil
}

func (r *DutyTaskRepository) DeletePending(ctx context.Context, taskID uuid.UUID) error {
	const query = `
		DELETE FROM duty_task
		WHERE id = ? AND completion_date IS NULL
	`

	return r.executeTransition(ctx, query, duty.ErrTaskStateConflict, taskID)
}

func (r *DutyTaskRepository) Assign(ctx context.Context, taskID, assigneeID uuid.UUID, assignedAt time.Time) error {
	const query = `
		UPDATE duty_task
		SET assignee_id = ?, assignment_date = ?,
			completion_date = NULL, verification_date = NULL, reviewer_id = NULL
		WHERE id = ? AND assignee_id IS NULL
	`
	return r.executeTransition(ctx, query, duty.ErrTaskAssigned, assigneeID, assignedAt, taskID)
}

func (r *DutyTaskRepository) Unassign(ctx context.Context, taskID, assigneeID uuid.UUID) error {
	const query = `
		UPDATE duty_task
		SET assignee_id = NULL, assignment_date = NULL,
			completion_date = NULL, verification_date = NULL, reviewer_id = NULL
		WHERE id = ? AND assignee_id = ? AND completion_date IS NULL
	`
	return r.executeTransition(ctx, query, duty.ErrTaskStateConflict, taskID, assigneeID)
}

func (r *DutyTaskRepository) Complete(ctx context.Context, taskID, assigneeID uuid.UUID, completedAt time.Time) error {
	const query = `
		UPDATE duty_task
		SET completion_date = ?, verification_date = NULL, reviewer_id = NULL
		WHERE id = ? AND assignee_id = ? AND completion_date IS NULL
	`
	return r.executeTransition(ctx, query, duty.ErrTaskStateConflict, completedAt, taskID, assigneeID)
}

func (r *DutyTaskRepository) CancelCompletion(ctx context.Context, taskID, assigneeID uuid.UUID) error {
	const query = `
		UPDATE duty_task
		SET completion_date = NULL, verification_date = NULL, reviewer_id = NULL
		WHERE id = ? AND assignee_id = ?
		  AND completion_date IS NOT NULL AND verification_date IS NULL
	`
	return r.executeTransition(ctx, query, duty.ErrTaskStateConflict, taskID, assigneeID)
}

func (r *DutyTaskRepository) Verify(ctx context.Context, taskID, reviewerID uuid.UUID, verifiedAt time.Time) error {
	const query = `
		UPDATE duty_task
		SET reviewer_id = ?, verification_date = ?
		WHERE id = ? AND completion_date IS NOT NULL AND verification_date IS NULL
	`
	return r.executeTransition(ctx, query, duty.ErrTaskStateConflict, reviewerID, verifiedAt, taskID)
}

func (r *DutyTaskRepository) Reopen(ctx context.Context, taskID uuid.UUID, _ uuid.UUID) error {
	const query = `
		UPDATE duty_task
		SET completion_date = NULL, verification_date = NULL, reviewer_id = NULL
		WHERE id = ? AND completion_date IS NOT NULL AND verification_date IS NULL
	`
	return r.executeTransition(ctx, query, duty.ErrTaskStateConflict, taskID)
}

func (r *DutyTaskRepository) executeTransition(ctx context.Context, query string, conflict error, args ...any) error {
	sqlArgs, err := dutyTaskSQLArgs(args...)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, query, sqlArgs...)
	if err != nil {
		return fmt.Errorf("execute duty task transition: %w", err)
	}
	return requireOneAffectedRow(result, conflict)
}

func requireOneAffectedRow(result sql.Result, conflict error) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if affected == 0 {
		return conflict
	}
	if affected != 1 {
		return fmt.Errorf("unexpected affected rows: %d", affected)
	}
	return nil
}

func dutyTaskSQLArgs(args ...any) ([]any, error) {
	result := make([]any, 0, len(args))
	for _, arg := range args {
		value, err := dutyTaskSQLArg(arg)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

func dutyTaskSQLArg(arg any) (any, error) {
	id, ok := arg.(uuid.UUID)
	if !ok {
		return arg, nil
	}
	bytes, err := id.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal uuid: %w", err)
	}
	return bytes, nil
}

func mapDutyTaskReadError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return duty.ErrTaskNotFound
	}
	return fmt.Errorf("find duty task: %w", err)
}

func isDuplicateDutyTaskError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
