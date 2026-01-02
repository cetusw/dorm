package repository

import (
	"context"
	"database/sql"

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
