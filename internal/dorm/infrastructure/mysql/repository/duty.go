package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type DutyRepository struct {
	db *sql.DB
}

func NewDutyRepository(db *sql.DB) *DutyRepository {
	return &DutyRepository{db: db}
}

func (r *DutyRepository) Find(dutyID uuid.UUID) (*model.Duty, error) {
	const sqlQuery = `
		SELECT duty_id, team_id, duty_start_date, duty_end_date 
		FROM duty 
		WHERE duty_id = UUID_TO_BIN(?)`

	duty := &model.Duty{}
	err := r.db.QueryRow(sqlQuery, dutyID).Scan(&duty.DutyID, &duty.TeamID, &duty.Start, &duty.End)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get duty: %w", err)
	}
	return duty, nil
}

func (r *DutyRepository) FindDutyByTeamIDAndStartDate(teamID uuid.UUID, startDate time.Time) (*model.Duty, error) {
	const sqlQuery = `
		SELECT duty_id, team_id, duty_start_date, duty_end_date 
		FROM duty
		WHERE team_id = UUID_TO_BIN(?) 
		  AND DATE(duty_start_date) = ?
		ORDER BY duty_start_date DESC
		LIMIT 1`

	duty := &model.Duty{}
	dateString := startDate.Format("2006-01-02")
	err := r.db.QueryRow(sqlQuery, teamID, dateString).Scan(&duty.DutyID, &duty.TeamID, &duty.Start, &duty.End)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get duty: %w", err)
	}
	return duty, nil
}

func (r *DutyRepository) FindLastDuty() (*model.Duty, error) {
	const sqlQuery = `
		SELECT duty_id, team_id, duty_start_date, duty_end_date 
		FROM duty 
		ORDER BY duty_start_date DESC 
		LIMIT 1`

	duty := &model.Duty{}
	err := r.db.QueryRow(sqlQuery).Scan(&duty.DutyID, &duty.TeamID, &duty.Start, &duty.End)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get duty: %w", err)
	}

	return duty, nil
}

func (r *DutyRepository) FindLastDutyByGroupID(groupID uuid.UUID) (*model.Duty, error) {
	const sqlQuery = `
		SELECT d.duty_id, d.team_id, d.duty_start_date, d.duty_end_date 
		FROM duty d
		    INNER JOIN team t ON d.team_id = t.team_id
		WHERE t.group_id = UUID_TO_BIN(?)
		ORDER BY d.duty_start_date DESC 
		LIMIT 1`

	duty := &model.Duty{}
	err := r.db.QueryRow(sqlQuery, groupID).Scan(&duty.DutyID, &duty.TeamID, &duty.Start, &duty.End)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last duty for team group %s: %w", groupID, err)
	}

	return duty, nil
}

func (r *DutyRepository) Store(duty *model.Duty) error {
	const sqlQuery = `
		INSERT INTO duty (duty_id, team_id, duty_start_date, duty_end_date) 
		VALUES (UUID_TO_BIN(?), ?, ?, ?)`
	_, err := r.db.Exec(sqlQuery, duty.DutyID, duty.TeamID, duty.Start, duty.End)
	if err != nil {
		return fmt.Errorf("failed to save duty: %w", err)
	}
	return nil
}

func (r *DutyRepository) StoreBatch(duties []model.Duty) error {
	if len(duties) == 0 {
		return nil
	}

	const sqlQuery = `
		INSERT INTO duty (duty_id, team_id, duty_start_date, duty_end_date) 
		VALUES `

	var valueStrings []string
	var valueArgs []interface{}

	for _, duty := range duties {
		valueStrings = append(valueStrings, "(UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?)")
		valueArgs = append(valueArgs, duty.DutyID, duty.TeamID, duty.Start, duty.End)
	}

	stmt := fmt.Sprintf("%s %s", sqlQuery, strings.Join(valueStrings, ","))

	_, err := r.db.Exec(stmt, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch insert duty: %w", err)
	}

	return nil
}

func (r *DutyRepository) CountDistinctStartDates() (int, error) {
	const sqlQuery = `
		SELECT COUNT(DISTINCT duty_start_date)
		FROM duty`

	var count int
	err := r.db.QueryRow(sqlQuery).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count distinct start dates: %w", err)
	}

	return count, nil
}
