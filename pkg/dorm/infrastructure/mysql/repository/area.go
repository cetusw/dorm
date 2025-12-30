package repository

import (
	"database/sql"
	"dorm/pkg/dorm/application/model"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type AreaRepository struct {
	db *sql.DB
}

func NewAreaRepository(db *sql.DB) *AreaRepository {
	return &AreaRepository{db: db}
}

func (r *AreaRepository) Find(areaFloor int, areaName string) (*model.Area, error) {
	const sqlQuery = `
		SELECT id, floor, name 
		FROM area 
		WHERE floor = ? AND name = ?`

	area := &model.Area{}
	err := r.db.QueryRow(sqlQuery, areaFloor, areaName).Scan(&area.ID, &area.Floor, &area.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get area: %w", err)
	}
	return area, nil
}

func (r *AreaRepository) FindAll() ([]model.Area, error) {
	const sqlQuery = `
		SELECT id, floor, name 
		FROM area 
		ORDER BY floor DESC, name DESC`

	rows, err := r.db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query areas: %w", err)
	}
	defer rows.Close()

	var areas []model.Area

	for rows.Next() {
		var area model.Area
		if err := rows.Scan(&area.ID, &area.Floor, &area.Name); err != nil {
			return nil, fmt.Errorf("failed to scan area row: %w", err)
		}
		areas = append(areas, area)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating area rows: %w", err)
	}

	return areas, nil
}

func (r *AreaRepository) FindUnassignedAreasByDutyID(dutyID uuid.UUID) ([]model.Area, error) {
	const sqlQuery = `
		SELECT DISTINCT a.id, a.floor, a.name
		FROM area a
			INNER JOIN task t ON t.area_id = a.id
		    INNER JOIN duty_task dt ON dt.task_id = t.id
		WHERE dt.duty_id = UUID_TO_BIN(?)
		  AND dt.assignee_id IS NULL
		ORDER BY a.floor DESC, a.name DESC`

	rows, err := r.db.Query(sqlQuery, dutyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unassigned areas for duty %s: %w", dutyID, err)
	}
	defer rows.Close()

	var areas []model.Area
	for rows.Next() {
		var area model.Area
		if err := rows.Scan(&area.ID, &area.Floor, &area.Name); err != nil {
			return nil, fmt.Errorf("failed to scan area row: %w", err)
		}
		areas = append(areas, area)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unassigned areas rows: %w", err)
	}
	return areas, nil
}

func (r *AreaRepository) FindAreasWithoutGroup() ([]model.Area, error) {
	const sqlQuery = `
		SELECT id, floor, name, group_id
		FROM area 
		WHERE group_id IS NULL
		ORDER BY floor DESC, name DESC`

	rows, err := r.db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query areas by scope: %w", err)
	}
	defer rows.Close()

	var areas []model.Area

	for rows.Next() {
		var area model.Area
		if err := rows.Scan(&area.ID, &area.Floor, &area.Name, &area.GroupID); err != nil {
			return nil, fmt.Errorf("failed to scan area row: %w", err)
		}
		areas = append(areas, area)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating area rows: %w", err)
	}

	return areas, nil
}

func (r *AreaRepository) Store(area *model.Area) error {
	const sqlQuery = `
		INSERT INTO area (floor, name) 
		VALUES (?, ?)`
	_, err := r.db.Exec(sqlQuery, area.Floor, area.Name)
	if err != nil {
		return fmt.Errorf("failed to save area: %w", err)
	}
	return nil
}
