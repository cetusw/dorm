package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/duty"
)

type dutyTaskScanner interface {
	Scan(dest ...any) error
}

type dutyTaskRecord struct {
	id               []byte
	dutyID           []byte
	taskDefID        []byte
	assigneeID       []byte
	reviewerID       []byte
	assignmentDate   sql.NullTime
	completionDate   sql.NullTime
	verificationDate sql.NullTime
}

func findDutyTasksByDutyID(ctx context.Context, db *sql.DB, dutyID uuid.UUID) ([]*duty.DutyTask, error) {
	const query = `
		SELECT id, duty_id, task_id, assignee_id, reviewer_id,
			assignment_date, completion_date, verification_date
		FROM duty_task
		WHERE duty_id = ?
	`
	dutyIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, query, dutyIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find duty tasks: %w", err)
	}
	defer rows.Close()
	return scanDutyTasks(rows)
}

func scanDutyTasks(rows *sql.Rows) ([]*duty.DutyTask, error) {
	var tasks []*duty.DutyTask
	for rows.Next() {
		task, err := scanDutyTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func scanDutyTask(scanner dutyTaskScanner) (*duty.DutyTask, error) {
	var record dutyTaskRecord
	err := scanner.Scan(&record.id, &record.dutyID, &record.taskDefID,
		&record.assigneeID, &record.reviewerID, &record.assignmentDate,
		&record.completionDate, &record.verificationDate)
	if err != nil {
		return nil, err
	}
	return restoreDutyTask(record)
}

func restoreDutyTask(record dutyTaskRecord) (*duty.DutyTask, error) {
	id, dutyID, taskDefID, err := parseDutyTaskIDs(record)
	if err != nil {
		return nil, err
	}
	return duty.RestoreDutyTask(duty.RestoreDutyTaskParams{
		ID:               id,
		DutyID:           dutyID,
		TaskDefID:        taskDefID,
		AssigneeID:       uuidFromBytes(record.assigneeID),
		ReviewerID:       uuidFromBytes(record.reviewerID),
		AssignmentDate:   timeFromNull(record.assignmentDate),
		CompletionDate:   timeFromNull(record.completionDate),
		VerificationDate: timeFromNull(record.verificationDate),
	}), nil
}

func parseDutyTaskIDs(record dutyTaskRecord) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	id, err := uuid.FromBytes(record.id)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("parse duty task id: %w", err)
	}
	dutyID, err := uuid.FromBytes(record.dutyID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("parse duty id: %w", err)
	}
	taskDefID, err := uuid.FromBytes(record.taskDefID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("parse task definition id: %w", err)
	}
	return id, dutyID, taskDefID, nil
}

func uuidFromBytes(bytes []byte) *uuid.UUID {
	if len(bytes) == 0 {
		return nil
	}
	value, err := uuid.FromBytes(bytes)
	if err != nil {
		return nil
	}
	return &value
}

func marshalUUID(id uuid.UUID, name string) ([]byte, error) {
	bytes, err := id.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal %s: %w", name, err)
	}
	return bytes, nil
}

func nullableUUIDBytes(id *uuid.UUID, name string) (any, error) {
	if id == nil {
		return nil, nil
	}
	bytes, err := marshalUUID(*id, name)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func nullTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}

func timeFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
