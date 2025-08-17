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
	query := "INSERT INTO area (floor, name) VALUES (?, ?)"
	_, err := r.db.Exec(query, area.Floor, area.Name)
	if err != nil {
		return fmt.Errorf("failed to save area: %w", err)
	}
	return nil
}

func (r *AreaRepository) Find(areaFloor int, areaName string) (*model.Area, error) {
	area := &model.Area{}
	query := "SELECT area_id, floor, name FROM area WHERE floor = ? AND name = ?"

	err := r.db.QueryRow(query, areaFloor, areaName).Scan(&area.AreaID, &area.Floor, &area.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get area: %w", err)
	}
	return area, nil
}
