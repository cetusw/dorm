package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"dorm/pkg/core/domain/catalog"

	"github.com/google/uuid"
)

type DutyTaskOverrideRepository struct {
	db *sql.DB
}

func NewDutyTaskOverrideRepository(db *sql.DB) *DutyTaskOverrideRepository {
	return &DutyTaskOverrideRepository{db: db}
}

func (r *DutyTaskOverrideRepository) FindAll(ctx context.Context) ([]*catalog.DutyTaskOverride, error) {
	const query = `SELECT task_id, include_in_next_duty FROM duty_task_override`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("DutyTaskOverrideRepository.FindAll: %w", err)
	}
	defer rows.Close()

	var overrides []*catalog.DutyTaskOverride
	for rows.Next() {
		var taskIDBytes []byte
		var include bool
		if err := rows.Scan(&taskIDBytes, &include); err != nil {
			return nil, fmt.Errorf("DutyTaskOverrideRepository.FindAll: scan row: %w", err)
		}
		taskID, err := uuid.FromBytes(taskIDBytes)
		if err != nil {
			return nil, fmt.Errorf("DutyTaskOverrideRepository.FindAll: parse task id: %w", err)
		}
		overrides = append(overrides, catalog.NewDutyTaskOverride(taskID, include))
	}
	return overrides, rows.Err()
}

func (r *DutyTaskOverrideRepository) ReplaceForTasks(ctx context.Context, taskIDs []uuid.UUID, overrides []*catalog.DutyTaskOverride) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("DutyTaskOverrideRepository.ReplaceForTasks: begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := deleteDutyTaskOverrides(ctx, tx, taskIDs); err != nil {
		return fmt.Errorf("DutyTaskOverrideRepository.ReplaceForTasks: %w", err)
	}
	if err := insertDutyTaskOverrides(ctx, tx, overrides); err != nil {
		return fmt.Errorf("DutyTaskOverrideRepository.ReplaceForTasks: %w", err)
	}
	return tx.Commit()
}

func (r *DutyTaskOverrideRepository) DeleteByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) error {
	if err := deleteDutyTaskOverrides(ctx, r.db, taskIDs); err != nil {
		return fmt.Errorf("DutyTaskOverrideRepository.DeleteByTaskIDs: %w", err)
	}
	return nil
}

func insertDutyTaskOverrides(ctx context.Context, exec sqlExecutor, overrides []*catalog.DutyTaskOverride) error {
	if len(overrides) == 0 {
		return nil
	}
	const query = `
		INSERT INTO duty_task_override (task_id, include_in_next_duty)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
			include_in_next_duty = VALUES(include_in_next_duty)
	`
	for _, override := range overrides {
		taskIDBytes, err := override.TaskID().MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal task id: %w", err)
		}
		if _, err := exec.ExecContext(ctx, query, taskIDBytes, override.IncludeInNextDuty()); err != nil {
			return fmt.Errorf("insert overrides: %w", err)
		}
	}
	return nil
}

func deleteDutyTaskOverrides(ctx context.Context, exec sqlExecutor, taskIDs []uuid.UUID) error {
	if len(taskIDs) == 0 {
		return nil
	}

	args := make([]interface{}, 0, len(taskIDs))
	placeholders := make([]string, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		taskIDBytes, err := taskID.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal task id: %w", err)
		}
		args = append(args, taskIDBytes)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf("DELETE FROM duty_task_override WHERE task_id IN (%s)", strings.Join(placeholders, ", "))
	if _, err := exec.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("delete overrides: %w", err)
	}
	return nil
}

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
