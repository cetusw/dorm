package warehouse

import (
	"context"
	"testing"
	"time"

	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	warehousedomain "dorm/pkg/core/domain/warehouse"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type warehouseRepositoryStub struct {
	item             *warehousedomain.Item
	createItem       *warehousedomain.Item
	createMovement   *warehousedomain.Movement
	updateItem       *warehousedomain.Item
	addedMovement    *warehousedomain.Movement
	writeOffMovement *warehousedomain.Movement
	addBalance       int64
	writeOffBalance  int64
	createErr        error
	findErr          error
	updateErr        error
	addErr           error
	writeOffErr      error
	deleteErr        error
}

func (s *warehouseRepositoryStub) CreateItem(_ context.Context, item *warehousedomain.Item, movement *warehousedomain.Movement) error {
	s.createItem = item
	s.createMovement = movement
	return s.createErr
}

func (s *warehouseRepositoryStub) FindByID(context.Context, int64, uuid.UUID) (*warehousedomain.Item, error) {
	return s.item, s.findErr
}

func (s *warehouseRepositoryStub) UpdateItem(_ context.Context, item *warehousedomain.Item) error {
	s.updateItem = item
	return s.updateErr
}

func (s *warehouseRepositoryStub) DeleteItem(context.Context, int64, uuid.UUID) error {
	return s.deleteErr
}

func (s *warehouseRepositoryStub) AddMovement(_ context.Context, _ int64, movement *warehousedomain.Movement) (*warehousedomain.Item, int64, error) {
	s.addedMovement = movement
	return s.item, s.addBalance, s.addErr
}

func (s *warehouseRepositoryStub) WriteOffMovement(_ context.Context, _ int64, movement *warehousedomain.Movement) (*warehousedomain.Item, int64, error) {
	s.writeOffMovement = movement
	return s.item, s.writeOffBalance, s.writeOffErr
}

type warehouseUserRepositoryStub struct {
	user *user.User
}

func (s *warehouseUserRepositoryStub) Save(context.Context, *user.User) error { return nil }
func (s *warehouseUserRepositoryStub) FindAll(context.Context) ([]*user.User, error) {
	return nil, nil
}
func (s *warehouseUserRepositoryStub) FindByID(context.Context, uuid.UUID) (*user.User, error) {
	return s.user, nil
}
func (s *warehouseUserRepositoryStub) FindByLogin(context.Context, string) (*user.User, error) {
	return nil, nil
}
func (s *warehouseUserRepositoryStub) FindByTeamID(context.Context, uuid.UUID) ([]*user.User, error) {
	return nil, nil
}
func (s *warehouseUserRepositoryStub) FindByDormitoryID(context.Context, int64) ([]*user.User, error) {
	return nil, nil
}
func (s *warehouseUserRepositoryStub) MoveUserToTeam(context.Context, uuid.UUID, *uuid.UUID) error {
	return nil
}
func (s *warehouseUserRepositoryStub) SoftDelete(context.Context, uuid.UUID) error { return nil }

type warehouseGroupRepositoryStub struct {
	groups []*structure.Group
}

func (s *warehouseGroupRepositoryStub) FindAll(context.Context) ([]*structure.Group, error) {
	return nil, nil
}
func (s *warehouseGroupRepositoryStub) FindByID(context.Context, uuid.UUID) (*structure.Group, error) {
	return nil, nil
}
func (s *warehouseGroupRepositoryStub) FindByDormitoryID(context.Context, int64) ([]*structure.Group, error) {
	return s.groups, nil
}
func (s *warehouseGroupRepositoryStub) Save(context.Context, *structure.Group) error { return nil }
func (s *warehouseGroupRepositoryStub) Delete(context.Context, uuid.UUID) error      { return nil }

type warehouseDormitoryRepositoryStub struct {
	dormitory *structure.Dormitory
}

func (s *warehouseDormitoryRepositoryStub) FindAll(context.Context) ([]*structure.Dormitory, error) {
	return nil, nil
}
func (s *warehouseDormitoryRepositoryStub) FindByID(context.Context, int64) (*structure.Dormitory, error) {
	return s.dormitory, nil
}
func (s *warehouseDormitoryRepositoryStub) ExistsByLeaderID(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (s *warehouseDormitoryRepositoryStub) Save(context.Context, *structure.Dormitory) error {
	return nil
}
func (s *warehouseDormitoryRepositoryStub) Delete(context.Context, int64) error { return nil }

type warehouseQueryServiceStub struct {
	history *dto.WarehouseItemHistoryResponse
}

func (s *warehouseQueryServiceStub) ListItems(context.Context, int64) ([]dto.WarehouseItem, error) {
	return nil, nil
}
func (s *warehouseQueryServiceStub) GetItemHistory(context.Context, int64, uuid.UUID) (*dto.WarehouseItemHistoryResponse, error) {
	return s.history, nil
}

func TestCreateItemCreatesInitialMovement(t *testing.T) {
	t.Parallel()

	currentUserID := uuid.New()
	dormitoryID := int64(7)
	now := time.Date(2026, time.August, 14, 12, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	repo := &warehouseRepositoryStub{}

	service := NewWarehouseService(
		repo,
		&warehouseUserRepositoryStub{user: user.RestoreUser(currentUserID, "leader", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)},
		&warehouseGroupRepositoryStub{},
		&warehouseDormitoryRepositoryStub{dormitory: structure.RestoreDormitory(dormitoryID, "Dorm", &currentUserID, "Moscow", "st", "Lenina", "1")},
		&warehouseQueryServiceStub{},
		now.Location(),
	)
	service.now = func() time.Time { return now }

	response, err := service.CreateItem(context.Background(), currentUserID, dto.CreateWarehouseItemRequest{
		Name:     "  Лампа E27  ",
		Quantity: ptrInt64(10),
		Comment:  ptrString("  Первая закупка "),
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, repo.createItem)
	require.NotNil(t, repo.createMovement)
	assert.Equal(t, "Лампа E27", repo.createItem.Name())
	assert.Equal(t, int64(10), response.Quantity)
	require.NotNil(t, repo.createMovement.Comment())
	assert.Equal(t, "Первая закупка", *repo.createMovement.Comment())
}

func TestWriteOffRejectsInsufficientBalance(t *testing.T) {
	t.Parallel()

	currentUserID := uuid.New()
	dormitoryID := int64(7)
	now := time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)
	item := warehousedomain.RestoreItem(uuid.New(), dormitoryID, "Лампа", now)
	repo := &warehouseRepositoryStub{
		item:        item,
		writeOffErr: warehousedomain.ErrInsufficientItems,
	}

	service := NewWarehouseService(
		repo,
		&warehouseUserRepositoryStub{user: user.RestoreUser(currentUserID, "leader", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)},
		&warehouseGroupRepositoryStub{},
		&warehouseDormitoryRepositoryStub{dormitory: structure.RestoreDormitory(dormitoryID, "Dorm", &currentUserID, "Moscow", "st", "Lenina", "1")},
		&warehouseQueryServiceStub{},
		time.UTC,
	)

	_, err := service.WriteOffItems(context.Background(), currentUserID, item.ID(), dto.WarehouseMovementRequest{
		Quantity: 4,
	})

	assert.ErrorIs(t, err, warehousedomain.ErrInsufficientItems)
}

func TestGetWarehouseDeniedForResidentWithoutLeadership(t *testing.T) {
	t.Parallel()

	currentUserID := uuid.New()
	dormitoryID := int64(7)
	now := time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)

	service := NewWarehouseService(
		&warehouseRepositoryStub{},
		&warehouseUserRepositoryStub{user: user.RestoreUser(currentUserID, "resident", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)},
		&warehouseGroupRepositoryStub{},
		&warehouseDormitoryRepositoryStub{dormitory: structure.RestoreDormitory(dormitoryID, "Dorm", nil, "Moscow", "st", "Lenina", "1")},
		&warehouseQueryServiceStub{},
		time.UTC,
	)

	_, err := service.GetWarehouse(context.Background(), currentUserID)

	assert.ErrorIs(t, err, warehousedomain.ErrAccessDenied)
}

func ptrInt64(value int64) *int64 {
	return &value
}

func ptrString(value string) *string {
	return &value
}
