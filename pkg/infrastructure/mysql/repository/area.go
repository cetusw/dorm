package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func (r *AreaRepository) FindByID(ctx context.Context, id int) (*catalog.Area, error) {
	const query = `SELECT id, name, floor, group_id FROM area WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	var areaID int
	var name string
	var floor sql.NullInt64
	var groupIDBytes []byte

	if err := row.Scan(&areaID, &name, &floor, &groupIDBytes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("find area by id: %w", err)
	}

	return catalog.RestoreArea(areaID, name, int(floor.Int64), areaUUIDPtrFromBytes(groupIDBytes)), nil
}

func (r *AreaRepository) Save(ctx context.Context, area *catalog.Area) error {
	const query = `
		INSERT INTO area (id, name, floor, group_id)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			floor = VALUES(floor),
			group_id = VALUES(group_id)
	`

	var areaID interface{} = nil
	if area.ID() > 0 {
		areaID = area.ID()
	}

	var groupID interface{} = nil
	if area.GroupID() != nil {
		groupIDBytes, _ := area.GroupID().MarshalBinary()
		groupID = groupIDBytes
	}

	var floor interface{} = nil
	if area.Floor() != 0 {
		floor = area.Floor()
	}

	result, err := r.db.ExecContext(ctx, query, areaID, area.Name(), floor, groupID)
	if err != nil {
		return fmt.Errorf("save area: %w", err)
	}

	if area.ID() == 0 {
		insertedID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("load inserted area id: %w", err)
		}
		*area = *catalog.RestoreArea(int(insertedID), area.Name(), area.Floor(), area.GroupID())
	}

	return nil
}

func (r *AreaRepository) Delete(ctx context.Context, id int) error {
	const query = `DELETE FROM area WHERE id = ?`

	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete area: %w", err)
	}

	return nil
}

func areaUUIDPtrFromBytes(value []byte) *uuid.UUID {
	if len(value) == 0 {
		return nil
	}

	parsedID, _ := uuid.FromBytes(value)
	tempID := parsedID
	return &tempID
}
