package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"dorm/pkg/core/ports/dto"
	queryports "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
)

type PenaltyQueryService struct {
	db *sql.DB
}

func NewPenaltyQueryService(db *sql.DB) *PenaltyQueryService {
	return &PenaltyQueryService{db: db}
}

func (q *PenaltyQueryService) ListResidentsWithPenaltyBalance(
	ctx context.Context,
	scope queryports.PenaltyScope,
) ([]dto.PenaltyResidentSummary, error) {
	baseArgs, scopeWhere, err := buildPenaltyScopeFilter(scope)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			u.id,
			TRIM(CONCAT(
				u.last_name, ' ',
				u.first_name,
				CASE
					WHEN u.middle_name IS NULL OR u.middle_name = '' THEN ''
					ELSE CONCAT(' ', u.middle_name)
				END
			)) AS full_name,
			COALESCE(SUM(
				CASE
					WHEN pe.type = 'ISSUE' THEN pe.weight
					WHEN pe.type = 'RESOLVE' THEN -pe.weight
					ELSE 0
				END
			), 0) AS total_weight
		FROM penalty_entry pe
		JOIN user u ON u.id = pe.user_id
		WHERE u.deleted_at IS NULL
		  AND ` + scopeWhere + `
		GROUP BY u.id, u.last_name, u.first_name, u.middle_name
		HAVING total_weight > 0
		ORDER BY total_weight DESC, u.last_name ASC, u.first_name ASC, u.middle_name ASC, u.id ASC
	`

	rows, err := q.db.QueryContext(ctx, query, baseArgs...)
	if err != nil {
		return nil, fmt.Errorf("list residents with penalties: %w", err)
	}
	defer rows.Close()

	items := make([]dto.PenaltyResidentSummary, 0)
	for rows.Next() {
		var idBytes []byte
		var item dto.PenaltyResidentSummary
		if err := rows.Scan(&idBytes, &item.FullName, &item.TotalWeight); err != nil {
			return nil, fmt.Errorf("scan resident with penalties: %w", err)
		}

		userID, err := uuid.FromBytes(idBytes)
		if err != nil {
			return nil, fmt.Errorf("decode penalty resident id: %w", err)
		}

		item.UserID = userID.String()
		item.ThresholdReached = item.TotalWeight >= 6
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate residents with penalties: %w", err)
	}

	return items, nil
}

func (q *PenaltyQueryService) SearchEligibleResidents(
	ctx context.Context,
	scope queryports.PenaltyScope,
	search string,
) ([]dto.PenaltyResidentOption, error) {
	baseArgs, scopeWhere, err := buildPenaltyScopeFilter(scope)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			u.id,
			TRIM(CONCAT(
				u.last_name, ' ',
				u.first_name,
				CASE
					WHEN u.middle_name IS NULL OR u.middle_name = '' THEN ''
					ELSE CONCAT(' ', u.middle_name)
				END
			)) AS full_name
		FROM user u
		WHERE u.deleted_at IS NULL
		  AND ` + scopeWhere

	search = strings.TrimSpace(strings.ToLower(search))
	args := append([]interface{}{}, baseArgs...)
	if search != "" {
		searchLike := "%" + search + "%"
		query += `
		  AND (
			LOWER(u.first_name) LIKE ?
			OR LOWER(u.last_name) LIKE ?
			OR LOWER(COALESCE(u.middle_name, '')) LIKE ?
			OR LOWER(CONCAT_WS(' ', u.last_name, u.first_name, NULLIF(u.middle_name, ''))) LIKE ?
			OR LOWER(CONCAT_WS(' ', u.first_name, u.last_name, NULLIF(u.middle_name, ''))) LIKE ?
			OR LOWER(CONCAT_WS(' ', u.last_name, u.first_name)) LIKE ?
			OR LOWER(CONCAT_WS(' ', u.first_name, u.last_name)) LIKE ?
		  )
		`
		args = append(args,
			searchLike,
			searchLike,
			searchLike,
			searchLike,
			searchLike,
			searchLike,
			searchLike,
		)
	}

	query += `
		ORDER BY u.last_name ASC, u.first_name ASC, u.middle_name ASC, u.id ASC
	`

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search eligible residents: %w", err)
	}
	defer rows.Close()

	items := make([]dto.PenaltyResidentOption, 0)
	for rows.Next() {
		var idBytes []byte
		var item dto.PenaltyResidentOption
		if err := rows.Scan(&idBytes, &item.Name); err != nil {
			return nil, fmt.Errorf("scan eligible resident: %w", err)
		}

		userID, err := uuid.FromBytes(idBytes)
		if err != nil {
			return nil, fmt.Errorf("decode eligible resident id: %w", err)
		}
		item.ID = userID.String()
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate eligible residents: %w", err)
	}

	return items, nil
}

func (q *PenaltyQueryService) ListPenaltyEntriesByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]dto.PenaltyEntryItem, error) {
	const query = `
		SELECT id, type, reason, weight, created_at
		FROM penalty_entry
		WHERE user_id = ?
		ORDER BY created_at DESC, id ASC
	`

	userIDBytes, err := userID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal penalty user id: %w", err)
	}

	rows, err := q.db.QueryContext(ctx, query, userIDBytes)
	if err != nil {
		return nil, fmt.Errorf("list penalty entries by user: %w", err)
	}
	defer rows.Close()

	items := make([]dto.PenaltyEntryItem, 0)
	for rows.Next() {
		var idBytes []byte
		var entryType string
		var item dto.PenaltyEntryItem
		var createdAt time.Time
		if err := rows.Scan(&idBytes, &entryType, &item.Reason, &item.Weight, &createdAt); err != nil {
			return nil, fmt.Errorf("scan penalty entry: %w", err)
		}

		entryID, err := uuid.FromBytes(idBytes)
		if err != nil {
			return nil, fmt.Errorf("decode penalty entry id: %w", err)
		}

		item.ID = entryID.String()
		item.Type = strings.ToLower(entryType)
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate penalty entries: %w", err)
	}

	return items, nil
}

func buildPenaltyScopeFilter(scope queryports.PenaltyScope) ([]interface{}, string, error) {
	if len(scope.DormitoryIDs) > 0 {
		placeholders := make([]string, len(scope.DormitoryIDs))
		args := make([]interface{}, 0, len(scope.DormitoryIDs))
		for index, dormitoryID := range scope.DormitoryIDs {
			placeholders[index] = "?"
			args = append(args, dormitoryID)
		}
		return args, "u.dormitory_id IN (" + strings.Join(placeholders, ", ") + ")", nil
	}

	return nil, "", fmt.Errorf("penalty scope is empty")
}
