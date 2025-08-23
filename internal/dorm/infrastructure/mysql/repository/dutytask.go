package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type DutyTaskRepository struct {
	db *sql.DB
}

func NewDutyTaskRepository(db *sql.DB) *DutyTaskRepository {
	return &DutyTaskRepository{db: db}
}

func (r *DutyTaskRepository) Store(dutyTask *model.DutyTask) error {
	query := `
	INSERT INTO duty_task (duty_task_id, 
	                       duty_id, 
	                       task_id, 
	                       reviewer_id) 
	VALUES (UUID_TO_BIN(?), 
	        UUID_TO_BIN(?), 
	        UUID_TO_BIN(?), 
	        UUID_TO_BIN(?))`
	_, err := r.db.Exec(query, dutyTask.DutyTaskID, dutyTask.DutyID, dutyTask.TaskID, dutyTask.ReviewerID)
	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}
	return nil
}

func (r *DutyTaskRepository) StoreBatch(dutyTasks []model.DutyTask) error {
	if len(dutyTasks) == 0 {
		return nil
	}

	query := "INSERT INTO duty_task (duty_task_id, duty_id, task_id, reviewer_id) VALUES "

	var valueStrings []string
	var valueArgs []interface{}

	for _, dutyTask := range dutyTasks {
		valueStrings = append(valueStrings, "(UUID_TO_BIN(?), UUID_TO_BIN(?), UUID_TO_BIN(?), UUID_TO_BIN(?))")
		valueArgs = append(valueArgs, dutyTask.DutyTaskID, dutyTask.DutyID, dutyTask.TaskID, dutyTask.ReviewerID)
	}

	stmt := fmt.Sprintf("%s %s", query, strings.Join(valueStrings, ","))

	_, err := r.db.Exec(stmt, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch insert duty_tasks: %w", err)
	}

	return nil
}

func (r *DutyTaskRepository) FindDutyTasksReadableByDutyID(
	dutyID uuid.UUID,
) ([]model.DutyTaskView, error) {
	query := `
		SELECT 
			a.area_floor, 
			a.area_name, 
			t.task_id,
			t.task_title, 
			t.task_cost, 
			u.user_id,
			u.first_name, 
			u.last_name, 
			dt.completion_date, 
			dt.verification_date 
		FROM duty_task dt 
			INNER JOIN task t ON dt.task_id = t.task_id 
			INNER JOIN area a ON a.area_id = t.area_id 
			LEFT JOIN user u ON u.user_id = dt.assignee_id 
		WHERE dt.duty_id = UUID_TO_BIN(?)`

	rows, err := r.db.Query(query, dutyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query readable tasks for duty %s: %w", dutyID, err)
	}
	defer rows.Close()

	var dutyTasksReadable []model.DutyTaskView

	for rows.Next() {
		var dutyTaskReadable model.DutyTaskView
		if err := rows.Scan(
			&dutyTaskReadable.AreaFloor,
			&dutyTaskReadable.AreaName,
			&dutyTaskReadable.TaskID,
			&dutyTaskReadable.TaskTitle,
			&dutyTaskReadable.TaskCost,
			&dutyTaskReadable.AssigneeID,
			&dutyTaskReadable.AssigneeFirstName,
			&dutyTaskReadable.AssigneeLastName,
			&dutyTaskReadable.CompletionDate,
			&dutyTaskReadable.VerificationDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan readable task row: %w", err)
		}

		dutyTasksReadable = append(dutyTasksReadable, dutyTaskReadable)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating readable task rows: %w", err)
	}

	return dutyTasksReadable, nil
}

// TODO: придумать, что делать, если дежурство прошло, но человек всё равно хочет увидеть задачи за прошлую неделю, которые он не выполнил
func (r *DutyTaskRepository) FindUncompletedDutyTasksView(
	assigneeID uuid.UUID,
	dutyID uuid.UUID,
) ([]model.DutyTaskView, error) {
	query := `
		SELECT 
			a.area_floor, 
			a.area_name, 
			t.task_id,
			t.task_title, 
			t.task_cost, 
			u.first_name, 
			u.last_name, 
			dt.completion_date, 
			dt.verification_date 
		FROM duty_task dt 
			INNER JOIN task t ON dt.task_id = t.task_id 
			INNER JOIN area a ON a.area_id = t.area_id
			LEFT JOIN user u ON u.user_id = dt.assignee_id
		WHERE dt.completion_date IS NULL 
		  AND dt.assignee_id = UUID_TO_BIN(?) 
		  AND dt.duty_id = UUID_TO_BIN(?)`

	rows, err := r.db.Query(query, assigneeID, dutyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query current duty_task by assignee id %s: %w", assigneeID, err)
	}
	defer rows.Close()

	var dutyTasksReadable []model.DutyTaskView

	for rows.Next() {
		var dutyTaskReadable model.DutyTaskView
		if err := rows.Scan(
			&dutyTaskReadable.AreaFloor,
			&dutyTaskReadable.AreaName,
			&dutyTaskReadable.TaskID,
			&dutyTaskReadable.TaskTitle,
			&dutyTaskReadable.TaskCost,
			&dutyTaskReadable.AssigneeFirstName,
			&dutyTaskReadable.AssigneeLastName,
			&dutyTaskReadable.CompletionDate,
			&dutyTaskReadable.VerificationDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan current duty_task row: %w", err)
		}

		dutyTasksReadable = append(dutyTasksReadable, dutyTaskReadable)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating current duty_task rows: %w", err)
	}

	return dutyTasksReadable, nil
}

func (r *DutyTaskRepository) FindUnassignedDutyTasksView(
	areaID int,
	dutyID uuid.UUID,
) ([]model.DutyTaskView, error) {
	query := `
		SELECT 
			a.area_floor, 
			a.area_name, 
			t.task_id,
			t.task_title, 
			t.task_cost,
			dt.completion_date, 
			dt.verification_date 
		FROM duty_task dt 
			INNER JOIN task t ON dt.task_id = t.task_id 
			INNER JOIN area a ON a.area_id = t.area_id
		WHERE dt.assignee_id IS NULL
		  AND t.area_id = ?
		  AND dt.duty_id = UUID_TO_BIN(?)`

	rows, err := r.db.Query(query, areaID, dutyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unassigned duty_tasks by area id %d: %w", areaID, err)
	}
	defer rows.Close()

	var dutyTasksReadable []model.DutyTaskView
	for rows.Next() {
		var dutyTaskReadable model.DutyTaskView
		if err := rows.Scan(
			&dutyTaskReadable.AreaFloor,
			&dutyTaskReadable.AreaName,
			&dutyTaskReadable.TaskID,
			&dutyTaskReadable.TaskTitle,
			&dutyTaskReadable.TaskCost,
			&dutyTaskReadable.CompletionDate,
			&dutyTaskReadable.VerificationDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan unassigned duty_task row: %w", err)
		}

		dutyTasksReadable = append(dutyTasksReadable, dutyTaskReadable)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unassigned duty_task rows: %w", err)
	}

	return dutyTasksReadable, nil
}

func (r *DutyTaskRepository) UpdateDutyTaskAssigneeID(
	assigneeID *uuid.UUID,
	dutyID uuid.UUID,
	taskID uuid.UUID,
) error {
	query := `
		UPDATE duty_task dt
		SET dt.assignee_id = UUID_TO_BIN(?) 
		WHERE dt.task_id = UUID_TO_BIN(?)
		  AND dt.duty_id = UUID_TO_BIN(?);`

	result, err := r.db.Exec(query, assigneeID, taskID, dutyID)
	if err != nil {
		return fmt.Errorf("failed to execute update for task_id %d: %w", taskID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected for task_id %d: %w", taskID, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *DutyTaskRepository) UpdateDutyTaskCompletionDate(dutyID uuid.UUID, taskID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE duty_task
		SET completion_date = ?
		WHERE duty_id = UUID_TO_BIN(?)
		  AND task_id = UUID_TO_BIN(?)`
	result, err := r.db.Exec(query, now, dutyID, taskID)
	if err != nil {
		return fmt.Errorf("failed to execute update for task_id %d: %w", taskID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected for task_id %d: %w", taskID, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
