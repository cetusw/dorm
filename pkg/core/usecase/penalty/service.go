package penalty

import (
	"context"
	"errors"
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
	teamRepo      structure.TeamRepository
	groupRepo     structure.GroupRepository
	dormitoryRepo structure.DormitoryRepository
	queryService  queryports.PenaltyQueryService
	location      *time.Location
	now           func() time.Time
}

func NewPenaltyService(
	penaltyRepo penaltydomain.Repository,
	userRepo user.Repository,
	teamRepo structure.TeamRepository,
	groupRepo structure.GroupRepository,
	dormitoryRepo structure.DormitoryRepository,
	queryService queryports.PenaltyQueryService,
	location *time.Location,
) *Service {
	return &Service{
		penaltyRepo:   penaltyRepo,
		userRepo:      userRepo,
		teamRepo:      teamRepo,
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

	items, err := s.queryService.ListResidentsWithActivePenalties(ctx, scope)
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

	items, err := s.queryService.ListActivePenaltiesByUser(ctx, resident.ID())
	if err != nil {
		return nil, fmt.Errorf("list resident penalties: %w", err)
	}

	return &dto.PenaltyResidentDetailsResponse{
		UserID:    resident.ID().String(),
		FullName:  buildUserFullName(resident),
		Penalties: items,
	}, nil
}

func (s *Service) CreatePenalty(
	ctx context.Context,
	currentUserID uuid.UUID,
	request dto.CreatePenaltyRequest,
) (*dto.PenaltyItem, error) {
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

	issuedOn, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(request.IssuedOn), s.location)
	if err != nil {
		return nil, penaltydomain.ErrInvalidIssuedOn
	}
	if isDateInFuture(issuedOn, s.now().In(s.location), s.location) {
		return nil, penaltydomain.ErrIssuedOnInFuture
	}

	createdAt := s.now().In(s.location)
	entity, err := penaltydomain.NewPenalty(
		resident.ID(),
		request.Weight,
		request.Reason,
		issuedOn,
		createdAt,
	)
	if err != nil {
		return nil, err
	}

	if err := s.penaltyRepo.Save(ctx, entity); err != nil {
		return nil, fmt.Errorf("create penalty: %w", err)
	}

	return penaltyItemFromDomain(entity), nil
}

func (s *Service) ResolvePenalty(
	ctx context.Context,
	currentUserID uuid.UUID,
	penaltyID uuid.UUID,
) error {
	scope, err := s.resolveScope(ctx, currentUserID)
	if err != nil {
		return err
	}

	entity, err := s.penaltyRepo.FindByID(ctx, penaltyID)
	if err != nil {
		return fmt.Errorf("find penalty: %w", err)
	}
	if entity == nil {
		return penaltydomain.ErrPenaltyNotFound
	}

	if _, err := s.loadAccessibleResident(ctx, scope, entity.UserID()); err != nil {
		return err
	}

	if entity.IsResolved() {
		return nil
	}

	entity.Resolve(s.now().In(s.location))
	if err := s.penaltyRepo.Save(ctx, entity); err != nil {
		return fmt.Errorf("resolve penalty: %w", err)
	}

	return nil
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

	groupIDs := make([]uuid.UUID, 0)
	for _, group := range groups {
		if group.LeaderID() != nil && *group.LeaderID() == currentUserID {
			groupIDs = append(groupIDs, group.ID())
		}
	}
	if len(groupIDs) == 0 {
		return queryports.PenaltyScope{}, penaltydomain.ErrAccessDenied
	}

	return queryports.PenaltyScope{GroupIDs: groupIDs}, nil
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

	if resident.TeamID() == nil {
		return nil, penaltydomain.ErrAccessDenied
	}

	team, err := s.teamRepo.FindByID(ctx, *resident.TeamID())
	if err != nil {
		return nil, fmt.Errorf("load resident team for penalty: %w", err)
	}
	if team == nil {
		return nil, penaltydomain.ErrAccessDenied
	}

	for _, groupID := range scope.GroupIDs {
		if team.GroupID() == groupID {
			return resident, nil
		}
	}

	return nil, penaltydomain.ErrAccessDenied
}

func penaltyItemFromDomain(entity *penaltydomain.Penalty) *dto.PenaltyItem {
	return &dto.PenaltyItem{
		ID:       entity.ID().String(),
		Reason:   entity.Reason(),
		Weight:   entity.Weight(),
		IssuedOn: entity.IssuedOn().Format("2006-01-02"),
	}
}

func buildUserFullName(resident *user.User) string {
	parts := []string{resident.LastName(), resident.FirstName()}
	if resident.MiddleName() != nil && strings.TrimSpace(*resident.MiddleName()) != "" {
		parts = append(parts, strings.TrimSpace(*resident.MiddleName()))
	}
	return strings.Join(parts, " ")
}

func isDateInFuture(value time.Time, now time.Time, location *time.Location) bool {
	issuedDate := time.Date(value.In(location).Year(), value.In(location).Month(), value.In(location).Day(), 0, 0, 0, 0, location)
	currentDate := time.Date(now.In(location).Year(), now.In(location).Month(), now.In(location).Day(), 0, 0, 0, 0, location)
	return issuedDate.After(currentDate)
}

func IsAccessError(err error) bool {
	return errors.Is(err, penaltydomain.ErrAccessDenied)
}
