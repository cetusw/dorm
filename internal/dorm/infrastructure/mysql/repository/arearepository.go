package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type AreaRepository struct {
	db *sql.DB
}

func NewAreaRepository(db *sql.DB) *AreaRepository {
	return &AreaRepository{db: db}
}

func (r *AreaRepository) Store(area *model.Area) error {
	query := "INSERT INTO area (area_floor, area_name) VALUES (?, ?)"
	_, err := r.db.Exec(query, area.Floor, area.Name)
	if err != nil {
		return fmt.Errorf("failed to save area: %w", err)
	}
	return nil
}

func (r *AreaRepository) Find(areaFloor int, areaName string) (*model.Area, error) {
	area := &model.Area{}
	query := "SELECT area_id, area_floor, area_name FROM area WHERE area_floor = ? AND area_name = ?"

	err := r.db.QueryRow(query, areaFloor, areaName).Scan(&area.AreaID, &area.Floor, &area.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get area: %w", err)
	}
	return area, nil
}

func (r *AreaRepository) FindAll() ([]model.Area, error) {
	query := `
		SELECT 
		    area_id, 
		    area_floor, 
		    area_name 
		FROM area 
		ORDER BY area_floor DESC, area_name DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query areas: %w", err)
	}
	defer rows.Close()

	var areas []model.Area

	for rows.Next() {
		var area model.Area
		if err := rows.Scan(
			&area.AreaID,
			&area.Floor,
			&area.Name,
		); err != nil {
			return nil, fmt.Errorf("failed to scan area row: %w", err)
		}
		areas = append(areas, area)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating area rows: %w", err)
	}

	return areas, nil
}
