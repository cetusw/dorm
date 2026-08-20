package penalty

import (
	"context"
	"fmt"
	"strings"
	"time"

	penaltydomain "dorm/pkg/core/domain/penalty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"
	queryports "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
)

type Service struct {
	penaltyRepo   penaltydomain.Repository
	userRepo      user.Repository
	groupRepo     structure.GroupRepository
	dormitoryRepo structure.DormitoryRepository
	queryService  queryports.PenaltyQueryService
	location      *time.Location
	now           func() time.Time
}

func NewPenaltyService(
	penaltyRepo penaltydomain.Repository,
	userRepo user.Repository,
	groupRepo structure.GroupRepository,
	dormitoryRepo structure.DormitoryRepository,
	queryService queryports.PenaltyQueryService,
	location *time.Location,
) *Service {
	return &Service{
		penaltyRepo:   penaltyRepo,
		userRepo:      userRepo,
		groupRepo:     groupRepo,
		dormitoryRepo: dormitoryRepo,
		queryService:  queryService,
		location:      location,
		now:           time.Now,
	}
}

func (s *Service) ListResidents(
	ctx context.Context,
	currentUserID uuid.UUID,
) (dto.PenaltyResidentsResponse, error) {
	scope, err := s.resolveScope(ctx, currentUserID)
	if err != nil {
		return dto.PenaltyResidentsResponse{}, err
	}

	items, err := s.queryService.ListResidentsWithPenaltyBalance(ctx, scope)
	if err != nil {
		return dto.PenaltyResidentsResponse{}, fmt.Errorf("list residents with penalties: %w", err)
	}

	return dto.PenaltyResidentsResponse{Residents: items}, nil
}

func (s *Service) SearchResidents(
	ctx context.Context,
	currentUserID uuid.UUID,
	search string,
) (dto.PenaltyResidentOptionsResponse, error) {
	scope, err := s.resolveScope(ctx, currentUserID)
	if err != nil {
		return dto.PenaltyResidentOptionsResponse{}, err
	}

	items, err := s.queryService.SearchEligibleResidents(ctx, scope, search)
	if err != nil {
		return dto.PenaltyResidentOptionsResponse{}, fmt.Errorf("search residents for penalty: %w", err)
	}

	return dto.PenaltyResidentOptionsResponse{Residents: items}, nil
}

func (s *Service) GetResidentPenalties(
	ctx context.Context,
	currentUserID uuid.UUID,
	residentID uuid.UUID,
) (*dto.PenaltyResidentDetailsResponse, error) {
	scope, err := s.resolveScope(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	resident, err := s.loadAccessibleResident(ctx, scope, residentID)
	if err != nil {
		return nil, err
	}

	items, err := s.queryService.ListPenaltyEntriesByUser(ctx, resident.ID())
	if err != nil {
		return nil, fmt.Errorf("list resident penalties: %w", err)
	}

	return &dto.PenaltyResidentDetailsResponse{
		UserID:      resident.ID().String(),
		FullName:    buildUserFullName(resident),
		TotalWeight: penaltyBalanceFromEntries(items),
		Entries:     items,
	}, nil
}

func (s *Service) GetCurrentUserPenalties(
	ctx context.Context,
	currentUserID uuid.UUID,
) (*dto.CurrentUserPenaltiesResponse, error) {
	currentUser, err := s.userRepo.FindByID(ctx, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("load current user for penalties: %w", err)
	}
	if currentUser == nil {
		return nil, penaltydomain.ErrAccessDenied
	}

	items, err := s.queryService.ListPenaltyEntriesByUser(ctx, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("list current user penalties: %w", err)
	}

	return &dto.CurrentUserPenaltiesResponse{
		TotalWeight: penaltyBalanceFromEntries(items),
		Entries:     items,
	}, nil
}

func (s *Service) CreatePenalty(
	ctx context.Context,
	currentUserID uuid.UUID,
	request dto.CreatePenaltyRequest,
) (*dto.PenaltyEntryItem, error) {
	scope, err := s.resolveScope(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	residentID, err := uuid.Parse(strings.TrimSpace(request.UserID))
	if err != nil {
		return nil, penaltydomain.ErrInvalidResidentID
	}

	resident, err := s.loadAccessibleResident(ctx, scope, residentID)
	if err != nil {
		return nil, err
	}

	createdAt := s.now().In(s.location)
	entity, err := penaltydomain.NewPenalty(
		resident.ID(),
		penaltydomain.EntryTypeIssue,
		request.Weight,
		request.Reason,
		createdAt,
	)
	if err != nil {
		return nil, err
	}

	if err := s.penaltyRepo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("create penalty: %w", err)
	}

	return penaltyItemFromDomain(entity), nil
}

func (s *Service) ResolvePenalty(
	ctx context.Context,
	currentUserID uuid.UUID,
	request dto.ResolvePenaltyRequest,
) error {
	scope, err := s.resolveScope(ctx, currentUserID)
	if err != nil {
		return err
	}

	residentID, err := uuid.Parse(strings.TrimSpace(request.UserID))
	if err != nil {
		return penaltydomain.ErrInvalidResidentID
	}

	resident, err := s.loadAccessibleResident(ctx, scope, residentID)
	if err != nil {
		return err
	}

	entity, err := penaltydomain.NewPenalty(
		resident.ID(),
		penaltydomain.EntryTypeResolve,
		request.Weight,
		request.Reason,
		s.now().In(s.location),
	)
	if err != nil {
		return err
	}

	if err := s.penaltyRepo.CreateResolve(ctx, entity); err != nil {
		return fmt.Errorf("resolve penalty: %w", err)
	}

	return nil
}

func (s *Service) DeletePenaltyEntry(
	ctx context.Context,
	currentUserID uuid.UUID,
	entryID uuid.UUID,
) error {
	scope, err := s.resolveScope(ctx, currentUserID)
	if err != nil {
		return err
	}

	entry, err := s.penaltyRepo.FindByID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("find penalty entry: %w", err)
	}
	if entry == nil {
		return penaltydomain.ErrPenaltyEntryNotFound
	}

	if _, err := s.loadAccessibleResident(ctx, scope, entry.UserID()); err != nil {
		return err
	}

	if err := s.penaltyRepo.Delete(ctx, entryID); err != nil {
		return fmt.Errorf("delete penalty entry: %w", err)
	}

	return nil
}

func (s *Service) UpdatePenaltyEntry(
	ctx context.Context,
	currentUserID uuid.UUID,
	entryID uuid.UUID,
	request dto.UpdatePenaltyEntryRequest,
) (*dto.PenaltyEntryItem, error) {
	scope, err := s.resolveScope(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	entry, err := s.penaltyRepo.FindByID(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("find penalty entry: %w", err)
	}
	if entry == nil {
		return nil, penaltydomain.ErrPenaltyEntryNotFound
	}

	if _, err := s.loadAccessibleResident(ctx, scope, entry.UserID()); err != nil {
		return nil, err
	}

	updatedEntry, err := penaltydomain.RestorePenalty(
		entry.ID(),
		entry.UserID(),
		entry.Type(),
		request.Weight,
		request.Reason,
		entry.CreatedAt(),
	)
	if err != nil {
		return nil, err
	}

	if err := s.penaltyRepo.Update(ctx, updatedEntry); err != nil {
		return nil, fmt.Errorf("update penalty entry: %w", err)
	}

	return penaltyItemFromDomain(updatedEntry), nil
}

func (s *Service) resolveScope(ctx context.Context, currentUserID uuid.UUID) (queryports.PenaltyScope, error) {
	currentUser, err := s.userRepo.FindByID(ctx, currentUserID)
	if err != nil {
		return queryports.PenaltyScope{}, fmt.Errorf("load current user: %w", err)
	}
	if currentUser == nil {
		return queryports.PenaltyScope{}, penaltydomain.ErrAccessDenied
	}

	dormitories, err := s.dormitoryRepo.FindAll(ctx)
	if err != nil {
		return queryports.PenaltyScope{}, fmt.Errorf("load dormitories for penalty scope: %w", err)
	}

	dormitoryIDs := make([]int64, 0)
	for _, dormitory := range dormitories {
		if dormitory.LeaderID() != nil && *dormitory.LeaderID() == currentUserID {
			dormitoryIDs = append(dormitoryIDs, dormitory.ID())
		}
	}
	if len(dormitoryIDs) > 0 {
		return queryports.PenaltyScope{DormitoryIDs: dormitoryIDs}, nil
	}

	if currentUser.DormitoryID() == nil {
		return queryports.PenaltyScope{}, penaltydomain.ErrAccessDenied
	}

	groups, err := s.groupRepo.FindByDormitoryID(ctx, *currentUser.DormitoryID())
	if err != nil {
		return queryports.PenaltyScope{}, fmt.Errorf("load groups for penalty scope: %w", err)
	}

	for _, group := range groups {
		if group.LeaderID() != nil && *group.LeaderID() == currentUserID {
			return queryports.PenaltyScope{DormitoryIDs: []int64{*currentUser.DormitoryID()}}, nil
		}
	}

	return queryports.PenaltyScope{}, penaltydomain.ErrAccessDenied
}

func (s *Service) loadAccessibleResident(
	ctx context.Context,
	scope queryports.PenaltyScope,
	residentID uuid.UUID,
) (*user.User, error) {
	resident, err := s.userRepo.FindByID(ctx, residentID)
	if err != nil {
		return nil, fmt.Errorf("load resident for penalty: %w", err)
	}
	if resident == nil {
		return nil, penaltydomain.ErrResidentNotFound
	}

	if len(scope.DormitoryIDs) > 0 {
		if resident.DormitoryID() == nil {
			return nil, penaltydomain.ErrAccessDenied
		}
		for _, dormitoryID := range scope.DormitoryIDs {
			if *resident.DormitoryID() == dormitoryID {
				return resident, nil
			}
		}
		return nil, penaltydomain.ErrAccessDenied
	}

	if resident.DormitoryID() == nil {
		return nil, penaltydomain.ErrAccessDenied
	}

	for _, dormitoryID := range scope.DormitoryIDs {
		if *resident.DormitoryID() == dormitoryID {
			return resident, nil
		}
	}

	return nil, penaltydomain.ErrAccessDenied
}

func penaltyItemFromDomain(entity *penaltydomain.Penalty) *dto.PenaltyEntryItem {
	return &dto.PenaltyEntryItem{
		ID:        entity.ID().String(),
		Type:      strings.ToLower(string(entity.Type())),
		Reason:    entity.Reason(),
		Weight:    entity.Weight(),
		CreatedAt: entity.CreatedAt().Format(time.RFC3339),
	}
}

func buildUserFullName(resident *user.User) string {
	parts := []string{resident.LastName(), resident.FirstName()}
	if resident.MiddleName() != nil && strings.TrimSpace(*resident.MiddleName()) != "" {
		parts = append(parts, strings.TrimSpace(*resident.MiddleName()))
	}
	return strings.Join(parts, " ")
}

func penaltyBalanceFromEntries(entries []dto.PenaltyEntryItem) float64 {
	balance := 0.0
	for _, entry := range entries {
		switch entry.Type {
		case "resolve":
			balance -= entry.Weight
		default:
			balance += entry.Weight
		}
	}
	return balance
}
