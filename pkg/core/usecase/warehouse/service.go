package warehouse

import (
	"context"
	"fmt"
	"math"
	"time"

	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	warehousedomain "dorm/pkg/core/domain/warehouse"
	"dorm/pkg/core/ports/dto"
	queryports "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
)

type Service struct {
	repo          warehousedomain.Repository
	userRepo      user.Repository
	groupRepo     structure.GroupRepository
	dormitoryRepo structure.DormitoryRepository
	queryService  queryports.WarehouseQueryService
	now           func() time.Time
	location      *time.Location
}

func NewWarehouseService(
	repo warehousedomain.Repository,
	userRepo user.Repository,
	groupRepo structure.GroupRepository,
	dormitoryRepo structure.DormitoryRepository,
	queryService queryports.WarehouseQueryService,
	location *time.Location,
) *Service {
	return &Service{
		repo:          repo,
		userRepo:      userRepo,
		groupRepo:     groupRepo,
		dormitoryRepo: dormitoryRepo,
		queryService:  queryService,
		now:           time.Now,
		location:      location,
	}
}

func (s *Service) GetWarehouse(ctx context.Context, currentUserID uuid.UUID) (*dto.WarehouseResponse, error) {
	dormitoryID, err := s.resolveManagedDormitoryID(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	items, err := s.queryService.ListItems(ctx, dormitoryID)
	if err != nil {
		return nil, fmt.Errorf("list warehouse items: %w", err)
	}

	return &dto.WarehouseResponse{Items: items}, nil
}

func (s *Service) GetItemHistory(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID) (*dto.WarehouseItemHistoryResponse, error) {
	dormitoryID, err := s.resolveManagedDormitoryID(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	response, err := s.queryService.GetItemHistory(ctx, dormitoryID, itemID)
	if err != nil {
		return nil, fmt.Errorf("get warehouse item history: %w", err)
	}

	return response, nil
}

func (s *Service) CreateItem(ctx context.Context, currentUserID uuid.UUID, request dto.CreateWarehouseItemRequest) (*dto.WarehouseItem, error) {
	dormitoryID, err := s.resolveManagedDormitoryID(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	quantity, err := normalizeCreateQuantity(request.Quantity)
	if err != nil {
		return nil, err
	}

	item, err := warehousedomain.NewItem(dormitoryID, request.Name)
	if err != nil {
		return nil, err
	}

	var movement *warehousedomain.Movement
	if quantity > 0 {
		movement, err = warehousedomain.NewMovement(
			item.ID(),
			warehousedomain.MovementTypeAdd,
			uint32(quantity),
			request.Comment,
			currentUserID,
			s.now().In(s.location),
		)
		if err != nil {
			return nil, err
		}
	}

	if err := s.repo.CreateItem(ctx, item, movement); err != nil {
		return nil, fmt.Errorf("create warehouse item: %w", err)
	}

	return &dto.WarehouseItem{
		ID:       item.ID().String(),
		Name:     item.Name(),
		Quantity: quantity,
	}, nil
}

func (s *Service) UpdateItem(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID, request dto.UpdateWarehouseItemRequest) (*dto.WarehouseItem, error) {
	dormitoryID, err := s.resolveManagedDormitoryID(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	item, err := s.repo.FindByID(ctx, dormitoryID, itemID)
	if err != nil {
		return nil, fmt.Errorf("find warehouse item: %w", err)
	}

	if err := item.Rename(request.Name); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("update warehouse item: %w", err)
	}

	history, err := s.queryService.GetItemHistory(ctx, dormitoryID, itemID)
	if err != nil {
		return nil, fmt.Errorf("load warehouse item after update: %w", err)
	}

	return &history.Item, nil
}

func (s *Service) DeleteItem(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID) error {
	dormitoryID, err := s.resolveManagedDormitoryID(ctx, currentUserID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteItem(ctx, dormitoryID, itemID); err != nil {
		return fmt.Errorf("delete warehouse item: %w", err)
	}

	return nil
}

func (s *Service) AddItems(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID, request dto.WarehouseMovementRequest) (*dto.WarehouseItem, error) {
	return s.applyMovement(ctx, currentUserID, itemID, request, warehousedomain.MovementTypeAdd)
}

func (s *Service) WriteOffItems(ctx context.Context, currentUserID uuid.UUID, itemID uuid.UUID, request dto.WarehouseMovementRequest) (*dto.WarehouseItem, error) {
	return s.applyMovement(ctx, currentUserID, itemID, request, warehousedomain.MovementTypeWriteOff)
}

func (s *Service) UpdateMovement(
	ctx context.Context,
	currentUserID uuid.UUID,
	itemID uuid.UUID,
	movementID uuid.UUID,
	request dto.UpdateWarehouseMovementRequest,
) (*dto.WarehouseItem, error) {
	dormitoryID, err := s.resolveManagedDormitoryID(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	quantity, err := normalizePositiveQuantity(request.Quantity)
	if err != nil {
		return nil, err
	}

	movement, err := warehousedomain.NewMovement(
		itemID,
		warehousedomain.MovementTypeAdd,
		uint32(quantity),
		request.Comment,
		currentUserID,
		s.now().In(s.location),
	)
	if err != nil {
		return nil, err
	}

	movement = warehousedomain.RestoreMovement(
		movementID,
		itemID,
		movement.Type(),
		movement.Quantity(),
		movement.Comment(),
		movement.CreatedBy(),
		movement.CreatedAt(),
	)

	item, balance, err := s.repo.UpdateMovement(ctx, dormitoryID, movement)
	if err != nil {
		return nil, fmt.Errorf("update warehouse movement: %w", err)
	}

	return &dto.WarehouseItem{
		ID:       item.ID().String(),
		Name:     item.Name(),
		Quantity: balance,
	}, nil
}

func (s *Service) DeleteMovement(
	ctx context.Context,
	currentUserID uuid.UUID,
	itemID uuid.UUID,
	movementID uuid.UUID,
) (*dto.WarehouseItem, error) {
	dormitoryID, err := s.resolveManagedDormitoryID(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	item, balance, err := s.repo.DeleteMovement(ctx, dormitoryID, itemID, movementID)
	if err != nil {
		return nil, fmt.Errorf("delete warehouse movement: %w", err)
	}

	return &dto.WarehouseItem{
		ID:       item.ID().String(),
		Name:     item.Name(),
		Quantity: balance,
	}, nil
}

func (s *Service) applyMovement(
	ctx context.Context,
	currentUserID uuid.UUID,
	itemID uuid.UUID,
	request dto.WarehouseMovementRequest,
	moveType warehousedomain.MovementType,
) (*dto.WarehouseItem, error) {
	dormitoryID, err := s.resolveManagedDormitoryID(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	quantity, err := normalizePositiveQuantity(request.Quantity)
	if err != nil {
		return nil, err
	}

	movement, err := warehousedomain.NewMovement(
		itemID,
		moveType,
		uint32(quantity),
		request.Comment,
		currentUserID,
		s.now().In(s.location),
	)
	if err != nil {
		return nil, err
	}

	var item *warehousedomain.Item
	var balance int64
	if moveType == warehousedomain.MovementTypeAdd {
		item, balance, err = s.repo.AddMovement(ctx, dormitoryID, movement)
	} else {
		item, balance, err = s.repo.WriteOffMovement(ctx, dormitoryID, movement)
	}
	if err != nil {
		return nil, fmt.Errorf("apply warehouse movement: %w", err)
	}

	return &dto.WarehouseItem{
		ID:       item.ID().String(),
		Name:     item.Name(),
		Quantity: balance,
	}, nil
}

func (s *Service) resolveManagedDormitoryID(ctx context.Context, currentUserID uuid.UUID) (int64, error) {
	currentUser, err := s.userRepo.FindByID(ctx, currentUserID)
	if err != nil {
		return 0, fmt.Errorf("load current user for warehouse: %w", err)
	}
	if currentUser == nil || currentUser.DormitoryID() == nil {
		return 0, warehousedomain.ErrAccessDenied
	}

	dormitory, err := s.dormitoryRepo.FindByID(ctx, *currentUser.DormitoryID())
	if err != nil {
		return 0, fmt.Errorf("load current dormitory for warehouse: %w", err)
	}
	if dormitory != nil && dormitory.LeaderID() != nil && *dormitory.LeaderID() == currentUserID {
		return *currentUser.DormitoryID(), nil
	}

	groups, err := s.groupRepo.FindByDormitoryID(ctx, *currentUser.DormitoryID())
	if err != nil {
		return 0, fmt.Errorf("load dormitory groups for warehouse: %w", err)
	}
	for _, group := range groups {
		if group.LeaderID() != nil && *group.LeaderID() == currentUserID {
			return *currentUser.DormitoryID(), nil
		}
	}

	return 0, warehousedomain.ErrAccessDenied
}

func normalizeCreateQuantity(quantity *int64) (int64, error) {
	if quantity == nil {
		return 0, nil
	}
	if *quantity < 0 || *quantity > math.MaxUint32 {
		return 0, warehousedomain.ErrInvalidMovementQuantity
	}
	return *quantity, nil
}

func normalizePositiveQuantity(quantity int64) (int64, error) {
	if quantity <= 0 || quantity > math.MaxUint32 {
		return 0, warehousedomain.ErrInvalidMovementQuantity
	}
	return quantity, nil
}
