package query

import (
	"context"
	"database/sql"
	"dorm/pkg/core/ports/dto"
	qport "dorm/pkg/core/ports/query"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

type IndividualTaskQueryService struct {
	db       *sql.DB
	location *time.Location
}

func NewIndividualTaskQueryService(db *sql.DB) *IndividualTaskQueryService {
	return &IndividualTaskQueryService{db: db, location: time.Local}
}

func (q *IndividualTaskQueryService) SetLocation(location *time.Location) {
	q.location = location
}

const taskRead = `SELECT it.id,it.dormitory_id,it.title,it.redemption_weight,it.status,it.deadline,it.completed_at,it.verified_at,it.created_at,it.updated_at,it.version,u.id,TRIM(CONCAT(u.last_name,' ',u.first_name,CASE WHEN u.middle_name IS NULL OR u.middle_name='' THEN '' ELSE CONCAT(' ',u.middle_name) END)),a.id,a.name,a.floor FROM individual_task it JOIN user u ON u.id=it.resident_id LEFT JOIN area a ON a.id=it.area_id`

func (q *IndividualTaskQueryService) Get(c context.Context, id uuid.UUID) (*dto.IndividualTaskItem, error) {
	x, e := q.list(c, taskRead+` WHERE it.id=? AND it.deleted_at IS NULL`, mustBinary(id))
	if e != nil || len(x) == 0 {
		return nil, e
	}
	return &x[0], nil
}
func (q *IndividualTaskQueryService) ListMine(c context.Context, id uuid.UUID) ([]dto.IndividualTaskItem, error) {
	return q.list(c, taskRead+` WHERE it.resident_id=? AND it.deleted_at IS NULL ORDER BY it.created_at DESC,it.id ASC`, mustBinary(id))
}
func (q *IndividualTaskQueryService) ListResident(c context.Context, id uuid.UUID, scope []int64) ([]dto.IndividualTaskItem, error) {
	a, w, e := scopeArgs(scope, "it.dormitory_id")
	if e != nil {
		return nil, e
	}
	a = append([]any{mustBinary(id)}, a...)
	return q.list(c, taskRead+` WHERE it.resident_id=? AND it.deleted_at IS NULL AND `+w+` ORDER BY it.created_at DESC,it.id ASC`, a...)
}
func (q *IndividualTaskQueryService) ListReview(c context.Context, scope []int64) ([]dto.IndividualTaskItem, error) {
	a, w, e := scopeArgs(scope, "it.dormitory_id")
	if e != nil {
		return nil, e
	}
	return q.list(c, taskRead+` WHERE it.deleted_at IS NULL AND it.status='COMPLETED' AND `+w+` ORDER BY it.completed_at ASC,it.id ASC`, a...)
}
func (q *IndividualTaskQueryService) list(c context.Context, s string, args ...any) ([]dto.IndividualTaskItem, error) {
	rows, e := q.db.QueryContext(c, s, args...)
	if e != nil {
		return nil, fmt.Errorf("query individual tasks: %w", e)
	}
	defer rows.Close()
	out := []dto.IndividualTaskItem{}
	for rows.Next() {
		var id, rid []byte
		var x dto.IndividualTaskItem
		var dl, co, ve sql.NullTime
		var aid sql.NullInt64
		var an sql.NullString
		var floor sql.NullInt64
		var st string
		if e = rows.Scan(&id, &x.DormitoryID, &x.Title, &x.RedemptionWeight, &st, &dl, &co, &ve, &x.CreatedAt, &x.UpdatedAt, &x.Version, &rid, &x.Resident.Name, &aid, &an, &floor); e != nil {
			return nil, e
		}
		uid, e := uuid.FromBytes(id)
		if e != nil {
			return nil, e
		}
		x.ID = uid.String()
		uid, e = uuid.FromBytes(rid)
		if e != nil {
			return nil, e
		}
		x.Resident.ID = uid.String()
		x.Status = strings.ToLower(st)
		x.Deadline = datePtr(dl)
		x.CompletedAt = timePtr(co)
		x.VerifiedAt = timePtr(ve)
		if aid.Valid {
			a := dto.IndividualTaskArea{ID: int(aid.Int64), Name: an.String}
			if floor.Valid {
				v := int(floor.Int64)
				a.Floor = &v
			}
			x.Area = &a
		}
		if dl.Valid {
			now := time.Now().In(q.location)
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, q.location)
			x.IsOverdue = st == "ISSUED" && dl.Time.Before(today)
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (q *IndividualTaskQueryService) SearchResidents(c context.Context, scope []int64, search string) ([]dto.IndividualTaskResidentOption, error) {
	a, w, e := scopeArgs(scope, "u.dormitory_id")
	if e != nil {
		return nil, e
	}
	search = strings.TrimSpace(strings.ToLower(search))
	s := `SELECT
		u.id,
		TRIM(CONCAT(u.last_name,' ',u.first_name,CASE WHEN u.middle_name IS NULL OR u.middle_name='' THEN '' ELSE CONCAT(' ',u.middle_name) END)),
		u.dormitory_id,
		COALESCE(penalties.balance, 0),
		COALESCE(reservations.weight, 0)
		FROM user u
		LEFT JOIN (
			SELECT user_id, SUM(CASE WHEN type='ISSUE' THEN weight ELSE -weight END) AS balance
			FROM penalty_entry
			GROUP BY user_id
		) penalties ON penalties.user_id=u.id
		LEFT JOIN (
			SELECT resident_id, SUM(redemption_weight) AS weight
			FROM individual_task
			WHERE deleted_at IS NULL AND status IN ('ISSUED','COMPLETED')
			GROUP BY resident_id
		) reservations ON reservations.resident_id=u.id
		WHERE u.deleted_at IS NULL AND ` + w
	if search != "" {
		s += ` AND LOWER(CONCAT_WS(' ',u.last_name,u.first_name,u.middle_name)) LIKE ?`
		a = append(a, "%"+search+"%")
	}
	s += ` ORDER BY u.last_name,u.first_name,u.middle_name,u.id`
	rows, e := q.db.QueryContext(c, s, a...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []dto.IndividualTaskResidentOption{}
	for rows.Next() {
		var b []byte
		var x dto.IndividualTaskResidentOption
		if e = rows.Scan(&b, &x.Name, &x.DormitoryID, &x.PenaltyBalance, &x.ReservedRedemptionWeight); e != nil {
			return nil, e
		}
		u, e := uuid.FromBytes(b)
		if e != nil {
			return nil, e
		}
		x.ID = u.String()
		x.AvailableRedemptionWeight = x.PenaltyBalance - x.ReservedRedemptionWeight
		if x.AvailableRedemptionWeight < 0 {
			x.AvailableRedemptionWeight = 0
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (q *IndividualTaskQueryService) ListAreas(c context.Context, dorm int64) ([]dto.IndividualTaskArea, error) {
	rows, e := q.db.QueryContext(c, "SELECT a.id,a.name,a.floor FROM area a LEFT JOIN `group` g ON g.id=a.group_id WHERE a.group_id IS NULL OR g.dormitory_id=? ORDER BY a.floor,a.name,a.id", dorm)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []dto.IndividualTaskArea{}
	for rows.Next() {
		var x dto.IndividualTaskArea
		var f sql.NullInt64
		if e = rows.Scan(&x.ID, &x.Name, &f); e != nil {
			return nil, e
		}
		if f.Valid {
			v := int(f.Int64)
			x.Floor = &v
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func scopeArgs(scope []int64, column string) ([]any, string, error) {
	if len(scope) == 0 {
		return nil, "", fmt.Errorf("empty individual task scope")
	}
	a := make([]any, len(scope))
	p := make([]string, len(scope))
	for i, v := range scope {
		a[i] = v
		p[i] = "?"
	}
	return a, column + " IN (" + strings.Join(p, ",") + ")", nil
}
func mustBinary(id uuid.UUID) []byte { b, _ := id.MarshalBinary(); return b }
func timePtr(v sql.NullTime) *string {
	if !v.Valid {
		return nil
	}
	s := v.Time.Format(time.RFC3339)
	return &s
}
func datePtr(v sql.NullTime) *string {
	if !v.Valid {
		return nil
	}
	s := v.Time.Format("2006-01-02")
	return &s
}

var _ qport.IndividualTaskQueryService = (*IndividualTaskQueryService)(nil)
