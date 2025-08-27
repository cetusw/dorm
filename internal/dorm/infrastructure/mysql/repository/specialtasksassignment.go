package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SpecialTaskAssignmentRepository struct {
	db *sql.DB
}

func NewSpecialTaskAssignmentRepository(db *sql.DB) *SpecialTaskAssignmentRepository {
	return &SpecialTaskAssignmentRepository{db: db}
}

type SpecialTaskAssignmentRow struct {
	AssignmentID   uuid.UUID
	TaskID         uuid.UUID
	AssigneeID     uuid.UUID
	AssignmentDate time.Time
	CompletionDate *time.Time
}

func (r *SpecialTaskAssignmentRepository) FindLatestByTaskID(taskID uuid.UUID) (*SpecialTaskAssignmentRow, error) {
	const q = `
		SELECT assignment_id, task_id, assignee_id, assignment_date, completion_date
		FROM special_task_assignment
		WHERE task_id = ?
		ORDER BY assignment_date DESC
		LIMIT 1`
	row := r.db.QueryRow(q, taskID[:])

	var res SpecialTaskAssignmentRow
	var assignmentID, tID, aID []byte
	var completion sql.NullTime

	if err := row.Scan(&assignmentID, &tID, &aID, &res.AssignmentDate, &completion); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindLatestByTaskID: %w", err)
	}

	var err error
	res.AssignmentID, err = uuid.FromBytes(assignmentID)
	if err != nil {
		return nil, fmt.Errorf("parse assignment_id: %w", err)
	}
	res.TaskID, err = uuid.FromBytes(tID)
	if err != nil {
		return nil, fmt.Errorf("parse task_id: %w", err)
	}
	res.AssigneeID, err = uuid.FromBytes(aID)
	if err != nil {
		return nil, fmt.Errorf("parse assignee_id: %w", err)
	}
	if completion.Valid {
		c := completion.Time
		res.CompletionDate = &c
	}
	return &res, nil
}

func (r *SpecialTaskAssignmentRepository) StoreBatch(rows []SpecialTaskAssignmentRow) error {
	if len(rows) == 0 {
		return nil
	}
	const q = `
		INSERT INTO special_task_assignment (assignment_id, task_id, assignee_id, assignment_date, completion_date)
		VALUES (?, ?, ?, ?, ?)`
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	stmt, err := tx.Prepare(q)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, rws := range rows {
		var completion any
		if rws.CompletionDate != nil {
			completion = *rws.CompletionDate
		} else {
			completion = nil
		}
		if _, err := stmt.Exec(rws.AssignmentID[:], rws.TaskID[:], rws.AssigneeID[:], rws.AssignmentDate, completion); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
