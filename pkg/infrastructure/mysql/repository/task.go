package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	const query = `SELECT id, area_id, title, cost, frequency FROM task`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*catalog.TaskDefinition
	for rows.Next() {
		var idBytes []byte
		var areaID, cost, freq int
		var title string

		if err := rows.Scan(&idBytes, &areaID, &title, &cost, &freq); err != nil {
			return nil, err
		}
		id, _ := uuid.FromBytes(idBytes)
		tasks = append(tasks, catalog.RestoreTaskDefinition(id, areaID, title, cost, freq))
	}
	return tasks, nil
}

func (r *TaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*catalog.TaskDefinition, error) {
	const query = `SELECT id, area_id, title, cost, frequency FROM task WHERE id = ?`
	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	var gotID []byte
	var areaID, cost, frequency int
	var title string
	if err := row.Scan(&gotID, &areaID, &title, &cost, &frequency); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	uid, _ := uuid.FromBytes(gotID)
	return catalog.RestoreTaskDefinition(uid, areaID, title, cost, frequency), nil
}

func (r *TaskRepository) Save(ctx context.Context, task *catalog.TaskDefinition) error {
	const query = `
		INSERT INTO task (id, area_id, title, cost, frequency)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			area_id = VALUES(area_id),
			title = VALUES(title),
			cost = VALUES(cost),
			frequency = VALUES(frequency)
	`
	idBytes, _ := task.ID().MarshalBinary()
	_, err := r.db.ExecContext(ctx, query, idBytes, task.AreaID(), task.Title(), task.Cost(), task.Frequency())
	if err != nil {
		return fmt.Errorf("TaskRepository.Save: %w", err)
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
