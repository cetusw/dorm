package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	const query = `SELECT id, area_id, title, cost, frequency, is_active FROM task`
	return r.fetchTaskDefinitions(ctx, query)
}

func (r *TaskRepository) GetActiveTaskDefinitions(ctx context.Context) ([]*catalog.TaskDefinition, error) {
	const query = `SELECT id, area_id, title, cost, frequency, is_active FROM task WHERE is_active = TRUE`
	return r.fetchTaskDefinitions(ctx, query)
}

func (r *TaskRepository) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*catalog.TaskDefinition, error) {
	const query = `
		SELECT t.id, t.area_id, t.title, t.cost, t.frequency, t.is_active
		FROM task t
		JOIN area a ON a.id = t.area_id
		WHERE a.group_id = ?
		ORDER BY a.floor DESC, a.name, t.title
	`
	groupIDBytes, _ := groupID.MarshalBinary()
	return r.fetchTaskDefinitions(ctx, query, groupIDBytes)
}

func (r *TaskRepository) FindCommon(ctx context.Context) ([]*catalog.TaskDefinition, error) {
	const query = `
		SELECT t.id, t.area_id, t.title, t.cost, t.frequency, t.is_active
		FROM task t
		JOIN area a ON a.id = t.area_id
		WHERE a.group_id IS NULL
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
		var areaID, cost, freq int
		var title string
		var isActive bool

		if err := rows.Scan(&idBytes, &areaID, &title, &cost, &freq, &isActive); err != nil {
			return nil, err
		}
		id, _ := uuid.FromBytes(idBytes)
		tasks = append(tasks, catalog.RestoreTaskDefinitionWithActive(id, areaID, title, cost, freq, isActive))
	}
	return tasks, rows.Err()
}

func (r *TaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*catalog.TaskDefinition, error) {
	const query = `SELECT id, area_id, title, cost, frequency, is_active FROM task WHERE id = ?`
	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var gotID []byte
	var areaID, cost, frequency int
	var title string
	var isActive bool
	if err := row.Scan(&gotID, &areaID, &title, &cost, &frequency, &isActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	uid, _ := uuid.FromBytes(gotID)
	return catalog.RestoreTaskDefinitionWithActive(uid, areaID, title, cost, frequency, isActive), nil
}

func (r *TaskRepository) Save(ctx context.Context, task *catalog.TaskDefinition) error {
	const query = `
		INSERT INTO task (id, area_id, title, cost, frequency, is_active)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			area_id = VALUES(area_id),
			title = VALUES(title),
			cost = VALUES(cost),
			frequency = VALUES(frequency),
			is_active = VALUES(is_active)
	`
	idBytes, _ := task.ID().MarshalBinary()
	_, err := r.db.ExecContext(ctx, query, idBytes, task.AreaID(), task.Title(), task.Cost(), task.Frequency(), task.IsActive())
	if err != nil {
		return fmt.Errorf("TaskRepository.Save: %w", err)
	}
	return nil
}

func (r *TaskRepository) UpdateGroupTaskActivity(ctx context.Context, groupID uuid.UUID, activeTaskIDs []uuid.UUID) error {
	args := make([]interface{}, 0, len(activeTaskIDs)+1)
	var setExpression string
	if len(activeTaskIDs) == 0 {
		setExpression = "FALSE"
	} else {
		placeholders := make([]string, 0, len(activeTaskIDs))
		for _, id := range activeTaskIDs {
			idBytes, _ := id.MarshalBinary()
			args = append(args, idBytes)
			placeholders = append(placeholders, "?")
		}
		setExpression = fmt.Sprintf("t.id IN (%s)", strings.Join(placeholders, ", "))
	}

	groupIDBytes, _ := groupID.MarshalBinary()
	args = append(args, groupIDBytes)
	query := fmt.Sprintf(`
		UPDATE task t
		JOIN area a ON a.id = t.area_id
		SET t.is_active = %s
		WHERE a.group_id = ?
	`, setExpression)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("TaskRepository.UpdateGroupTaskActivity: %w", err)
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM task WHERE id = ?`
	idBytes, _ := id.MarshalBinary()
	_, err := r.db.ExecContext(ctx, query, idBytes)
	if err != nil {
		return fmt.Errorf("TaskRepository.Delete: %w", err)
	}
	return nil
}
