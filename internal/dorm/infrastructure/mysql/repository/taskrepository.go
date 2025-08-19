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

func (r *TaskRepository) Store(task *model.Task) error {
	query := "INSERT INTO task (task_id, area_id, title, cost) VALUES (UUID_TO_BIN(?), ?, ?, ?)"
	_, err := r.db.Exec(query, task.TaskID, task.AreaID, task.Title, task.Cost)
	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}
	return nil
}

func (r *TaskRepository) Find(taskID uuid.UUID) (*model.Task, error) {
	task := &model.Task{}
	query := "SELECT task_id, area_id, title, cost FROM task WHERE task_id = ?"

	err := r.db.QueryRow(query, taskID).Scan(&task.TaskID, &task.AreaID, &task.Title, &task.Cost)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

func (r *DutyRepository) FindAll() ([]model.Task, error) {
	query := "SELECT task_id, area_id, title, cost FROM task"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.TaskID, &task.AreaID, &task.Title, &task.Cost); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}
