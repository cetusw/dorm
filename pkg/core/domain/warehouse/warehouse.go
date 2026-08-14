package warehouse

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAccessDenied            = errors.New("warehouse access denied")
	ErrItemNotFound            = errors.New("warehouse item not found")
	ErrDuplicateItemName       = errors.New("warehouse item name already exists")
	ErrInvalidItemName         = errors.New("warehouse item name is invalid")
	ErrInvalidMovementQuantity = errors.New("warehouse movement quantity is invalid")
	ErrInvalidMovementType     = errors.New("warehouse movement type is invalid")
	ErrCommentTooLong          = errors.New("warehouse movement comment is too long")
	ErrInsufficientItems       = errors.New("warehouse item balance is insufficient")
	ErrItemHasNonZeroBalance   = errors.New("warehouse item has non-zero balance")
)

const (
	maxItemNameLength = 255
	maxCommentLength  = 256
)

type MovementType string

const (
	MovementTypeAdd      MovementType = "add"
	MovementTypeWriteOff MovementType = "write-off"
)

type Item struct {
	id          uuid.UUID
	dormitoryID int64
	name        string
	createdAt   time.Time
}

func NewItem(dormitoryID int64, name string) (*Item, error) {
	normalized, err := normalizeItemName(name)
	if err != nil {
		return nil, err
	}

	return &Item{
		id:          uuid.New(),
		dormitoryID: dormitoryID,
		name:        normalized,
		createdAt:   time.Now(),
	}, nil
}

func RestoreItem(id uuid.UUID, dormitoryID int64, name string, createdAt time.Time) *Item {
	return &Item{
		id:          id,
		dormitoryID: dormitoryID,
		name:        name,
		createdAt:   createdAt,
	}
}

func (i *Item) Rename(name string) error {
	normalized, err := normalizeItemName(name)
	if err != nil {
		return err
	}
	i.name = normalized
	return nil
}

func (i *Item) ID() uuid.UUID      { return i.id }
func (i *Item) DormitoryID() int64 { return i.dormitoryID }
func (i *Item) Name() string       { return i.name }
func (i *Item) CreatedAt() time.Time {
	return i.createdAt
}

type Movement struct {
	id        uuid.UUID
	itemID    uuid.UUID
	moveType  MovementType
	quantity  uint32
	comment   *string
	createdBy uuid.UUID
	createdAt time.Time
}

func NewMovement(
	itemID uuid.UUID,
	moveType MovementType,
	quantity uint32,
	comment *string,
	createdBy uuid.UUID,
	createdAt time.Time,
) (*Movement, error) {
	if !moveType.IsValid() {
		return nil, ErrInvalidMovementType
	}
	if quantity == 0 {
		return nil, ErrInvalidMovementQuantity
	}

	normalizedComment, err := normalizeComment(comment)
	if err != nil {
		return nil, err
	}

	return &Movement{
		id:        uuid.New(),
		itemID:    itemID,
		moveType:  moveType,
		quantity:  quantity,
		comment:   normalizedComment,
		createdBy: createdBy,
		createdAt: createdAt,
	}, nil
}

func RestoreMovement(
	id uuid.UUID,
	itemID uuid.UUID,
	moveType MovementType,
	quantity uint32,
	comment *string,
	createdBy uuid.UUID,
	createdAt time.Time,
) *Movement {
	return &Movement{
		id:        id,
		itemID:    itemID,
		moveType:  moveType,
		quantity:  quantity,
		comment:   comment,
		createdBy: createdBy,
		createdAt: createdAt,
	}
}

func (m MovementType) IsValid() bool {
	return m == MovementTypeAdd || m == MovementTypeWriteOff
}

func (m *Movement) ID() uuid.UUID        { return m.id }
func (m *Movement) ItemID() uuid.UUID    { return m.itemID }
func (m *Movement) Type() MovementType   { return m.moveType }
func (m *Movement) Quantity() uint32     { return m.quantity }
func (m *Movement) Comment() *string     { return m.comment }
func (m *Movement) CreatedBy() uuid.UUID { return m.createdBy }
func (m *Movement) CreatedAt() time.Time { return m.createdAt }

type Repository interface {
	CreateItem(ctx context.Context, item *Item, initialMovement *Movement) error
	FindByID(ctx context.Context, dormitoryID int64, itemID uuid.UUID) (*Item, error)
	UpdateItem(ctx context.Context, item *Item) error
	DeleteItem(ctx context.Context, dormitoryID int64, itemID uuid.UUID) error
	AddMovement(ctx context.Context, dormitoryID int64, movement *Movement) (*Item, int64, error)
	WriteOffMovement(ctx context.Context, dormitoryID int64, movement *Movement) (*Item, int64, error)
}

func normalizeItemName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len([]rune(trimmed)) > maxItemNameLength {
		return "", ErrInvalidItemName
	}
	return trimmed, nil
}

func normalizeComment(comment *string) (*string, error) {
	if comment == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*comment)
	if trimmed == "" {
		return nil, nil
	}
	if len([]rune(trimmed)) > maxCommentLength {
		return nil, ErrCommentTooLong
	}

	return &trimmed, nil
}
