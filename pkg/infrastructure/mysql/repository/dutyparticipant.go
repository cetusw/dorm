package repository

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"dorm/pkg/core/domain/duty"
	"github.com/google/uuid"
)

type DutyParticipantRepository struct{ db *sql.DB }

var _ duty.ParticipantRepository = (*DutyParticipantRepository)(nil)

func NewDutyParticipantRepository(db *sql.DB) *DutyParticipantRepository {
	return &DutyParticipantRepository{db: db}
}

func (r *DutyParticipantRepository) List(ctx context.Context, dutyID uuid.UUID, includeExcluded bool) ([]duty.DutyParticipant, error) {
	query := `SELECT duty_id, participant_id, type, excluded_at FROM duty_participants WHERE duty_id = ?`
	if !includeExcluded {
		query += ` AND excluded_at IS NULL`
	}
	query += ` ORDER BY participant_id`
	dutyIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, query, dutyIDBytes)
	if err != nil {
		return nil, fmt.Errorf("list duty participants: %w", err)
	}
	defer rows.Close()
	return scanParticipants(rows)
}

func (r *DutyParticipantRepository) IsActive(ctx context.Context, dutyID, participantID uuid.UUID) (bool, error) {
	dutyIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return false, err
	}
	participantIDBytes, err := marshalUUID(participantID, "participant id")
	if err != nil {
		return false, err
	}
	var marker int
	err = r.db.QueryRowContext(ctx, `SELECT 1 FROM duty_participants WHERE duty_id = ? AND participant_id = ? AND excluded_at IS NULL`, dutyIDBytes, participantIDBytes).Scan(&marker)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check active duty participant: %w", err)
	}
	return true, nil
}

func (r *DutyParticipantRepository) Add(ctx context.Context, p duty.DutyParticipant, start, end time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin add duty participant: %w", err)
	}
	defer tx.Rollback()
	if err := lockDuty(ctx, tx, p.DutyID); err != nil {
		return err
	}
	if err := ensureNoActiveParticipantOverlap(ctx, tx, p.ParticipantID, p.DutyID, start, end); err != nil {
		return err
	}
	dutyIDBytes, err := marshalUUID(p.DutyID, "duty id")
	if err != nil {
		return err
	}
	participantIDBytes, err := marshalUUID(p.ParticipantID, "participant id")
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO duty_participants (duty_id, participant_id, type, excluded_at) VALUES (?, ?, ?, NULL)`, dutyIDBytes, participantIDBytes, p.Type)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return duty.ErrParticipantAlreadyActive
		}
		return fmt.Errorf("add duty participant: %w", err)
	}
	return tx.Commit()
}

func (r *DutyParticipantRepository) Restore(ctx context.Context, dutyID, participantID uuid.UUID, start, end time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin restore duty participant: %w", err)
	}
	defer tx.Rollback()
	if err := lockDuty(ctx, tx, dutyID); err != nil {
		return err
	}
	if err := ensureNoActiveParticipantOverlap(ctx, tx, participantID, dutyID, start, end); err != nil {
		return err
	}
	dutyIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return err
	}
	participantIDBytes, err := marshalUUID(participantID, "participant id")
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE duty_participants SET excluded_at = NULL WHERE duty_id = ? AND participant_id = ? AND excluded_at IS NOT NULL`, dutyIDBytes, participantIDBytes)
	if err != nil {
		return fmt.Errorf("restore duty participant: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return duty.ErrParticipantAlreadyActive
	}
	return tx.Commit()
}

func (r *DutyParticipantRepository) Exclude(ctx context.Context, dutyID, participantID uuid.UUID, replacementLeaderID *uuid.UUID, excludedAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin exclude duty participant: %w", err)
	}
	defer tx.Rollback()
	if err := lockDuty(ctx, tx, dutyID); err != nil {
		return err
	}
	dutyIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return err
	}
	participantIDBytes, err := marshalUUID(participantID, "participant id")
	if err != nil {
		return err
	}
	var leaderRaw []byte
	if err := tx.QueryRowContext(ctx, `SELECT leader_id FROM duty WHERE id = ? FOR UPDATE`, dutyIDBytes).Scan(&leaderRaw); err != nil {
		return fmt.Errorf("load duty leader: %w", err)
	}
	if len(leaderRaw) > 0 && bytes.Equal(leaderRaw, participantIDBytes) {
		if replacementLeaderID == nil || *replacementLeaderID == participantID {
			return duty.ErrLeaderReplacementRequired
		}
		replacementBytes, err := marshalUUID(*replacementLeaderID, "replacement leader id")
		if err != nil {
			return err
		}
		var marker int
		if err := tx.QueryRowContext(ctx, `SELECT 1 FROM duty_participants WHERE duty_id = ? AND participant_id = ? AND excluded_at IS NULL`, dutyIDBytes, replacementBytes).Scan(&marker); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return duty.ErrParticipantNotFound
			}
			return fmt.Errorf("check replacement leader: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE duty SET leader_id = ? WHERE id = ?`, replacementBytes, dutyIDBytes); err != nil {
			return fmt.Errorf("replace duty leader: %w", err)
		}
	}
	result, err := tx.ExecContext(ctx, `UPDATE duty_participants SET excluded_at = ? WHERE duty_id = ? AND participant_id = ? AND excluded_at IS NULL`, excludedAt, dutyIDBytes, participantIDBytes)
	if err != nil {
		return fmt.Errorf("exclude duty participant: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return duty.ErrParticipantAlreadyExcluded
	}
	if _, err := tx.ExecContext(ctx, `UPDATE duty_task SET assignee_id = NULL, assignment_date = NULL, reviewer_id = NULL WHERE duty_id = ? AND assignee_id = ? AND completion_date IS NULL`, dutyIDBytes, participantIDBytes); err != nil {
		return fmt.Errorf("release participant tasks: %w", err)
	}
	return tx.Commit()
}

func (r *DutyParticipantRepository) ChangeLeader(ctx context.Context, dutyID, leaderID uuid.UUID) error {
	dutyIDBytes, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return err
	}
	leaderIDBytes, err := marshalUUID(leaderID, "leader id")
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `UPDATE duty d JOIN duty_participants p ON p.duty_id = d.id AND p.participant_id = ? AND p.excluded_at IS NULL SET d.leader_id = ? WHERE d.id = ?`, leaderIDBytes, leaderIDBytes, dutyIDBytes)
	if err != nil {
		return fmt.Errorf("change duty leader: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return duty.ErrParticipantNotFound
	}
	return nil
}

func lockDuty(ctx context.Context, tx *sql.Tx, dutyID uuid.UUID) error {
	b, err := marshalUUID(dutyID, "duty id")
	if err != nil {
		return err
	}
	var id []byte
	if err := tx.QueryRowContext(ctx, `SELECT id FROM duty WHERE id = ? FOR UPDATE`, b).Scan(&id); err != nil {
		return fmt.Errorf("lock duty: %w", err)
	}
	return nil
}
func ensureNoActiveParticipantOverlap(ctx context.Context, tx *sql.Tx, participantID, ignoredDutyID uuid.UUID, start, end time.Time) error {
	p, err := marshalUUID(participantID, "participant id")
	if err != nil {
		return err
	}
	ignored, err := marshalUUID(ignoredDutyID, "duty id")
	if err != nil {
		return err
	}
	var marker int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM duty_participants p JOIN duty d ON d.id = p.duty_id WHERE p.participant_id = ? AND p.excluded_at IS NULL AND d.id <> ? AND d.start_date < ? AND d.end_date > ? LIMIT 1`, p, ignored, end, start).Scan(&marker)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("check participant duty overlap: %w", err)
	}
	return duty.ErrParticipantPeriodConflict
}
func scanParticipants(rows *sql.Rows) ([]duty.DutyParticipant, error) {
	out := []duty.DutyParticipant{}
	for rows.Next() {
		var d, p []byte
		var typ string
		var excluded sql.NullTime
		if err := rows.Scan(&d, &p, &typ, &excluded); err != nil {
			return nil, err
		}
		dutyID, err := uuid.FromBytes(d)
		if err != nil {
			return nil, err
		}
		participantID, err := uuid.FromBytes(p)
		if err != nil {
			return nil, err
		}
		item := duty.DutyParticipant{DutyID: dutyID, ParticipantID: participantID, Type: duty.ParticipantType(typ)}
		if excluded.Valid {
			v := excluded.Time
			item.ExcludedAt = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
