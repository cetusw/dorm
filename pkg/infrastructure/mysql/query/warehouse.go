package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	warehousedomain "dorm/pkg/core/domain/warehouse"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type WarehouseQueryService struct {
	db *sql.DB
}

func NewWarehouseQueryService(db *sql.DB) *WarehouseQueryService {
	return &WarehouseQueryService{db: db}
}

func (q *WarehouseQueryService) ListItems(ctx context.Context, dormitoryID int64) ([]dto.WarehouseItem, error) {
	const query = `
		SELECT
			wi.id,
			wi.name,
			COALESCE(SUM(
				CASE wm.type
					WHEN 'add' THEN wm.quantity
					WHEN 'write-off' THEN -wm.quantity
				END
			), 0) AS quantity
		FROM warehouse_item wi
		LEFT JOIN warehouse_movement wm ON wm.item_id = wi.id
		WHERE wi.dormitory_id = ?
		GROUP BY wi.id, wi.name
		ORDER BY wi.name ASC
	`

	rows, err := q.db.QueryContext(ctx, query, dormitoryID)
	if err != nil {
		return nil, fmt.Errorf("query warehouse items: %w", err)
	}
	defer rows.Close()

	items := make([]dto.WarehouseItem, 0)
	for rows.Next() {
		var idBytes []byte
		var item dto.WarehouseItem
		if err := rows.Scan(&idBytes, &item.Name, &item.Quantity); err != nil {
			return nil, fmt.Errorf("scan warehouse item: %w", err)
		}

		itemID, err := uuid.FromBytes(idBytes)
		if err != nil {
			return nil, fmt.Errorf("decode warehouse item id: %w", err)
		}
		item.ID = itemID.String()
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate warehouse items: %w", err)
	}

	return items, nil
}

func (q *WarehouseQueryService) GetItemHistory(
	ctx context.Context,
	dormitoryID int64,
	itemID uuid.UUID,
) (*dto.WarehouseItemHistoryResponse, error) {
	item, err := q.loadHistoryItem(ctx, dormitoryID, itemID)
	if err != nil {
		return nil, err
	}

	movements, err := q.loadItemMovements(ctx, itemID)
	if err != nil {
		return nil, err
	}

	return &dto.WarehouseItemHistoryResponse{
		Item:      *item,
		Movements: movements,
	}, nil
}

func (q *WarehouseQueryService) loadHistoryItem(
	ctx context.Context,
	dormitoryID int64,
	itemID uuid.UUID,
) (*dto.WarehouseItem, error) {
	const query = `
		SELECT
			wi.id,
			wi.dormitory_id,
			wi.name,
			COALESCE(SUM(
				CASE wm.type
					WHEN 'add' THEN wm.quantity
					WHEN 'write-off' THEN -wm.quantity
				END
			), 0) AS quantity
		FROM warehouse_item wi
		LEFT JOIN warehouse_movement wm ON wm.item_id = wi.id
		WHERE wi.id = ?
		GROUP BY wi.id, wi.dormitory_id, wi.name
	`

	itemIDBytes, err := itemID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal warehouse item id: %w", err)
	}

	var idBytes []byte
	var storedDormitoryID int64
	var item dto.WarehouseItem
	if err := q.db.QueryRowContext(ctx, query, itemIDBytes).Scan(&idBytes, &storedDormitoryID, &item.Name, &item.Quantity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, warehousedomain.ErrItemNotFound
		}
		return nil, fmt.Errorf("query warehouse item history root: %w", err)
	}

	if storedDormitoryID != dormitoryID {
		return nil, warehousedomain.ErrAccessDenied
	}

	parsedID, err := uuid.FromBytes(idBytes)
	if err != nil {
		return nil, fmt.Errorf("decode warehouse item history root id: %w", err)
	}
	item.ID = parsedID.String()

	return &item, nil
}

func (q *WarehouseQueryService) loadItemMovements(ctx context.Context, itemID uuid.UUID) ([]dto.WarehouseMovement, error) {
	const query = `
		SELECT
			wm.id,
			wm.type,
			wm.quantity,
			wm.comment,
			wm.created_at,
			u.id,
			TRIM(CONCAT(
				u.last_name, ' ',
				u.first_name,
				CASE
					WHEN u.middle_name IS NULL OR u.middle_name = '' THEN ''
					ELSE CONCAT(' ', u.middle_name)
				END
			)) AS full_name,
			SUM(
				CASE wm.type
					WHEN 'add' THEN wm.quantity
					WHEN 'write-off' THEN -wm.quantity
				END
			) OVER (
				PARTITION BY wm.item_id
				ORDER BY wm.created_at ASC, wm.id ASC
				ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
			) AS balance_after
		FROM warehouse_movement wm
		INNER JOIN user u ON u.id = wm.created_by
		WHERE wm.item_id = ?
		ORDER BY wm.created_at ASC, wm.id ASC
	`

	itemIDBytes, err := itemID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal warehouse movement item id: %w", err)
	}

	rows, err := q.db.QueryContext(ctx, query, itemIDBytes)
	if err != nil {
		return nil, fmt.Errorf("query warehouse movements: %w", err)
	}
	defer rows.Close()

	movesAsc := make([]dto.WarehouseMovement, 0)
	for rows.Next() {
		var movementIDBytes []byte
		var createdByIDBytes []byte
		var createdAt time.Time
		var movement dto.WarehouseMovement
		if err := rows.Scan(
			&movementIDBytes,
			&movement.Type,
			&movement.Quantity,
			&movement.Comment,
			&createdAt,
			&createdByIDBytes,
			&movement.CreatedBy.FullName,
			&movement.BalanceAfter,
		); err != nil {
			return nil, fmt.Errorf("scan warehouse movement: %w", err)
		}

		movementID, err := uuid.FromBytes(movementIDBytes)
		if err != nil {
			return nil, fmt.Errorf("decode warehouse movement id: %w", err)
		}
		createdByID, err := uuid.FromBytes(createdByIDBytes)
		if err != nil {
			return nil, fmt.Errorf("decode warehouse movement actor id: %w", err)
		}

		movement.ID = movementID.String()
		movement.CreatedBy.ID = createdByID.String()
		movement.CreatedAt = createdAt.Format("2006-01-02T15:04:05Z07:00")
		movesAsc = append(movesAsc, movement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate warehouse movements: %w", err)
	}

	movements := make([]dto.WarehouseMovement, 0, len(movesAsc))
	for i := len(movesAsc) - 1; i >= 0; i-- {
		movements = append(movements, movesAsc[i])
	}

	return movements, nil
}
