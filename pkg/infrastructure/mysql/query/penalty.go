package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func (q *PenaltyQueryService) ListResidentsWithActivePenalties(
	ctx context.Context,
	scope queryports.PenaltyScope,
) ([]dto.PenaltyResidentSummary, error) {
	baseArgs, scopeJoin, scopeWhere, err := buildPenaltyScopeFilter(scope)
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
			COALESCE(SUM(p.weight), 0) AS total_weight
		FROM penalty p
		JOIN user u ON u.id = p.user_id
	` + scopeJoin + `
		WHERE p.resolved_at IS NULL
		  AND u.deleted_at IS NULL
		  AND ` + scopeWhere + `
		GROUP BY u.id, u.last_name, u.first_name, u.middle_name
		HAVING SUM(p.weight) > 0
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
	baseArgs, scopeJoin, scopeWhere, err := buildPenaltyScopeFilter(scope)
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
	` + scopeJoin + `
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

func (q *PenaltyQueryService) ListActivePenaltiesByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]dto.PenaltyItem, error) {
	const query = `
		SELECT id, reason, weight, DATE_FORMAT(issued_on, '%Y-%m-%d')
		FROM penalty
		WHERE user_id = ?
		  AND resolved_at IS NULL
		ORDER BY issued_on DESC, created_at DESC, id ASC
	`

	userIDBytes, err := userID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal penalty user id: %w", err)
	}

	rows, err := q.db.QueryContext(ctx, query, userIDBytes)
	if err != nil {
		return nil, fmt.Errorf("list active penalties by user: %w", err)
	}
	defer rows.Close()

	items := make([]dto.PenaltyItem, 0)
	for rows.Next() {
		var idBytes []byte
		var item dto.PenaltyItem
		var issuedOn string
		if err := rows.Scan(&idBytes, &item.Reason, &item.Weight, &issuedOn); err != nil {
			return nil, fmt.Errorf("scan active penalty: %w", err)
		}

		penaltyID, err := uuid.FromBytes(idBytes)
		if err != nil {
			return nil, fmt.Errorf("decode active penalty id: %w", err)
		}

		item.ID = penaltyID.String()
		item.IssuedOn = issuedOn
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active penalties: %w", err)
	}

	return items, nil
}

func buildPenaltyScopeFilter(scope queryports.PenaltyScope) ([]interface{}, string, string, error) {
	if len(scope.DormitoryIDs) > 0 && len(scope.GroupIDs) > 0 {
		return nil, "", "", fmt.Errorf("penalty scope cannot mix dormitory and group filters")
	}

	if len(scope.DormitoryIDs) > 0 {
		placeholders := make([]string, len(scope.DormitoryIDs))
		args := make([]interface{}, 0, len(scope.DormitoryIDs))
		for index, dormitoryID := range scope.DormitoryIDs {
			placeholders[index] = "?"
			args = append(args, dormitoryID)
		}
		return args, "", "u.dormitory_id IN (" + strings.Join(placeholders, ", ") + ")", nil
	}

	if len(scope.GroupIDs) > 0 {
		placeholders := make([]string, len(scope.GroupIDs))
		args := make([]interface{}, 0, len(scope.GroupIDs))
		for index, groupID := range scope.GroupIDs {
			groupIDBytes, err := groupID.MarshalBinary()
			if err != nil {
				return nil, "", "", fmt.Errorf("marshal scope group id: %w", err)
			}
			placeholders[index] = "?"
			args = append(args, groupIDBytes)
		}
		return args,
			"JOIN team t ON t.id = u.team_id",
			"t.group_id IN (" + strings.Join(placeholders, ", ") + ")",
			nil
	}

	return nil, "", "", fmt.Errorf("penalty scope is empty")
}
