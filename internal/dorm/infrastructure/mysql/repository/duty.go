package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type DutyRepository struct {
	db *sql.DB
}

func NewDutyRepository(db *sql.DB) *DutyRepository {
	return &DutyRepository{db: db}
}

func (r *DutyRepository) Store(duty *model.Duty) error {
	query := "INSERT INTO duty (duty_id, team_id, duty_start_date, duty_end_date) VALUES (UUID_TO_BIN(?), ?, ?, ?)"
	_, err := r.db.Exec(query, duty.DutyID, duty.TeamID, duty.Start, duty.End)
	if err != nil {
		return fmt.Errorf("failed to save duty: %w", err)
	}
	return nil
}

func (r *DutyRepository) Find(dutyID uuid.UUID) (*model.Duty, error) {
	duty := &model.Duty{}
	query := "SELECT duty_id, team_id, duty_start_date, duty_end_date FROM duty WHERE duty_id = UUID_TO_BIN(?)"

	err := r.db.QueryRow(query, dutyID).Scan(&duty.DutyID, &duty.TeamID, &duty.Start, &duty.End)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get duty: %w", err)
	}
	return duty, nil
}

func (r *DutyRepository) FindLast() (*model.Duty, error) {
	duty := &model.Duty{}
	query := "SELECT duty_id, team_id, duty_start_date, duty_end_date FROM duty ORDER BY duty_start_date DESC LIMIT 1"

	err := r.db.QueryRow(query).Scan(&duty.DutyID, &duty.TeamID, &duty.Start, &duty.End)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get duty: %w", err)
	}

	return duty, nil
}
