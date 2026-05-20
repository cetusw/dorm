package dormitory

import (
	"context"
	"fmt"
	"strings"

	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type Service struct {
	dormitoryRepo structure.DormitoryRepository
	groupRepo     structure.GroupRepository
	teamRepo      structure.TeamRepository
	userRepo      user.Repository
}

func NewDormitoryService(
	dormitoryRepo structure.DormitoryRepository,
	groupRepo structure.GroupRepository,
	teamRepo structure.TeamRepository,
	userRepo user.Repository,
) *Service {
	return &Service{
		dormitoryRepo: dormitoryRepo,
		groupRepo:     groupRepo,
		teamRepo:      teamRepo,
		userRepo:      userRepo,
	}
}

func (s *Service) GetDormitories(ctx context.Context) ([]*structure.Dormitory, error) {
	return s.dormitoryRepo.FindAll(ctx)
}

func (s *Service) GetDormitoryByID(ctx context.Context, id int64) (*structure.Dormitory, error) {
	return s.dormitoryRepo.FindByID(ctx, id)
}

func (s *Service) GetDormitoriesList(ctx context.Context) ([]dto.DormitoryListItem, error) {
	dormitories, err := s.dormitoryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]dto.DormitoryListItem, 0, len(dormitories))
	userNames, err := s.userNames(ctx)
	if err != nil {
		return nil, err
	}
	for _, dormitory := range dormitories {
		groups, err := s.groupRepo.FindByDormitoryID(ctx, dormitory.ID())
		if err != nil {
			return nil, err
		}
		items = append(items, dto.DormitoryListItem{
			ID:          dormitory.ID(),
			Name:        dormitory.Name(),
			Address:     dormitory.Address(),
			LeaderName:  leaderNameFromMap(userNames, dormitory.LeaderID()),
			GroupsCount: len(groups),
		})
	}
	return items, nil
}

func (s *Service) CreateDormitory(ctx context.Context, req dto.UpsertDormitoryRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("dormitory name is required")
	}
	if strings.TrimSpace(req.City) == "" {
		return fmt.Errorf("dormitory city is required")
	}
	if strings.TrimSpace(req.StreetType) == "" {
		return fmt.Errorf("dormitory street type is required")
	}
	if strings.TrimSpace(req.StreetName) == "" {
		return fmt.Errorf("dormitory street name is required")
	}
	if strings.TrimSpace(req.HouseNumber) == "" {
		return fmt.Errorf("dormitory house number is required")
	}
	leaderID, err := parseOptionalUUID(req.LeaderID)
	if err != nil {
		return err
	}
	if err := s.requireUser(ctx, leaderID); err != nil {
		return err
	}

	dormitory := structure.RestoreDormitory(0, req.Name, leaderID, req.City, req.StreetType, req.StreetName, req.HouseNumber)
	return s.dormitoryRepo.Save(ctx, dormitory)
}

func (s *Service) UpdateDormitory(ctx context.Context, id int64, req dto.UpsertDormitoryRequest) error {
	current, err := s.dormitoryRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return fmt.Errorf("dormitory not found")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("dormitory name is required")
	}
	if strings.TrimSpace(req.City) == "" {
		return fmt.Errorf("dormitory city is required")
	}
	if strings.TrimSpace(req.StreetType) == "" {
		return fmt.Errorf("dormitory street type is required")
	}
	if strings.TrimSpace(req.StreetName) == "" {
		return fmt.Errorf("dormitory street name is required")
	}
	if strings.TrimSpace(req.HouseNumber) == "" {
		return fmt.Errorf("dormitory house number is required")
	}
	leaderID, err := parseOptionalUUID(req.LeaderID)
	if err != nil {
		return err
	}
	if err := s.requireUser(ctx, leaderID); err != nil {
		return err
	}

	return s.dormitoryRepo.Save(ctx, structure.RestoreDormitory(id, req.Name, leaderID, req.City, req.StreetType, req.StreetName, req.HouseNumber))
}

func (s *Service) DeleteDormitory(ctx context.Context, id int64) error {
	users, err := s.userRepo.FindByDormitoryID(ctx, id)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return fmt.Errorf("cannot delete dormitory with residents")
	}
	return s.dormitoryRepo.Delete(ctx, id)
}

func (s *Service) GetGroupsList(ctx context.Context, dormitoryID int64) ([]dto.GroupListItem, error) {
	groups, err := s.groupRepo.FindByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.GroupListItem, 0, len(groups))
	residentNames, err := s.dormitoryUserNames(ctx, dormitoryID)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		teams, err := s.teamRepo.FindByGroupID(ctx, group.ID())
		if err != nil {
			return nil, err
		}
		items = append(items, dto.GroupListItem{
			ID:             group.ID(),
			Name:           group.Name(),
			LeaderName:     leaderNameFromMap(residentNames, group.LeaderID()),
			SpreadsheetID:  group.SpreadsheetID(),
			SpreadsheetURL: spreadsheetURL(group.SpreadsheetID()),
			DormitoryID:    group.DormitoryID(),
			TeamsCount:     len(teams),
		})
	}
	return items, nil
}

func (s *Service) GetGroupByID(ctx context.Context, id uuid.UUID) (*structure.Group, error) {
	return s.groupRepo.FindByID(ctx, id)
}

func (s *Service) GetDormitoryUserOptions(ctx context.Context, dormitoryID int64) ([]dto.UserOption, error) {
	users, err := s.userRepo.FindByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return nil, err
	}
	options := make([]dto.UserOption, 0, len(users))
	for _, resident := range users {
		options = append(options, dto.UserOption{
			ID:       resident.ID(),
			FullName: fullName(resident),
		})
	}
	return options, nil
}

func (s *Service) GetUserOptions(ctx context.Context) ([]dto.UserOption, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	options := make([]dto.UserOption, 0, len(users))
	for _, resident := range users {
		options = append(options, dto.UserOption{
			ID:       resident.ID(),
			FullName: fullName(resident),
		})
	}
	return options, nil
}

func (s *Service) CreateGroup(ctx context.Context, dormitoryID int64, req dto.UpsertGroupRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("group name is required")
	}
	if _, err := s.requireDormitory(ctx, dormitoryID); err != nil {
		return err
	}
	leaderID, err := parseOptionalUUID(req.LeaderID)
	if err != nil {
		return err
	}
	if err := s.requireUserInDormitory(ctx, leaderID, dormitoryID); err != nil {
		return err
	}

	return s.groupRepo.Save(ctx, structure.NewGroup(req.Name, leaderID, req.SpreadsheetID, dormitoryID))
}

func (s *Service) UpdateGroup(ctx context.Context, id uuid.UUID, req dto.UpsertGroupRequest) error {
	current, err := s.groupRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return fmt.Errorf("group not found")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("group name is required")
	}

	leaderID, err := parseOptionalUUID(req.LeaderID)
	if err != nil {
		return err
	}
	if err := s.requireUserInDormitory(ctx, leaderID, current.DormitoryID()); err != nil {
		return err
	}

	updated := structure.RestoreGroup(id, leaderID, req.Name, req.SpreadsheetID, current.DormitoryID(), current.NextDutyTeam())
	return s.groupRepo.Save(ctx, updated)
}

func (s *Service) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return s.groupRepo.Delete(ctx, id)
}

func parseOptionalUUID(rawID string) (*uuid.UUID, error) {
	if strings.TrimSpace(rawID) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	return &id, nil
}

func spreadsheetURL(spreadsheetID string) string {
	if strings.TrimSpace(spreadsheetID) == "" {
		return ""
	}
	return "https://docs.google.com/spreadsheets/d/" + spreadsheetID
}

func fullName(resident *user.User) string {
	return strings.TrimSpace(fmt.Sprintf("%s %s", resident.LastName(), resident.FirstName()))
}

func (s *Service) requireDormitory(ctx context.Context, id int64) (*structure.Dormitory, error) {
	dormitory, err := s.dormitoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dormitory == nil {
		return nil, fmt.Errorf("dormitory not found")
	}
	return dormitory, nil
}

func (s *Service) requireUser(ctx context.Context, userID *uuid.UUID) error {
	if userID == nil {
		return nil
	}
	resident, err := s.userRepo.FindByID(ctx, *userID)
	if err != nil {
		return err
	}
	if resident == nil {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (s *Service) requireUserInDormitory(ctx context.Context, userID *uuid.UUID, dormitoryID int64) error {
	if userID == nil {
		return nil
	}
	resident, err := s.userRepo.FindByID(ctx, *userID)
	if err != nil {
		return err
	}
	if resident == nil {
		return fmt.Errorf("user not found")
	}
	if resident.DormitoryID() == nil || *resident.DormitoryID() != dormitoryID {
		return fmt.Errorf("user does not belong to dormitory")
	}
	return nil
}

func (s *Service) userNames(ctx context.Context) (map[uuid.UUID]string, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return userNames(users), nil
}

func (s *Service) dormitoryUserNames(ctx context.Context, dormitoryID int64) (map[uuid.UUID]string, error) {
	users, err := s.userRepo.FindByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return nil, err
	}
	return userNames(users), nil
}

func userNames(users []*user.User) map[uuid.UUID]string {
	names := make(map[uuid.UUID]string, len(users))
	for _, resident := range users {
		names[resident.ID()] = fullName(resident)
	}
	return names
}

func leaderNameFromMap(names map[uuid.UUID]string, leaderID *uuid.UUID) string {
	if leaderID == nil {
		return "Не назначен"
	}
	name, ok := names[*leaderID]
	if !ok || strings.TrimSpace(name) == "" {
		return "Не назначен"
	}
	return name
}
