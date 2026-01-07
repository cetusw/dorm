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

func (r *DutyRepository) Save(ctx context.Context, d *duty.Duty) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const dutyQuery = `
		INSERT INTO duty (id, team_id, start_date, end_date)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			team_id = VALUES(team_id),
			start_date = VALUES(start_date),
			end_date = VALUES(end_date)
	`
	dIDBytes, _ := d.ID().MarshalBinary()
	tIDBytes, _ := d.TeamID().MarshalBinary()

	_, err = tx.ExecContext(ctx, dutyQuery, dIDBytes, tIDBytes, d.Start(), d.End())
	if err != nil {
		return fmt.Errorf("failed to save duty root: %w", err)
	}

	const taskQuery = `
		INSERT INTO duty_task (id, duty_id, task_id, assignee_id, completion_date)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			assignee_id = VALUES(assignee_id),
			completion_date = VALUES(completion_date)
	`

	for _, task := range d.Tasks() {
		dtIDBytes, _ := task.ID().MarshalBinary()
		defIDBytes, _ := task.TaskDefID().MarshalBinary()

		var assignee interface{}
		if task.AssigneeID() != nil {
			assignee, _ = task.AssigneeID().MarshalBinary()
		}

		var completion sql.NullTime
		if task.CompletionDate() != nil {
			completion.Valid = true
			completion.Time = *task.CompletionDate()
		}

		_, err = tx.ExecContext(ctx, taskQuery, dtIDBytes, dIDBytes, defIDBytes, assignee, completion)
		if err != nil {
			return fmt.Errorf("failed to save duty task %s: %w", task.ID(), err)
		}
	}

	return tx.Commit()
}

func (r *DutyRepository) FindCurrentByTeamID(ctx context.Context, teamID uuid.UUID) (*duty.Duty, error) {
	const dutyQuery = `
		SELECT id, team_id, start_date, end_date
		FROM duty
		WHERE team_id = ? AND start_date = (
			SELECT MAX(start_date) FROM duty
		)
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

func (r *DutyRepository) findTasksByDutyID(ctx context.Context, dutyID uuid.UUID) ([]*duty.DutyTask, error) {
	const query = `
		SELECT id, task_id, assignee_id, completion_date
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
		var dtID, defID, assignID []byte
		var compDate sql.NullTime

		if err := rows.Scan(&dtID, &defID, &assignID, &compDate); err != nil {
			return nil, err
		}

		id, _ := uuid.FromBytes(dtID)
		taskDefID, _ := uuid.FromBytes(defID)

		var assigneeUUID *uuid.UUID
		if len(assignID) > 0 {
			u, _ := uuid.FromBytes(assignID)
			assigneeUUID = &u
		}

		var completion *time.Time
		if compDate.Valid {
			completion = &compDate.Time
		}

		tasks = append(tasks, duty.RestoreDutyTask(id, taskDefID, assigneeUUID, completion))
	}
	return tasks, nil
}
