package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"fmt"
	"strings"

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

func (r *DutyTaskRepository) GetDutyTaskRows(dutyID uuid.UUID) (*sql.Rows, error) {
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
	WHERE dt.duty_id = UUID_TO_BIN(?)`

	rows, err := r.db.Query(query, dutyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasksReadable: %w", err)
	}

	return rows, nil
}
