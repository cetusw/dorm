package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/catalog"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) GetAllTaskDefinitions(ctx context.Context) ([]*catalog.TaskDefinition, error) {
	const query = `
		SELECT id, area_id, title, cost, recurrence_interval, start_sequence
		FROM task
		WHERE deleted_at IS NULL
	`
	return r.fetchTaskDefinitions(ctx, query)
}

func (r *TaskRepository) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*catalog.TaskDefinition, error) {
	const query = `
		SELECT t.id, t.area_id, t.title, t.cost, t.recurrence_interval, t.start_sequence
		FROM task t
		JOIN area a ON a.id = t.area_id
		WHERE a.group_id = ? AND t.deleted_at IS NULL
		ORDER BY a.floor DESC, a.name, t.title
	`
	groupIDBytes, err := groupID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("TaskRepository.FindByGroupID: marshal group id: %w", err)
	}
	return r.fetchTaskDefinitions(ctx, query, groupIDBytes)
}

func (r *TaskRepository) FindCommon(ctx context.Context) ([]*catalog.TaskDefinition, error) {
	const query = `
		SELECT t.id, t.area_id, t.title, t.cost, t.recurrence_interval, t.start_sequence
		FROM task t
		JOIN area a ON a.id = t.area_id
		WHERE a.group_id IS NULL AND t.deleted_at IS NULL
		ORDER BY a.floor DESC, a.name, t.title
	`
	return r.fetchTaskDefinitions(ctx, query)
}

func (r *TaskRepository) fetchTaskDefinitions(ctx context.Context, query string, args ...interface{}) ([]*catalog.TaskDefinition, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*catalog.TaskDefinition
	for rows.Next() {
		var idBytes []byte
		var areaID, cost, recurrenceInterval, startSequence int
		var title string

		if err := rows.Scan(&idBytes, &areaID, &title, &cost, &recurrenceInterval, &startSequence); err != nil {
			return nil, err
		}
		id, err := uuid.FromBytes(idBytes)
		if err != nil {
			return nil, fmt.Errorf("TaskRepository.fetchTaskDefinitions: parse task id: %w", err)
		}
		tasks = append(tasks, catalog.RestoreTaskDefinition(id, areaID, title, cost, recurrenceInterval, startSequence))
	}
	return tasks, rows.Err()
}

func (r *TaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*catalog.TaskDefinition, error) {
	const query = `
		SELECT id, area_id, title, cost, recurrence_interval, start_sequence
		FROM task
		WHERE id = ? AND deleted_at IS NULL
	`
	idBytes, err := id.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("TaskRepository.FindByID: marshal task id: %w", err)
	}
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var gotID []byte
	var areaID, cost, recurrenceInterval, startSequence int
	var title string
	if err := row.Scan(&gotID, &areaID, &title, &cost, &recurrenceInterval, &startSequence); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	uid, err := uuid.FromBytes(gotID)
	if err != nil {
		return nil, fmt.Errorf("TaskRepository.FindByID: parse task id: %w", err)
	}
	return catalog.RestoreTaskDefinition(uid, areaID, title, cost, recurrenceInterval, startSequence), nil
}

func (r *TaskRepository) FindLastCompletionDates(ctx context.Context, taskIDs []uuid.UUID) (map[uuid.UUID]*time.Time, error) {
	result := make(map[uuid.UUID]*time.Time, len(taskIDs))
	if len(taskIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(taskIDs))
	args := make([]interface{}, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		taskIDBytes, err := taskID.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("TaskRepository.FindLastCompletionDates: marshal task id: %w", err)
		}
		placeholders = append(placeholders, "?")
		args = append(args, taskIDBytes)
	}

	query := fmt.Sprintf(`
		SELECT dt.task_id, MAX(dt.completion_date)
		FROM duty_task dt
		WHERE dt.completion_date IS NOT NULL
		  AND dt.task_id IN (%s)
		GROUP BY dt.task_id
	`, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("TaskRepository.FindLastCompletionDates: query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskIDBytes []byte
		var completedAt time.Time
		if err := rows.Scan(&taskIDBytes, &completedAt); err != nil {
			return nil, fmt.Errorf("TaskRepository.FindLastCompletionDates: scan: %w", err)
		}

		taskID, err := uuid.FromBytes(taskIDBytes)
		if err != nil {
			return nil, fmt.Errorf("TaskRepository.FindLastCompletionDates: parse task id: %w", err)
		}

		completedAtCopy := completedAt
		result[taskID] = &completedAtCopy
	}

	return result, rows.Err()
}

func (r *TaskRepository) Save(ctx context.Context, task *catalog.TaskDefinition) error {
	const query = `
		INSERT INTO task (id, area_id, title, cost, recurrence_interval, start_sequence)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			area_id = VALUES(area_id),
			title = VALUES(title),
			cost = VALUES(cost),
			recurrence_interval = VALUES(recurrence_interval),
			start_sequence = VALUES(start_sequence),
			deleted_at = NULL
	`
	idBytes, err := task.ID().MarshalBinary()
	if err != nil {
		return fmt.Errorf("TaskRepository.Save: marshal task id: %w", err)
	}
	_, err = r.db.ExecContext(
		ctx,
		query,
		idBytes,
		task.AreaID(),
		task.Title(),
		task.Cost(),
		task.RecurrenceInterval(),
		task.StartSequence(),
	)
	if err != nil {
		return fmt.Errorf("TaskRepository.Save: %w", err)
	}
	return nil
}

func (r *TaskRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	idBytes, err := id.MarshalBinary()
	if err != nil {
		return fmt.Errorf("TaskRepository.SoftDelete: marshal task id: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("TaskRepository.SoftDelete: begin tx: %w", err)
	}

	if err := softDeleteTask(ctx, tx, idBytes); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("TaskRepository.SoftDelete: %w", err)
	}

	if err := deleteActiveDutyTasksByTaskID(ctx, tx, idBytes); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("TaskRepository.SoftDelete: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("TaskRepository.SoftDelete: commit: %w", err)
	}

	return nil
}

func softDeleteTask(ctx context.Context, tx *sql.Tx, taskID []byte) error {
	const query = `
		UPDATE task
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`

	if _, err := tx.ExecContext(ctx, query, taskID); err != nil {
		return fmt.Errorf("soft delete task: %w", err)
	}

	return nil
}

func deleteActiveDutyTasksByTaskID(ctx context.Context, tx *sql.Tx, taskID []byte) error {
	const query = `
		DELETE dt
		FROM duty_task dt
		INNER JOIN duty d ON d.id = dt.duty_id
		WHERE dt.task_id = ?
		  AND d.start_date <= NOW()
		  AND d.end_date > NOW()
	`

	if _, err := tx.ExecContext(ctx, query, taskID); err != nil {
		return fmt.Errorf("delete active duty tasks for removed task: %w", err)
	}

	return nil
}
