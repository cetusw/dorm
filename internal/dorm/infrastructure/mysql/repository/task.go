package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Find(taskID uuid.UUID) (*model.Task, error) {
	const sqlQuery = `
		SELECT id, area_id, title, cost, frequency 
		FROM task WHERE id = UUID_TO_BIN(?)`

	task := &model.Task{}
	err := r.db.QueryRow(sqlQuery, taskID).Scan(&task.TaskID, &task.AreaID, &task.Title, &task.Cost, &task.Frequency, &task.Scope)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

func (r *TaskRepository) FindAll() ([]model.Task, error) {
	const sqlQuery = `
		SELECT id, area_id, title, cost, frequency
		FROM task`

	rows, err := r.db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.TaskID, &task.AreaID, &task.Title, &task.Cost, &task.Frequency, &task.Scope); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) FindGroupTasks(groupID uuid.UUID) ([]model.Task, error) {
	const sqlQuery = `
		SELECT t.id, t.area_id, t.title, t.cost, t.frequency
		FROM task t
			INNER JOIN area a ON a.id = t.area_id
		WHERE a.group_id = UUID_TO_BIN(?)`

	rows, err := r.db.Query(sqlQuery, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by groupID %s: %w", groupID, err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.TaskID, &task.AreaID, &task.Title, &task.Cost, &task.Frequency); err != nil {
			return nil, fmt.Errorf("failed to scan task ID row: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task ID rows: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) FindPublicTasks() ([]model.Task, error) {
	const sqlQuery = `
		SELECT t.id, t.area_id, t.title, t.cost, t.frequency
		FROM task t
			INNER JOIN area a ON a.id = t.area_id
		WHERE a.group_id IS NULL`

	rows, err := r.db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query public tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.TaskID, &task.AreaID, &task.Title, &task.Cost, &task.Frequency); err != nil {
			return nil, fmt.Errorf("failed to scan task ID row: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task ID rows: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) Store(task *model.Task) error {
	const sqlQuery = `
		INSERT INTO task (id, area_id, title, cost, frequency) 
		VALUES (UUID_TO_BIN(?), ?, ?, ?, ?)`

	_, err := r.db.Exec(sqlQuery, task.TaskID, task.AreaID, task.Title, task.Cost, task.Frequency, task.Scope)
	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}
	return nil
}
