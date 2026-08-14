package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	warehousedomain "dorm/pkg/core/domain/warehouse"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type WarehouseRepository struct {
	db *sql.DB
}

func NewWarehouseRepository(db *sql.DB) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

func (r *WarehouseRepository) CreateItem(ctx context.Context, item *warehousedomain.Item, initialMovement *warehousedomain.Movement) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin warehouse create transaction: %w", err)
	}

	if err := insertWarehouseItem(ctx, tx, item); err != nil {
		_ = tx.Rollback()
		return err
	}

	if initialMovement != nil {
		if err := insertWarehouseMovement(ctx, tx, initialMovement); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit warehouse create transaction: %w", err)
	}

	return nil
}

func (r *WarehouseRepository) FindByID(ctx context.Context, dormitoryID int64, itemID uuid.UUID) (*warehousedomain.Item, error) {
	item, err := loadWarehouseItem(ctx, r.db, dormitoryID, itemID, false)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *WarehouseRepository) UpdateItem(ctx context.Context, item *warehousedomain.Item) error {
	const query = `
		UPDATE warehouse_item
		SET name = ?
		WHERE id = ? AND dormitory_id = ?
	`

	itemIDBytes, err := marshalUUID(item.ID(), "warehouse item id")
	if err != nil {
		return err
	}

	result, execErr := r.db.ExecContext(ctx, query, item.Name(), itemIDBytes, item.DormitoryID())
	if execErr != nil {
		if isWarehouseItemDuplicateError(execErr) {
			return warehousedomain.ErrDuplicateItemName
		}
		return fmt.Errorf("update warehouse item: %w", execErr)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("warehouse item update rows affected: %w", err)
	}
	if affected == 0 {
		return warehousedomain.ErrItemNotFound
	}

	return nil
}

func (r *WarehouseRepository) DeleteItem(ctx context.Context, dormitoryID int64, itemID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin warehouse delete transaction: %w", err)
	}

	item, err := loadWarehouseItem(ctx, tx, dormitoryID, itemID, true)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	balance, err := warehouseItemBalance(ctx, tx, item.ID())
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if balance != 0 {
		_ = tx.Rollback()
		return warehousedomain.ErrItemHasNonZeroBalance
	}

	itemIDBytes, err := marshalUUID(item.ID(), "warehouse item id")
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM warehouse_item WHERE id = ?`, itemIDBytes); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete warehouse item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit warehouse delete transaction: %w", err)
	}

	return nil
}

func (r *WarehouseRepository) AddMovement(
	ctx context.Context,
	dormitoryID int64,
	movement *warehousedomain.Movement,
) (*warehousedomain.Item, int64, error) {
	return r.applyMovement(ctx, dormitoryID, movement, false)
}

func (r *WarehouseRepository) WriteOffMovement(
	ctx context.Context,
	dormitoryID int64,
	movement *warehousedomain.Movement,
) (*warehousedomain.Item, int64, error) {
	return r.applyMovement(ctx, dormitoryID, movement, true)
}

func (r *WarehouseRepository) applyMovement(
	ctx context.Context,
	dormitoryID int64,
	movement *warehousedomain.Movement,
	checkBalance bool,
) (*warehousedomain.Item, int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("begin warehouse movement transaction: %w", err)
	}

	item, err := loadWarehouseItem(ctx, tx, dormitoryID, movement.ItemID(), true)
	if err != nil {
		_ = tx.Rollback()
		return nil, 0, err
	}

	balance, err := warehouseItemBalance(ctx, tx, item.ID())
	if err != nil {
		_ = tx.Rollback()
		return nil, 0, err
	}

	change := int64(movement.Quantity())
	if checkBalance {
		if balance < change {
			_ = tx.Rollback()
			return nil, 0, warehousedomain.ErrInsufficientItems
		}
		change = -change
	}

	if err := insertWarehouseMovement(ctx, tx, movement); err != nil {
		_ = tx.Rollback()
		return nil, 0, err
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("commit warehouse movement transaction: %w", err)
	}

	return item, balance + change, nil
}

type warehouseItemReader interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func loadWarehouseItem(
	ctx context.Context,
	reader warehouseItemReader,
	dormitoryID int64,
	itemID uuid.UUID,
	forUpdate bool,
) (*warehousedomain.Item, error) {
	query := `
		SELECT id, dormitory_id, name, created_at
		FROM warehouse_item
		WHERE id = ?
	`
	if forUpdate {
		query += " FOR UPDATE"
	}

	itemIDBytes, err := marshalUUID(itemID, "warehouse item id")
	if err != nil {
		return nil, err
	}

	var idBytes []byte
	var storedDormitoryID int64
	var name string
	var createdAt sql.NullTime
	if err := reader.QueryRowContext(ctx, query, itemIDBytes).Scan(&idBytes, &storedDormitoryID, &name, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, warehousedomain.ErrItemNotFound
		}
		return nil, fmt.Errorf("load warehouse item: %w", err)
	}

	if storedDormitoryID != dormitoryID {
		return nil, warehousedomain.ErrAccessDenied
	}

	parsedItemID, err := uuid.FromBytes(idBytes)
	if err != nil {
		return nil, fmt.Errorf("parse warehouse item id: %w", err)
	}

	return warehousedomain.RestoreItem(parsedItemID, storedDormitoryID, name, createdAt.Time), nil
}

func warehouseItemBalance(ctx context.Context, tx *sql.Tx, itemID uuid.UUID) (int64, error) {
	const query = `
		SELECT COALESCE(SUM(
			CASE type
				WHEN 'add' THEN quantity
				WHEN 'write-off' THEN -quantity
			END
		), 0)
		FROM warehouse_movement
		WHERE item_id = ?
	`

	itemIDBytes, err := marshalUUID(itemID, "warehouse item id")
	if err != nil {
		return 0, err
	}

	var balance int64
	if err := tx.QueryRowContext(ctx, query, itemIDBytes).Scan(&balance); err != nil {
		return 0, fmt.Errorf("calculate warehouse item balance: %w", err)
	}

	return balance, nil
}

func insertWarehouseItem(ctx context.Context, tx *sql.Tx, item *warehousedomain.Item) error {
	const query = `
		INSERT INTO warehouse_item (id, dormitory_id, name, created_at)
		VALUES (?, ?, ?, ?)
	`

	itemIDBytes, err := marshalUUID(item.ID(), "warehouse item id")
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, query, itemIDBytes, item.DormitoryID(), item.Name(), item.CreatedAt()); err != nil {
		if isWarehouseItemDuplicateError(err) {
			return warehousedomain.ErrDuplicateItemName
		}
		return fmt.Errorf("insert warehouse item: %w", err)
	}

	return nil
}

func insertWarehouseMovement(ctx context.Context, tx *sql.Tx, movement *warehousedomain.Movement) error {
	const query = `
		INSERT INTO warehouse_movement (id, item_id, type, quantity, comment, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	movementIDBytes, err := marshalUUID(movement.ID(), "warehouse movement id")
	if err != nil {
		return err
	}
	itemIDBytes, err := marshalUUID(movement.ItemID(), "warehouse movement item id")
	if err != nil {
		return err
	}
	createdByBytes, err := marshalUUID(movement.CreatedBy(), "warehouse movement created by")
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		query,
		movementIDBytes,
		itemIDBytes,
		string(movement.Type()),
		movement.Quantity(),
		movement.Comment(),
		createdByBytes,
		movement.CreatedAt(),
	); err != nil {
		return fmt.Errorf("insert warehouse movement: %w", err)
	}

	return nil
}

func isWarehouseItemDuplicateError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) {
		return false
	}

	return mysqlErr.Number == 1062 &&
		strings.Contains(strings.ToLower(mysqlErr.Message), "uq_warehouse_item_dormitory_name")
}
