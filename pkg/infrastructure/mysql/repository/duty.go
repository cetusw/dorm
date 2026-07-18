package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/duty"
)

type DutyRepository struct {
	db *sql.DB
}

func NewDutyRepository(db *sql.DB) *DutyRepository {
	return &DutyRepository{db: db}
}

func (r *DutyRepository) CreateWithTasks(
	ctx context.Context,
	currentDuty *duty.Duty,
	tasks []*duty.DutyTask,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin duty transaction: %w", err)
	}

	if err := r.createDutyData(ctx, tx, currentDuty, tasks); err != nil {
		_ = tx.Rollback()
		return err
	}

	return commitDutyTransaction(tx)
}

func (r *DutyRepository) createDutyData(
	ctx context.Context,
	tx *sql.Tx,
	currentDuty *duty.Duty,
	tasks []*duty.DutyTask,
) error {
	if err := insertDuty(ctx, tx, currentDuty); err != nil {
		return err
	}
	return insertDutyTasks(ctx, tx, currentDuty.ID(), tasks)
}

func insertDuty(ctx context.Context, tx *sql.Tx, d *duty.Duty) error {
	const dutyQuery = `
		INSERT INTO duty (id, team_id, start_date, end_date)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			team_id = VALUES(team_id),
			start_date = VALUES(start_date),
			end_date = VALUES(end_date)
	`
	dIDBytes, err := marshalUUID(d.ID(), "duty id")
	if err != nil {
		return err
	}
	tIDBytes, err := marshalUUID(d.TeamID(), "team id")
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, dutyQuery, dIDBytes, tIDBytes, d.Start(), d.End())
	if err != nil {
		return fmt.Errorf("failed to save duty root: %w", err)
	}
	return nil
}

func insertDutyTasks(
	ctx context.Context,
	tx *sql.Tx,
	dutyID uuid.UUID,
	tasks []*duty.DutyTask,
) error {
	const taskQuery = `
		INSERT INTO duty_task (
			id, duty_id, task_id, assignee_id, reviewer_id,
			assignment_date, completion_date, verification_date
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			assignee_id = VALUES(assignee_id),
			reviewer_id = VALUES(reviewer_id),
			assignment_date = VALUES(assignment_date),
			completion_date = VALUES(completion_date),
			verification_date = VALUES(verification_date)
	`

	dIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if err := insertDutyTask(ctx, tx, taskQuery, dIDBytes, task); err != nil {
			return err
		}
	}
	return nil
}

func insertDutyTask(ctx context.Context, tx *sql.Tx, query string, dutyID []byte, task *duty.DutyTask) error {
	taskID, taskDefID, err := marshalDutyTaskInsertIDs(task)
	if err != nil {
		return err
	}
	assigneeID, reviewerID, err := nullableDutyTaskUserIDs(task)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, query, taskID, dutyID, taskDefID,
		assigneeID, reviewerID,
		nullTime(task.AssignmentDate()), nullTime(task.CompletionDate()),
		nullTime(task.VerificationDate()))
	if err != nil {
		return fmt.Errorf("failed to save duty task %s: %w", task.ID(), err)
	}
	return nil
}

func marshalDutyTaskInsertIDs(task *duty.DutyTask) ([]byte, []byte, error) {
	taskID, err := marshalUUID(task.ID(), "duty task id")
	if err != nil {
		return nil, nil, err
	}
	taskDefID, err := marshalUUID(task.TaskDefID(), "task definition id")
	return taskID, taskDefID, err
}

func nullableDutyTaskUserIDs(task *duty.DutyTask) (any, any, error) {
	assigneeID, err := nullableUUIDBytes(task.AssigneeID(), "assignee id")
	if err != nil {
		return nil, nil, err
	}
	reviewerID, err := nullableUUIDBytes(task.ReviewerID(), "reviewer id")
	return assigneeID, reviewerID, err
}

func commitDutyTransaction(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit duty transaction: %w", err)
	}
	return nil
}

func (r *DutyRepository) FindCurrentByTeamID(ctx context.Context, teamID uuid.UUID) (*duty.Duty, error) {
	const dutyQuery = `
		SELECT id, team_id, start_date, end_date
		FROM duty
		WHERE team_id = ? AND start_date = (
			SELECT MAX(start_date) FROM duty WHERE team_id = ?
		)
	`
	tIDBytes, _ := teamID.MarshalBinary()
	row := r.db.QueryRowContext(ctx, dutyQuery, tIDBytes, tIDBytes)

	var dID, teamIDBytes []byte
	var start, end time.Time

	err := row.Scan(&dID, &teamIDBytes, &start, &end)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find duty root: %w", err)
	}

	dutyID, _ := uuid.FromBytes(dID)

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	return duty.RestoreDuty(dutyID, teamID, start, end, tasks), nil
}

func (r *DutyRepository) FindActiveByTeamID(
	ctx context.Context,
	teamID uuid.UUID,
	at time.Time,
) (*duty.Duty, error) {
	const query = `
		SELECT id, team_id, start_date, end_date
		FROM duty
		WHERE team_id = ?
		  AND start_date <= ?
		  AND end_date > ?
		ORDER BY start_date DESC
		LIMIT 1
	`

	teamIDBytes, err := teamID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("find active duty: marshal team id: %w", err)
	}

	row := r.db.QueryRowContext(ctx, query, teamIDBytes, at, at)

	var dutyIDBytes, teamIDResultBytes []byte
	var startDate, endDate time.Time

	if err := row.Scan(&dutyIDBytes, &teamIDResultBytes, &startDate, &endDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find active duty: %w", err)
	}

	dutyID, err := uuid.FromBytes(dutyIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find active duty: parse duty id: %w", err)
	}

	teamIDResult, err := uuid.FromBytes(teamIDResultBytes)
	if err != nil {
		return nil, fmt.Errorf("find active duty: parse team id: %w", err)
	}

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	return duty.RestoreDuty(dutyID, teamIDResult, startDate, endDate, tasks), nil
}

func (r *DutyRepository) FindLatestByTeamID(ctx context.Context, teamID uuid.UUID) (*duty.Duty, error) {
	const dutyQuery = `
		SELECT id, team_id, start_date, end_date
		FROM duty
		WHERE team_id = ?
		ORDER BY start_date DESC
		LIMIT 1
	`
	tIDBytes, _ := teamID.MarshalBinary()
	row := r.db.QueryRowContext(ctx, dutyQuery, tIDBytes)

	var dID, teamIDBytes []byte
	var start, end time.Time

	err := row.Scan(&dID, &teamIDBytes, &start, &end)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find duty root: %w", err)
	}

	dutyID, _ := uuid.FromBytes(dID)

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	return duty.RestoreDuty(dutyID, teamID, start, end, tasks), nil
}

func (r *DutyRepository) FindByID(ctx context.Context, id uuid.UUID) (*duty.Duty, error) {
	const dutyQuery = `SELECT id, team_id, start_date, end_date FROM duty WHERE id = ?`

	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, dutyQuery, idBytes)

	var dID, tID []byte
	var start, end time.Time

	err := row.Scan(&dID, &tID, &start, &end)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	dutyID, _ := uuid.FromBytes(dID)
	teamID, _ := uuid.FromBytes(tID)

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	return duty.RestoreDuty(dutyID, teamID, start, end, tasks), nil
}

func (r *DutyRepository) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*duty.Duty, error) {
	const query = `
		SELECT d.id, d.team_id, d.start_date, d.end_date
		FROM duty d
		JOIN team t ON t.id = d.team_id
		WHERE t.group_id = ?
		ORDER BY d.start_date DESC, d.end_date DESC
	`

	groupIDBytes, _ := groupID.MarshalBinary()
	rows, err := r.db.QueryContext(ctx, query, groupIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find duties by group: %w", err)
	}
	defer rows.Close()

	var result []*duty.Duty
	for rows.Next() {
		var dID, tID []byte
		var start, end time.Time
		if err := rows.Scan(&dID, &tID, &start, &end); err != nil {
			return nil, err
		}
		id, _ := uuid.FromBytes(dID)
		teamID, _ := uuid.FromBytes(tID)

		tasks, err := r.findTasksByDutyID(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, duty.RestoreDuty(id, teamID, start, end, tasks))
	}
	return result, rows.Err()
}

func (r *DutyRepository) CountDistinctStartDates(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(DISTINCT start_date) FROM duty`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count distinct start dates: %w", err)
	}

	return count, nil
}

func (r *DutyRepository) FindLastByTaskDefID(ctx context.Context, taskDefID uuid.UUID) (*duty.Duty, error) {
	const query = `
		SELECT d.id, d.team_id, d.start_date, d.end_date
		FROM duty d
		JOIN duty_task dt ON d.id = dt.duty_id
		WHERE dt.task_id = ?
		ORDER BY d.start_date DESC
		LIMIT 1
	`

	defIDBytes, _ := taskDefID.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, defIDBytes)

	var dID, tID []byte
	var start, end time.Time

	err := row.Scan(&dID, &tID, &start, &end)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find last by task def: %w", err)
	}

	dutyID, _ := uuid.FromBytes(dID)
	teamID, _ := uuid.FromBytes(tID)

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	return duty.RestoreDuty(dutyID, teamID, start, end, tasks), nil
}

func (r *DutyRepository) FindAllLatest(ctx context.Context) ([]*duty.Duty, error) {
	const query = `
		SELECT id, team_id, start_date, end_date 
		FROM duty 
		WHERE start_date = (SELECT MAX(start_date) FROM duty)`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all latest duties: %w", err)
	}
	defer rows.Close()

	var result []*duty.Duty
	for rows.Next() {
		var dID, tID []byte
		var start, end time.Time
		if err := rows.Scan(&dID, &tID, &start, &end); err != nil {
			return nil, err
		}
		id, _ := uuid.FromBytes(dID)
		teamID, _ := uuid.FromBytes(tID)

		tasks, err := r.findTasksByDutyID(ctx, id)
		if err != nil {
			return nil, err
		}

		result = append(result, duty.RestoreDuty(id, teamID, start, end, tasks))
	}
	return result, rows.Err()
}

func (r *DutyRepository) findTasksByDutyID(ctx context.Context, dutyID uuid.UUID) ([]*duty.DutyTask, error) {
	const query = `
		SELECT id, duty_id, task_id, assignee_id, reviewer_id,
			assignment_date, completion_date, verification_date
		FROM duty_task
		WHERE duty_id = ?
	`
	dIDBytes, _ := dutyID.MarshalBinary()
	rows, err := r.db.QueryContext(ctx, query, dIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*duty.DutyTask
	for rows.Next() {
		task, err := scanDutyTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
