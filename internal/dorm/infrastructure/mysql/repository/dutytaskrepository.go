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

func (r *DutyTaskRepository) FindReadableTasksByDutyID(dutyID uuid.UUID) ([]model.DutyTaskReadable, error) {
	query := `
		SELECT 
			a.area_floor, 
			a.area_name, 
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
		WHERE dt.duty_id = ?`

	rows, err := r.db.Query(query, dutyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query readable tasks for duty %s: %w", dutyID, err)
	}
	defer rows.Close()

	var tasksReadable []model.DutyTaskReadable

	for rows.Next() {
		var taskReadable model.DutyTaskReadable
		if err := rows.Scan(
			&taskReadable.AreaFloor,
			&taskReadable.AreaName,
			&taskReadable.TaskTitle,
			&taskReadable.TaskCost,
			&taskReadable.AssigneeFirstName,
			&taskReadable.AssigneeLastName,
			&taskReadable.CompletionDate,
			&taskReadable.VerificationDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan readable task row: %w", err)
		}

		tasksReadable = append(tasksReadable, taskReadable)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating readable task rows: %w", err)
	}

	return tasksReadable, nil
}

// TODO: придумать, что делать, если дежурство прошло, но человек всё равно хочет увидеть задачи за прошлую неделю, которые он не выполнил
func (r *DutyTaskRepository) FindCurrentDutyTasksByAssigneeID(assigneeID uuid.UUID) ([]model.DutyTask, error) {
	now := time.Now()
	query := `
		SELECT 
		    dt.duty_task_id, 
		    dt.duty_id, 
		    dt.task_id, 
		    dt.assignee_id, 
		    dt.reviewer_id, 
		    dt.assignment_date, 
		    dt.completion_date, 
		    dt.verification_date
		FROM duty_task dt
			INNER JOIN duty d ON dt.duty_id = d.duty_id
		WHERE d.duty_start_date < ? AND d.duty_end_date > ? AND dt.assignee_id = ?`

	rows, err := r.db.Query(query, now, now, assigneeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query current duty_task by assignee id %s: %w", assigneeID, err)
	}
	defer rows.Close()

	var dutyTasks []model.DutyTask

	for rows.Next() {
		var dutyTask model.DutyTask
		if err := rows.Scan(
			&dutyTask.DutyTaskID,
			&dutyTask.DutyID,
			&dutyTask.TaskID,
			&dutyTask.AssigneeID,
			&dutyTask.ReviewerID,
			&dutyTask.AssignmentDate,
			&dutyTask.CompletionDate,
			&dutyTask.VerificationDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan current duty_task row: %w", err)
		}

		dutyTasks = append(dutyTasks, dutyTask)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating readable task rows: %w", err)
	}

	return dutyTasks, nil
}
