package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/catalog"
)

type CatalogRepository struct {
	db *sql.DB
}

func NewCatalogRepository(db *sql.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) GetAllAreas(ctx context.Context) ([]*catalog.Area, error) {
	const query = `SELECT id, name, floor FROM area ORDER BY floor DESC, name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []*catalog.Area
	for rows.Next() {
		var id, floor int
		var name string
		if err := rows.Scan(&id, &name, &floor); err != nil {
			return nil, err
		}
		areas = append(areas, catalog.RestoreArea(id, name, floor))
	}
	return areas, nil
}

func (r *CatalogRepository) GetAllTaskDefinitions(ctx context.Context) ([]*catalog.TaskDefinition, error) {
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
