package repository

import (
	"database/sql"
	"dorm/pkg/dorm/application/model"
	"errors"
	"fmt"
	"strings"

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
		SELECT id, team_id, start_date, end_date 
		FROM duty 
		WHERE id = UUID_TO_BIN(?)`

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

func (r *DutyRepository) FindLastDutyByGroupID(groupID uuid.UUID) (*model.Duty, error) {
	const sqlQuery = `
		SELECT d.id, d.team_id, d.start_date, d.end_date 
		FROM duty d
		    INNER JOIN team t ON d.team_id = t.id
		WHERE t.group_id = UUID_TO_BIN(?)
		ORDER BY d.start_date DESC 
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

func (r *DutyRepository) FindLastDutyByUserID(userID uuid.UUID) (*model.Duty, error) {
	const sqlQuery = `
		SELECT d.id, d.team_id, d.start_date, d.end_date 
		FROM duty d
		    INNER JOIN team t ON d.team_id = t.id
			INNER JOIN user u ON u.team_id = t.id
		WHERE u.id = UUID_TO_BIN(?)
		ORDER BY d.start_date DESC 
		LIMIT 1`

	duty := &model.Duty{}
	err := r.db.QueryRow(sqlQuery, userID).Scan(&duty.DutyID, &duty.TeamID, &duty.Start, &duty.End)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last duty for team group %s: %w", userID, err)
	}

	return duty, nil
}

func (r *DutyRepository) FindLastDutyByTaskID(taskID uuid.UUID) (*model.Duty, error) {
	const sqlQuery = `
		SELECT d.id, d.team_id, d.start_date, d.end_date 
		FROM duty d
		    JOIN duty_task dt ON d.id = dt.duty_id
		WHERE dt.task_id = UUID_TO_BIN(?)
		ORDER BY d.start_date DESC 
		LIMIT 1`

	duty := &model.Duty{}
	err := r.db.QueryRow(sqlQuery, taskID).Scan(&duty.DutyID, &duty.TeamID, &duty.Start, &duty.End)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last duty where task %s existed: %w", taskID, err)
	}

	return duty, nil
}

func (r *DutyRepository) Store(duty *model.Duty) error {
	const sqlQuery = `
		INSERT INTO duty (id, team_id, start_date, end_date) 
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
		INSERT INTO duty (id, team_id, start_date, end_date) 
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
		SELECT COUNT(DISTINCT start_date)
		FROM duty`

	var count int
	err := r.db.QueryRow(sqlQuery).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count distinct start dates: %w", err)
	}

	return count, nil
}
