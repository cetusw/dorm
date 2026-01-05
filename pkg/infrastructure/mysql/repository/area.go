package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/catalog"
)

type AreaRepository struct {
	db *sql.DB
}

func NewAreaRepository(db *sql.DB) *AreaRepository {
	return &AreaRepository{db: db}
}

func (r *AreaRepository) GetAllAreas(ctx context.Context) ([]*catalog.Area, error) {
	const query = `SELECT id, name, floor, group_id FROM area ORDER BY floor DESC, name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []*catalog.Area
	for rows.Next() {
		var id, floor int
		var name string
		var groupIDBytes []byte

		if err := rows.Scan(&id, &name, &floor, &groupIDBytes); err != nil {
			return nil, err
		}

		var groupID *uuid.UUID
		if len(groupIDBytes) > 0 {
			parsedID, _ := uuid.FromBytes(groupIDBytes)
			tempID := parsedID
			groupID = &tempID
		}

		areas = append(areas, catalog.RestoreArea(id, name, floor, groupID))
	}
	return areas, nil
}
