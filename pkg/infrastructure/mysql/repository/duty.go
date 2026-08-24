package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/duty"
)

type DutyRepository struct {
	db *sql.DB
}

func NewDutyRepository(db *sql.DB) *DutyRepository {
	return &DutyRepository{db: db}
}

func (r *DutyRepository) CreateWithTasks(
	ctx context.Context,
	currentDuty *duty.Duty,
	tasks []*duty.DutyTask,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin duty transaction: %w", err)
	}

	groupIDBytes, err := dutyGroupIDBytes(ctx, tx, currentDuty.TeamID())
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := lockDutyGroup(ctx, tx, groupIDBytes); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := ensureNoDutyOverlap(ctx, tx, groupIDBytes, currentDuty.Start(), currentDuty.End()); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := r.createDutyData(ctx, tx, currentDuty, tasks, groupIDBytes); err != nil {
		_ = tx.Rollback()
		return err
	}

	return commitDutyTransaction(tx)
}

func (r *DutyRepository) createDutyData(
	ctx context.Context,
	tx *sql.Tx,
	currentDuty *duty.Duty,
	tasks []*duty.DutyTask,
	groupIDBytes []byte,
) error {
	leaderID, err := dutyTeamLeaderID(ctx, tx, currentDuty.TeamID())
	if err != nil {
		return err
	}
	currentDuty.SetLeaderID(leaderID)
	if err := insertDuty(ctx, tx, currentDuty, groupIDBytes); err != nil {
		return err
	}
	if err := insertDutyParticipants(ctx, tx, currentDuty.ID(), currentDuty.TeamID(), leaderID); err != nil {
		return err
	}
	return insertDutyTasks(ctx, tx, currentDuty.ID(), tasks)
}

func insertDuty(ctx context.Context, tx *sql.Tx, d *duty.Duty, groupIDBytes []byte) error {
	const dutyQuery = `
		INSERT INTO duty (id, team_id, leader_id, group_id, start_date, end_date, sequence_number)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	dIDBytes, err := marshalUUID(d.ID(), "duty id")
	if err != nil {
		return err
	}
	tIDBytes, err := marshalUUID(d.TeamID(), "team id")
	if err != nil {
		return err
	}
	leaderID, err := nullableUUIDBytes(d.LeaderID(), "duty leader id")
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, dutyQuery, dIDBytes, tIDBytes, leaderID, groupIDBytes, d.Start(), d.End(), d.SequenceNumber())
	if err != nil {
		return fmt.Errorf("failed to save duty root: %w", err)
	}
	return nil
}

func dutyTeamLeaderID(ctx context.Context, tx *sql.Tx, teamID uuid.UUID) (*uuid.UUID, error) {
	teamIDBytes, err := marshalUUID(teamID, "team id")
	if err != nil {
		return nil, err
	}
	var raw []byte
	if err := tx.QueryRowContext(ctx, `SELECT leader_id FROM team WHERE id = ? AND deleted_at IS NULL`, teamIDBytes).Scan(&raw); err != nil {
		return nil, fmt.Errorf("load duty team leader: %w", err)
	}
	if len(raw) == 0 {
		return nil, nil
	}
	id, err := uuid.FromBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("parse duty team leader: %w", err)
	}
	return &id, nil
}

func insertDutyParticipants(ctx context.Context, tx *sql.Tx, dutyID, teamID uuid.UUID, leaderID *uuid.UUID) error {
	dutyIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return err
	}
	teamIDBytes, err := marshalUUID(teamID, "team id")
	if err != nil {
		return err
	}
	const query = `
		INSERT IGNORE INTO duty_participants (duty_id, participant_id, type, excluded_at)
		SELECT ?, id, 'REGULAR', NULL FROM user WHERE team_id = ? AND deleted_at IS NULL
	`
	if _, err := tx.ExecContext(ctx, query, dutyIDBytes, teamIDBytes); err != nil {
		return fmt.Errorf("snapshot duty members: %w", err)
	}
	if leaderID == nil {
		return nil
	}
	leaderIDBytes, err := marshalUUID(*leaderID, "leader id")
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO duty_participants (duty_id, participant_id, type, excluded_at) VALUES (?, ?, 'REGULAR', NULL)`, dutyIDBytes, leaderIDBytes); err != nil {
		return fmt.Errorf("snapshot duty leader: %w", err)
	}
	return nil
}

func dutyGroupIDBytes(ctx context.Context, tx *sql.Tx, teamID uuid.UUID) ([]byte, error) {
	const query = `SELECT group_id FROM team WHERE id = ? AND deleted_at IS NULL`

	teamIDBytes, err := marshalUUID(teamID, "team id")
	if err != nil {
		return nil, err
	}

	var groupIDBytes []byte
	if err := tx.QueryRowContext(ctx, query, teamIDBytes).Scan(&groupIDBytes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("load duty group: team not found")
		}
		return nil, fmt.Errorf("load duty group: %w", err)
	}

	return groupIDBytes, nil
}

func lockDutyGroup(ctx context.Context, tx *sql.Tx, groupIDBytes []byte) error {
	const query = "SELECT id FROM `group` WHERE id = ? FOR UPDATE"

	var lockedGroupID []byte
	if err := tx.QueryRowContext(ctx, query, groupIDBytes).Scan(&lockedGroupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("lock duty group: group not found")
		}
		return fmt.Errorf("lock duty group: %w", err)
	}

	return nil
}

func ensureNoDutyOverlap(ctx context.Context, tx *sql.Tx, groupIDBytes []byte, startDate, endDate time.Time) error {
	const query = `
		SELECT 1
		FROM duty
		WHERE group_id = ?
		  AND start_date < ?
		  AND end_date > ?
		LIMIT 1
	`

	var marker int
	if err := tx.QueryRowContext(ctx, query, groupIDBytes, endDate, startDate).Scan(&marker); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("check duty overlap: %w", err)
	}

	return duty.ErrDutyPeriodOverlap
}

func insertDutyTasks(
	ctx context.Context,
	tx *sql.Tx,
	dutyID uuid.UUID,
	tasks []*duty.DutyTask,
) error {
	const taskQuery = `
		INSERT INTO duty_task (
			id, duty_id, task_id, assignee_id, reviewer_id,
			assignment_date, completion_date, verification_date
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	dIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if err := insertDutyTask(ctx, tx, taskQuery, dIDBytes, task); err != nil {
			return err
		}
	}
	return nil
}

func insertDutyTask(ctx context.Context, tx *sql.Tx, query string, dutyID []byte, task *duty.DutyTask) error {
	taskID, taskDefID, err := marshalDutyTaskInsertIDs(task)
	if err != nil {
		return err
	}
	assigneeID, reviewerID, err := nullableDutyTaskUserIDs(task)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, query, taskID, dutyID, taskDefID,
		assigneeID, reviewerID,
		nullTime(task.AssignmentDate()), nullTime(task.CompletionDate()),
		nullTime(task.VerificationDate()))
	if err != nil {
		return fmt.Errorf("failed to save duty task %s: %w", task.ID(), err)
	}
	return nil
}

func marshalDutyTaskInsertIDs(task *duty.DutyTask) ([]byte, []byte, error) {
	taskID, err := marshalUUID(task.ID(), "duty task id")
	if err != nil {
		return nil, nil, err
	}
	taskDefID, err := marshalUUID(task.TaskDefID(), "task definition id")
	return taskID, taskDefID, err
}

func nullableDutyTaskUserIDs(task *duty.DutyTask) (any, any, error) {
	assigneeID, err := nullableUUIDBytes(task.AssigneeID(), "assignee id")
	if err != nil {
		return nil, nil, err
	}
	reviewerID, err := nullableUUIDBytes(task.ReviewerID(), "reviewer id")
	return assigneeID, reviewerID, err
}

func commitDutyTransaction(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit duty transaction: %w", err)
	}
	return nil
}

func (r *DutyRepository) FindCurrentByTeamID(ctx context.Context, teamID uuid.UUID) (*duty.Duty, error) {
	const dutyQuery = `
		SELECT id, team_id, start_date, end_date, sequence_number
		FROM duty
		WHERE team_id = ? AND start_date = (
			SELECT MAX(start_date) FROM duty WHERE team_id = ?
		)
	`
	tIDBytes, _ := teamID.MarshalBinary()
	row := r.db.QueryRowContext(ctx, dutyQuery, tIDBytes, tIDBytes)

	var dID, teamIDBytes []byte
	var start, end time.Time
	var sequenceNumber int

	err := row.Scan(&dID, &teamIDBytes, &start, &end, &sequenceNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find duty root: %w", err)
	}

	dutyID, _ := uuid.FromBytes(dID)

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	return duty.RestoreDuty(dutyID, teamID, start, end, sequenceNumber, tasks), nil
}

func (r *DutyRepository) ReassignTeamAndResetTasks(ctx context.Context, dutyID uuid.UUID, teamID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reassign duty transaction: %w", err)
	}

	if err := reassignDutyTeam(ctx, tx, dutyID, teamID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := resetDutyTasks(ctx, tx, dutyID); err != nil {
		_ = tx.Rollback()
		return err
	}

	return commitDutyTransaction(tx)
}

func reassignDutyTeam(ctx context.Context, tx *sql.Tx, dutyID uuid.UUID, teamID uuid.UUID) error {
	const query = `
		UPDATE duty
		SET team_id = ?,
			group_id = (SELECT group_id FROM team WHERE id = ?)
		WHERE id = ?
	`

	teamIDBytes, err := marshalUUID(teamID, "team id")
	if err != nil {
		return err
	}
	dutyIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, query, teamIDBytes, teamIDBytes, dutyIDBytes); err != nil {
		return fmt.Errorf("reassign duty team: %w", err)
	}

	return nil
}

func resetDutyTasks(ctx context.Context, tx *sql.Tx, dutyID uuid.UUID) error {
	const query = `
		UPDATE duty_task
		SET assignee_id = NULL,
			reviewer_id = NULL,
			assignment_date = NULL,
			completion_date = NULL,
			verification_date = NULL
		WHERE duty_id = ?
	`

	dutyIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, query, dutyIDBytes); err != nil {
		return fmt.Errorf("reset duty tasks: %w", err)
	}

	return nil
}

func (r *DutyRepository) FindActiveByTeamID(
	ctx context.Context,
	teamID uuid.UUID,
	at time.Time,
) (*duty.Duty, error) {
	const query = `
		SELECT id, team_id, leader_id, start_date, end_date, sequence_number
		FROM duty
		WHERE team_id = ?
		  AND start_date <= ?
		  AND end_date > ?
		ORDER BY start_date DESC
		LIMIT 1
	`

	teamIDBytes, err := teamID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("find active duty: marshal team id: %w", err)
	}

	row := r.db.QueryRowContext(ctx, query, teamIDBytes, at, at)

	var dutyIDBytes, teamIDResultBytes, leaderIDBytes []byte
	var startDate, endDate time.Time
	var sequenceNumber int

	if err := row.Scan(&dutyIDBytes, &teamIDResultBytes, &leaderIDBytes, &startDate, &endDate, &sequenceNumber); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find active duty: %w", err)
	}

	dutyID, err := uuid.FromBytes(dutyIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find active duty: parse duty id: %w", err)
	}

	teamIDResult, err := uuid.FromBytes(teamIDResultBytes)
	if err != nil {
		return nil, fmt.Errorf("find active duty: parse team id: %w", err)
	}

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	result := duty.RestoreDuty(dutyID, teamIDResult, startDate, endDate, sequenceNumber, tasks)
	if len(leaderIDBytes) > 0 {
		leaderID, err := uuid.FromBytes(leaderIDBytes)
		if err != nil {
			return nil, fmt.Errorf("find active duty: parse leader id: %w", err)
		}
		result.SetLeaderID(&leaderID)
	}
	return result, nil
}

func (r *DutyRepository) FindLatestByTeamID(ctx context.Context, teamID uuid.UUID) (*duty.Duty, error) {
	const dutyQuery = `
		SELECT id, team_id, start_date, end_date, sequence_number
		FROM duty
		WHERE team_id = ?
		ORDER BY sequence_number DESC, start_date DESC, id DESC
		LIMIT 1
	`
	tIDBytes, _ := teamID.MarshalBinary()
	row := r.db.QueryRowContext(ctx, dutyQuery, tIDBytes)

	var dID, teamIDBytes []byte
	var start, end time.Time
	var sequenceNumber int

	err := row.Scan(&dID, &teamIDBytes, &start, &end, &sequenceNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find duty root: %w", err)
	}

	dutyID, _ := uuid.FromBytes(dID)

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	return duty.RestoreDuty(dutyID, teamID, start, end, sequenceNumber, tasks), nil
}

func (r *DutyRepository) FindByID(ctx context.Context, id uuid.UUID) (*duty.Duty, error) {
	const dutyQuery = `SELECT id, team_id, leader_id, start_date, end_date, sequence_number FROM duty WHERE id = ?`

	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, dutyQuery, idBytes)

	var dID, tID, leaderIDBytes []byte
	var start, end time.Time
	var sequenceNumber int

	err := row.Scan(&dID, &tID, &leaderIDBytes, &start, &end, &sequenceNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	dutyID, _ := uuid.FromBytes(dID)
	teamID, _ := uuid.FromBytes(tID)

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	result := duty.RestoreDuty(dutyID, teamID, start, end, sequenceNumber, tasks)
	if len(leaderIDBytes) > 0 {
		leaderID, err := uuid.FromBytes(leaderIDBytes)
		if err != nil {
			return nil, fmt.Errorf("parse duty leader: %w", err)
		}
		result.SetLeaderID(&leaderID)
	}
	return result, nil
}

func (r *DutyRepository) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*duty.Duty, error) {
	const query = `
		SELECT d.id, d.team_id, d.start_date, d.end_date, d.sequence_number
		FROM duty d
		WHERE d.group_id = ?
		ORDER BY d.start_date DESC, d.sequence_number DESC, d.id DESC
	`

	groupIDBytes, _ := groupID.MarshalBinary()
	rows, err := r.db.QueryContext(ctx, query, groupIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find duties by group: %w", err)
	}
	defer rows.Close()

	var result []*duty.Duty
	for rows.Next() {
		var dID, tID []byte
		var start, end time.Time
		var sequenceNumber int
		if err := rows.Scan(&dID, &tID, &start, &end, &sequenceNumber); err != nil {
			return nil, err
		}
		id, _ := uuid.FromBytes(dID)
		teamID, _ := uuid.FromBytes(tID)

		tasks, err := r.findTasksByDutyID(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, duty.RestoreDuty(id, teamID, start, end, sequenceNumber, tasks))
	}
	return result, rows.Err()
}

func (r *DutyRepository) FindLatestByGroupID(ctx context.Context, groupID uuid.UUID) (*duty.Duty, error) {
	const query = `
		SELECT d.id, d.team_id, d.start_date, d.end_date, d.sequence_number
		FROM duty d
		WHERE d.group_id = ?
		ORDER BY d.sequence_number DESC, d.start_date DESC, d.id DESC
		LIMIT 1
	`

	groupIDBytes, _ := groupID.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, groupIDBytes)

	var dutyIDBytes, teamIDBytes []byte
	var startDate, endDate time.Time
	var sequenceNumber int

	if err := row.Scan(&dutyIDBytes, &teamIDBytes, &startDate, &endDate, &sequenceNumber); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find latest duty by group: %w", err)
	}

	dutyID, _ := uuid.FromBytes(dutyIDBytes)
	teamID, _ := uuid.FromBytes(teamIDBytes)

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	return duty.RestoreDuty(dutyID, teamID, startDate, endDate, sequenceNumber, tasks), nil
}

func (r *DutyRepository) CountDistinctStartDates(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(DISTINCT start_date) FROM duty`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count distinct start dates: %w", err)
	}

	return count, nil
}

func (r *DutyRepository) FindLastByTaskDefID(ctx context.Context, taskDefID uuid.UUID) (*duty.Duty, error) {
	const query = `
		SELECT d.id, d.team_id, d.start_date, d.end_date, d.sequence_number
		FROM duty d
		JOIN duty_task dt ON d.id = dt.duty_id
		WHERE dt.task_id = ?
		ORDER BY d.sequence_number DESC
		LIMIT 1
	`

	defIDBytes, _ := taskDefID.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, defIDBytes)

	var dID, tID []byte
	var start, end time.Time
	var sequenceNumber int

	err := row.Scan(&dID, &tID, &start, &end, &sequenceNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find last by task def: %w", err)
	}

	dutyID, _ := uuid.FromBytes(dID)
	teamID, _ := uuid.FromBytes(tID)

	tasks, err := r.findTasksByDutyID(ctx, dutyID)
	if err != nil {
		return nil, err
	}

	return duty.RestoreDuty(dutyID, teamID, start, end, sequenceNumber, tasks), nil
}

func (r *DutyRepository) FindAllLatest(ctx context.Context) ([]*duty.Duty, error) {
	const query = `
		SELECT id, team_id, start_date, end_date, sequence_number 
		FROM duty 
		WHERE start_date = (SELECT MAX(start_date) FROM duty)`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all latest duties: %w", err)
	}
	defer rows.Close()

	var result []*duty.Duty
	for rows.Next() {
		var dID, tID []byte
		var start, end time.Time
		var sequenceNumber int
		if err := rows.Scan(&dID, &tID, &start, &end, &sequenceNumber); err != nil {
			return nil, err
		}
		id, _ := uuid.FromBytes(dID)
		teamID, _ := uuid.FromBytes(tID)

		tasks, err := r.findTasksByDutyID(ctx, id)
		if err != nil {
			return nil, err
		}

		result = append(result, duty.RestoreDuty(id, teamID, start, end, sequenceNumber, tasks))
	}
	return result, rows.Err()
}

func (r *DutyRepository) FindHistoryByGroupID(ctx context.Context, groupID uuid.UUID) ([]duty.DutyHistoryEntry, error) {
	const query = `
		SELECT
			d.id,
			d.team_id,
			d.start_date,
			d.end_date,
			d.sequence_number,
			COALESCE(NULLIF(CONCAT_WS(' ', leader.first_name, leader.last_name), ''), 'Глава команды не назначен') AS team_leader_name,
			COALESCE(SUM(tk.cost), 0) AS total_cost_sum,
			COALESCE(SUM(CASE WHEN dt.assignee_id IS NOT NULL THEN tk.cost ELSE 0 END), 0) AS taken_cost_sum,
			COUNT(dt.id) AS total_tasks_count,
			COALESCE(SUM(CASE WHEN dt.assignee_id IS NOT NULL THEN 1 ELSE 0 END), 0) AS taken_tasks_count,
			COALESCE(SUM(CASE WHEN dt.completion_date IS NOT NULL THEN 1 ELSE 0 END), 0) AS completed_tasks_count,
			COALESCE(SUM(CASE WHEN dt.verification_date IS NOT NULL THEN 1 ELSE 0 END), 0) AS verified_tasks_count
		FROM duty d
		LEFT JOIN user leader ON leader.id = d.leader_id
		LEFT JOIN duty_task dt ON dt.duty_id = d.id
		LEFT JOIN task tk ON tk.id = dt.task_id
		WHERE d.group_id = ?
		GROUP BY d.id, d.team_id, d.start_date, d.end_date, d.sequence_number, team_leader_name
		ORDER BY d.start_date DESC, d.sequence_number DESC, d.id DESC
	`

	groupIDBytes, err := marshalUUID(groupID, "group id")
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, groupIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find duty history by group: %w", err)
	}
	defer rows.Close()

	result := make([]duty.DutyHistoryEntry, 0)
	for rows.Next() {
		var dutyIDBytes []byte
		var teamIDBytes []byte
		var entry duty.DutyHistoryEntry

		if err := rows.Scan(
			&dutyIDBytes,
			&teamIDBytes,
			&entry.Start,
			&entry.End,
			&entry.SequenceNumber,
			&entry.TeamLeaderName,
			&entry.TotalCostSum,
			&entry.TakenCostSum,
			&entry.TotalTasksCount,
			&entry.TakenTasksCount,
			&entry.CompletedTasksCount,
			&entry.VerifiedTasksCount,
		); err != nil {
			return nil, fmt.Errorf("scan duty history row: %w", err)
		}

		entry.DutyID, err = uuid.FromBytes(dutyIDBytes)
		if err != nil {
			return nil, fmt.Errorf("parse duty history duty id: %w", err)
		}
		entry.TeamID, err = uuid.FromBytes(teamIDBytes)
		if err != nil {
			return nil, fmt.Errorf("parse duty history team id: %w", err)
		}

		result = append(result, entry)
	}

	return result, rows.Err()
}

func (r *DutyRepository) findTasksByDutyID(ctx context.Context, dutyID uuid.UUID) ([]*duty.DutyTask, error) {
	const query = `
		SELECT id, duty_id, task_id, assignee_id, reviewer_id,
			assignment_date, completion_date, verification_date
		FROM duty_task
		WHERE duty_id = ?
	`
	dIDBytes, _ := dutyID.MarshalBinary()
	rows, err := r.db.QueryContext(ctx, query, dIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*duty.DutyTask
	for rows.Next() {
		task, err := scanDutyTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
