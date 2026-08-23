package repository

import (
	"context"
	"database/sql"
	individual "dorm/pkg/core/domain/individualtask"
	penalty "dorm/pkg/core/domain/penalty"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sort"
	"time"
)

type IndividualTaskRepository struct{ db *sql.DB }

func NewIndividualTaskRepository(db *sql.DB) *IndividualTaskRepository {
	return &IndividualTaskRepository{db}
}

func (r *IndividualTaskRepository) FindByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*individual.IndividualTask, error) {
	q := taskSelect + " WHERE it.id = ?"
	if !includeDeleted {
		q += " AND it.deleted_at IS NULL"
	}
	return scanTask(r.db.QueryRowContext(ctx, q, mustUUID(id)))
}
func (r *IndividualTaskRepository) Create(ctx context.Context, t *individual.IndividualTask) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		if err := lockUsers(ctx, tx, []uuid.UUID{t.ResidentID}); err != nil {
			return err
		}
		if err := validateCapacity(ctx, tx, t.ResidentID, t.RedemptionWeight, 0); err != nil {
			return err
		}
		return insertTask(ctx, tx, t)
	})
}
func (r *IndividualTaskRepository) Update(ctx context.Context, t *individual.IndividualTask, expected uint64) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		cur, e := loadTaskForUpdate(ctx, tx, t.ID)
		if e != nil {
			return e
		}
		if cur.DeletedAt != nil {
			return individual.ErrNotFound
		}
		if cur.Version != expected {
			return individual.ErrVersionConflict
		}
		if !cur.CanEdit() {
			return individual.ErrInvalidTransition
		}
		ids := []uuid.UUID{cur.ResidentID, t.ResidentID}
		if e = lockUsers(ctx, tx, ids); e != nil {
			return e
		}
		own := 0.0
		if cur.ResidentID == t.ResidentID {
			own = cur.RedemptionWeight
		}
		if e = validateCapacity(ctx, tx, t.ResidentID, t.RedemptionWeight, own); e != nil {
			return e
		}
		_, e = tx.ExecContext(ctx, `UPDATE individual_task SET resident_id=?, area_id=?, title=?, redemption_weight=?, deadline=?, updated_at=?, version=version+1 WHERE id=? AND version=?`, mustUUID(t.ResidentID), nullableInt(t.AreaID), t.Title, fmt.Sprintf("%.1f", t.RedemptionWeight), nullableTime(t.Deadline), t.UpdatedAt, mustUUID(t.ID), expected)
		return e
	})
}
func (r *IndividualTaskRepository) Complete(ctx context.Context, id, resident uuid.UUID) (*individual.IndividualTask, error) {
	var out *individual.IndividualTask
	e := r.withTx(ctx, func(tx *sql.Tx) error {
		t, e := loadTaskForUpdate(ctx, tx, id)
		if e != nil {
			return e
		}
		if t.DeletedAt != nil {
			return individual.ErrNotFound
		}
		if t.ResidentID != resident {
			return individual.ErrAccessDenied
		}
		before := t.Version
		if e = t.Complete(time.Now()); e != nil {
			return e
		}
		if t.Version != before {
			if _, e = tx.ExecContext(ctx, `UPDATE individual_task SET status=?, completed_at=?, updated_at=?, version=? WHERE id=?`, t.Status, t.CompletedAt, t.UpdatedAt, t.Version, mustUUID(id)); e != nil {
				return e
			}
		}
		out = t
		return nil
	})
	return out, e
}
func (r *IndividualTaskRepository) Open(ctx context.Context, id, resident uuid.UUID) (*individual.IndividualTask, error) {
	return r.transition(ctx, id, func(t *individual.IndividualTask) error {
		if t.ResidentID != resident {
			return individual.ErrAccessDenied
		}
		return t.Reject(time.Now())
	})
}
func (r *IndividualTaskRepository) Reject(ctx context.Context, id uuid.UUID) (*individual.IndividualTask, error) {
	return r.transition(ctx, id, func(t *individual.IndividualTask) error { return t.Reject(time.Now()) })
}
func (r *IndividualTaskRepository) Verify(ctx context.Context, id uuid.UUID) (*individual.IndividualTask, error) {
	var out *individual.IndividualTask
	e := r.withTx(ctx, func(tx *sql.Tx) error {
		t, e := loadTaskForUpdate(ctx, tx, id)
		if e != nil {
			return e
		}
		if t.DeletedAt != nil {
			return individual.ErrNotFound
		}
		if t.Status == individual.StatusVerified {
			out = t
			return nil
		}
		if t.Status != individual.StatusCompleted {
			return individual.ErrInvalidTransition
		}
		if e = lockUsers(ctx, tx, []uuid.UUID{t.ResidentID}); e != nil {
			return e
		}
		if t.RedemptionWeight > 0 { // This task's own reservation is consumed by the resolve.
			if e = validateVerifyCapacity(ctx, tx, t); e != nil {
				return e
			}
			p, e := penalty.NewPenalty(t.ResidentID, penalty.EntryTypeResolve, t.RedemptionWeight, "Выполнил индивидуальную задачу: "+t.Title, time.Now())
			if e != nil {
				return e
			}
			if e = insertPenaltyEntry(ctx, tx, p); e != nil {
				return e
			}
		}
		if e = t.Verify(time.Now()); e != nil {
			return e
		}
		_, e = tx.ExecContext(ctx, `UPDATE individual_task SET status=?, verified_at=?, updated_at=?, version=? WHERE id=?`, t.Status, t.VerifiedAt, t.UpdatedAt, t.Version, mustUUID(id))
		out = t
		return e
	})
	return out, e
}
func (r *IndividualTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		t, e := loadTaskForUpdate(ctx, tx, id)
		if e != nil {
			return e
		}
		if t.DeletedAt != nil {
			return nil
		}
		if e = lockUsers(ctx, tx, []uuid.UUID{t.ResidentID}); e != nil {
			return e
		}
		if e = t.Delete(time.Now()); e != nil {
			return e
		}
		_, e = tx.ExecContext(ctx, `UPDATE individual_task SET deleted_at=?,updated_at=?,version=? WHERE id=?`, t.DeletedAt, t.UpdatedAt, t.Version, mustUUID(id))
		return e
	})
}
func (r *IndividualTaskRepository) transition(ctx context.Context, id uuid.UUID, fn func(*individual.IndividualTask) error) (*individual.IndividualTask, error) {
	var out *individual.IndividualTask
	e := r.withTx(ctx, func(tx *sql.Tx) error {
		t, e := loadTaskForUpdate(ctx, tx, id)
		if e != nil {
			return e
		}
		if t.DeletedAt != nil {
			return individual.ErrNotFound
		}
		before := t.Version
		if e = fn(t); e != nil {
			return e
		}
		if t.Version != before {
			_, e = tx.ExecContext(ctx, `UPDATE individual_task SET status=?,completed_at=?,updated_at=?,version=? WHERE id=?`, t.Status, nullableTime(t.CompletedAt), t.UpdatedAt, t.Version, mustUUID(id))
			if e != nil {
				return e
			}
		}
		out = t
		return nil
	})
	return out, e
}
func (r *IndividualTaskRepository) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, e := r.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	if e = fn(tx); e != nil {
		_ = tx.Rollback()
		return e
	}
	return tx.Commit()
}

func validateCapacity(ctx context.Context, tx *sql.Tx, user uuid.UUID, newWeight, own float64) error {
	b, e := penaltyBalance(ctx, tx, user)
	if e != nil {
		return e
	}
	reserved, e := reservedWeight(ctx, tx, user)
	if e != nil {
		return e
	}
	if newWeight > b-(reserved-own)+0.000001 {
		return individual.ErrCapacityExceeded
	}
	return nil
}
func validateVerifyCapacity(ctx context.Context, tx *sql.Tx, t *individual.IndividualTask) error {
	b, e := penaltyBalance(ctx, tx, t.ResidentID)
	if e != nil {
		return e
	}
	res, e := reservedWeight(ctx, tx, t.ResidentID)
	if e != nil {
		return e
	}
	if b-t.RedemptionWeight < res-t.RedemptionWeight-0.000001 {
		return individual.ErrCapacityExceeded
	}
	return nil
}
func reservedWeight(ctx context.Context, tx *sql.Tx, user uuid.UUID) (float64, error) {
	var v float64
	e := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(redemption_weight),0) FROM individual_task WHERE resident_id=? AND deleted_at IS NULL AND status IN ('ISSUED','COMPLETED')`, mustUUID(user)).Scan(&v)
	return v, e
}
func lockUsers(ctx context.Context, tx *sql.Tx, ids []uuid.UUID) error {
	sort.Slice(ids, func(i, j int) bool { return string(ids[i][:]) < string(ids[j][:]) })
	var prev uuid.UUID
	for _, id := range ids {
		if id == prev {
			continue
		}
		if e := lockPenaltyResident(ctx, tx, id); e != nil {
			return e
		}
		prev = id
	}
	return nil
}
func loadTaskForUpdate(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*individual.IndividualTask, error) {
	return scanTask(tx.QueryRowContext(ctx, taskSelect+` WHERE it.id=? FOR UPDATE`, mustUUID(id)))
}

const taskSelect = `SELECT it.id,it.dormitory_id,it.resident_id,it.area_id,it.title,it.redemption_weight,it.status,it.deadline,it.completed_at,it.verified_at,it.created_at,it.updated_at,it.deleted_at,it.version FROM individual_task it`

type individualTaskRowScanner interface{ Scan(...any) error }

func scanTask(row individualTaskRowScanner) (*individual.IndividualTask, error) {
	var t individual.IndividualTask
	var id, res []byte
	var area sql.NullInt64
	var dl, co, ve, de sql.NullTime
	var st string
	if e := row.Scan(&id, &t.DormitoryID, &res, &area, &t.Title, &t.RedemptionWeight, &st, &dl, &co, &ve, &t.CreatedAt, &t.UpdatedAt, &de, &t.Version); e != nil {
		if errors.Is(e, sql.ErrNoRows) {
			return nil, individual.ErrNotFound
		}
		return nil, fmt.Errorf("scan individual task: %w", e)
	}
	var e error
	if t.ID, e = uuid.FromBytes(id); e != nil {
		return nil, e
	}
	if t.ResidentID, e = uuid.FromBytes(res); e != nil {
		return nil, e
	}
	t.Status = individual.Status(st)
	if area.Valid {
		x := int(area.Int64)
		t.AreaID = &x
	}
	if dl.Valid {
		x := dl.Time
		t.Deadline = &x
	}
	if co.Valid {
		x := co.Time
		t.CompletedAt = &x
	}
	if ve.Valid {
		x := ve.Time
		t.VerifiedAt = &x
	}
	if de.Valid {
		x := de.Time
		t.DeletedAt = &x
	}
	return &t, nil
}
func insertTask(ctx context.Context, tx *sql.Tx, t *individual.IndividualTask) error {
	_, e := tx.ExecContext(ctx, `INSERT INTO individual_task (id,dormitory_id,resident_id,area_id,title,redemption_weight,status,deadline,created_at,updated_at,version) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, mustUUID(t.ID), t.DormitoryID, mustUUID(t.ResidentID), nullableInt(t.AreaID), t.Title, fmt.Sprintf("%.1f", t.RedemptionWeight), t.Status, nullableTime(t.Deadline), t.CreatedAt, t.UpdatedAt, t.Version)
	return e
}
func mustUUID(id uuid.UUID) []byte { b, _ := id.MarshalBinary(); return b }
func nullableInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}
func nullableTime(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}
