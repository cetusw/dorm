package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"dorm/pkg/dorm/application/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type DutyTaskRepository struct {
	db *sql.DB
}

func NewDutyTaskRepository(db *sql.DB) *DutyTaskRepository {
	return &DutyTaskRepository{db: db}
}

func (r *DutyTaskRepository) FindDutyTaskViewsByDutyID(dutyID uuid.UUID) ([]model.DutyTaskView, error) {
	const sqlQuery = `
		SELECT 
			a.floor, 
			a.name, 
			t.id,
			t.title, 
			t.cost, 
			u.id,
			u.first_name, 
			u.last_name, 
			dt.completion_date, 
			dt.verification_date 
		FROM duty_task dt 
			INNER JOIN task t ON dt.task_id = t.id 
			INNER JOIN area a ON a.id = t.area_id 
			LEFT JOIN user u ON u.id = dt.assignee_id 
		WHERE dt.duty_id = UUID_TO_BIN(?)
		ORDER BY a.floor DESC, a.name DESC, t.cost DESC`

	rows, err := r.db.Query(sqlQuery, dutyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query readable tasks for duty %s: %w", dutyID, err)
	}
	defer rows.Close()

	var dutyTasksView []model.DutyTaskView
	for rows.Next() {
		var dutyTaskView model.DutyTaskView
		if err := rows.Scan(
			&dutyTaskView.AreaFloor,
			&dutyTaskView.AreaName,
			&dutyTaskView.TaskID,
			&dutyTaskView.TaskTitle,
			&dutyTaskView.TaskCost,
			&dutyTaskView.AssigneeID,
			&dutyTaskView.AssigneeFirstName,
			&dutyTaskView.AssigneeLastName,
			&dutyTaskView.CompletionDate,
			&dutyTaskView.VerificationDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan readable task row: %w", err)
		}

		dutyTasksView = append(dutyTasksView, dutyTaskView)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating readable task rows: %w", err)
	}

	return dutyTasksView, nil
}

// TODO: придумать, что делать, если дежурство прошло, но человек всё равно хочет увидеть задачи за прошлую неделю, которые он не выполнил
func (r *DutyTaskRepository) FindUncompletedDutyTasksView(
	assigneeID uuid.UUID,
	dutyID uuid.UUID,
) ([]model.DutyTaskView, error) {
	const sqlQuery = `
		SELECT 
			a.floor, 
			a.name, 
			t.id,
			t.title, 
			t.cost, 
			u.first_name, 
			u.last_name, 
			dt.completion_date, 
			dt.verification_date 
		FROM duty_task dt 
			INNER JOIN task t ON dt.task_id = t.id 
			INNER JOIN area a ON a.id = t.area_id
			LEFT JOIN user u ON u.id = dt.assignee_id
		WHERE dt.completion_date IS NULL 
		  AND dt.assignee_id = UUID_TO_BIN(?) 
		  AND dt.duty_id = UUID_TO_BIN(?)`

	rows, err := r.db.Query(sqlQuery, assigneeID, dutyID)
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
	const sqlQuery = `
		SELECT 
			a.floor, 
			a.name, 
			t.id,
			t.title, 
			t.cost,
			dt.completion_date, 
			dt.verification_date 
		FROM duty_task dt 
			INNER JOIN task t ON dt.task_id = t.id 
			INNER JOIN area a ON a.id = t.area_id
		WHERE dt.assignee_id IS NULL
		  AND t.area_id = ?
		  AND dt.duty_id = UUID_TO_BIN(?)`

	rows, err := r.db.Query(sqlQuery, areaID, dutyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unassigned duty_tasks by area id %d: %w", areaID, err)
	}
	defer rows.Close()

	var dutyTasksView []model.DutyTaskView
	for rows.Next() {
		var dutyTaskView model.DutyTaskView
		if err := rows.Scan(
			&dutyTaskView.AreaFloor,
			&dutyTaskView.AreaName,
			&dutyTaskView.TaskID,
			&dutyTaskView.TaskTitle,
			&dutyTaskView.TaskCost,
			&dutyTaskView.CompletionDate,
			&dutyTaskView.VerificationDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan unassigned duty_task row: %w", err)
		}

		dutyTasksView = append(dutyTasksView, dutyTaskView)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unassigned duty_task rows: %w", err)
	}

	return dutyTasksView, nil
}

func (r *DutyTaskRepository) UpdateDutyTaskAssigneeID(
	assigneeID *uuid.UUID,
	dutyID uuid.UUID,
	taskID uuid.UUID,
) error {
	const sqlQuery = `
		UPDATE duty_task dt
		SET dt.assignee_id = UUID_TO_BIN(?) 
		WHERE dt.task_id = UUID_TO_BIN(?)
		  AND dt.duty_id = UUID_TO_BIN(?);`

	result, err := r.db.Exec(sqlQuery, assigneeID, taskID, dutyID)
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

func (r *DutyTaskRepository) SetDutyTaskCompletionDate(dutyID uuid.UUID, taskID uuid.UUID, completionDate *time.Time) error {
	const sqlQuery = `
		UPDATE duty_task
		SET completion_date = ?
		WHERE duty_id = UUID_TO_BIN(?)
		  AND task_id = UUID_TO_BIN(?)`

	result, err := r.db.Exec(sqlQuery, completionDate, dutyID, taskID)
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

func (r *DutyTaskRepository) Store(dutyTask *model.DutyTask) error {
	const sqlQuery = `
	INSERT INTO duty_task (id, duty_id, task_id, reviewer_id) 
	VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?), UUID_TO_BIN(?), UUID_TO_BIN(?))`

	_, err := r.db.Exec(sqlQuery, dutyTask.ID, dutyTask.DutyID, dutyTask.TaskID, dutyTask.ReviewerID)
	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	return nil
}

func (r *DutyTaskRepository) StoreBatch(dutyTasks []model.DutyTask) error {
	if len(dutyTasks) == 0 {
		return nil
	}

	const sqlQuery = `
		INSERT INTO duty_task (id, duty_id, task_id, reviewer_id) 
		VALUES `

	var valueStrings []string
	var valueArgs []interface{}

	for _, dutyTask := range dutyTasks {
		valueStrings = append(valueStrings, "(UUID_TO_BIN(?), UUID_TO_BIN(?), UUID_TO_BIN(?), UUID_TO_BIN(?))")
		valueArgs = append(valueArgs, dutyTask.ID, dutyTask.DutyID, dutyTask.TaskID, dutyTask.ReviewerID)
	}

	stmt := fmt.Sprintf("%s %s", sqlQuery, strings.Join(valueStrings, ","))

	_, err := r.db.Exec(stmt, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch insert duty_tasks: %w", err)
	}

	return nil
}
