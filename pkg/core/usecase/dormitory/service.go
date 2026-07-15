package dormitory

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

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

func (s *Service) GetDormitoriesResponse(ctx context.Context) (dto.DormitoryListResponse, error) {
	dormitories, err := s.dormitoryRepo.FindAll(ctx)
	if err != nil {
		return dto.DormitoryListResponse{}, fmt.Errorf("load dormitories: %w", err)
	}

	userNames, err := s.userNames(ctx)
	if err != nil {
		return dto.DormitoryListResponse{}, fmt.Errorf("load dormitory leaders: %w", err)
	}

	items := make([]dto.DormitoryListItem, 0, len(dormitories))
	for _, dormitory := range dormitories {
		items = append(items, dto.DormitoryListItem{
			ID:      dormitory.ID(),
			Name:    dormitory.Name(),
			Address: dormitory.Address(),
			Leader:  userSummaryFromMap(userNames, dormitory.LeaderID()),
		})
	}

	return dto.DormitoryListResponse{
		Dormitories: items,
	}, nil
}

func (s *Service) GetDormitoryDetails(ctx context.Context, id int64) (*dto.DormitoryDetails, error) {
	dormitory, err := s.dormitoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load dormitory: %w", err)
	}
	if dormitory == nil {
		return nil, nil
	}

	userNames, err := s.userNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("load dormitory leaders: %w", err)
	}

	return dormitoryDetailsFromDomain(dormitory, userNames), nil
}

func (s *Service) CanManageDormitories(ctx context.Context, userID uuid.UUID) (bool, error) {
	canManage, err := s.dormitoryRepo.ExistsByLeaderID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("check dormitory management access: %w", err)
	}

	return canManage, nil
}

func (s *Service) GetUserOptionsResponse(ctx context.Context) (dto.UserOptionsResponse, error) {
	options, err := s.GetUserOptions(ctx)
	if err != nil {
		return dto.UserOptionsResponse{}, fmt.Errorf("load user options: %w", err)
	}

	items := make([]dto.UserOptionItem, 0, len(options))
	for _, option := range options {
		items = append(items, dto.UserOptionItem{
			ID:   option.ID.String(),
			Name: option.FullName,
		})
	}

	return dto.UserOptionsResponse{
		Users: items,
	}, nil
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

func (s *Service) CreateDormitoryDetails(ctx context.Context, req dto.CreateDormitoryRequest) (*dto.DormitoryDetails, error) {
	input, err := s.normalizeDormitoryInput(
		ctx,
		req.Name,
		req.City,
		req.StreetType,
		req.StreetName,
		req.HouseNumber,
		req.LeaderID,
	)
	if err != nil {
		return nil, err
	}

	dormitory := structure.RestoreDormitory(
		0,
		input.name,
		input.leaderID,
		input.city,
		input.streetType,
		input.streetName,
		input.houseNumber,
	)
	if err := s.dormitoryRepo.Save(ctx, dormitory); err != nil {
		return nil, fmt.Errorf("create dormitory: %w", err)
	}

	userNames, err := s.userNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("load dormitory leaders: %w", err)
	}

	return dormitoryDetailsFromDomain(dormitory, userNames), nil
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

func (s *Service) UpdateDormitoryDetails(ctx context.Context, id int64, req dto.UpdateDormitoryRequest) (*dto.DormitoryDetails, error) {
	current, err := s.dormitoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load dormitory: %w", err)
	}
	if current == nil {
		return nil, nil
	}

	input, err := s.normalizeDormitoryInput(
		ctx,
		req.Name,
		req.City,
		req.StreetType,
		req.StreetName,
		req.HouseNumber,
		req.LeaderID,
	)
	if err != nil {
		return nil, err
	}

	updated := structure.RestoreDormitory(
		id,
		input.name,
		input.leaderID,
		input.city,
		input.streetType,
		input.streetName,
		input.houseNumber,
	)
	if err := s.dormitoryRepo.Save(ctx, updated); err != nil {
		return nil, fmt.Errorf("update dormitory: %w", err)
	}

	userNames, err := s.userNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("load dormitory leaders: %w", err)
	}

	return dormitoryDetailsFromDomain(updated, userNames), nil
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

func (s *Service) GetGroupsResponse(ctx context.Context, dormitoryID int64) (dto.GroupListResponse, error) {
	groups, err := s.groupRepo.FindByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return dto.GroupListResponse{}, fmt.Errorf("load groups: %w", err)
	}

	residentNames, err := s.dormitoryUserNames(ctx, dormitoryID)
	if err != nil {
		return dto.GroupListResponse{}, fmt.Errorf("load group leaders: %w", err)
	}

	items := make([]dto.GroupResponseItem, 0, len(groups))
	for _, group := range groups {
		items = append(items, dto.GroupResponseItem{
			ID:     group.ID().String(),
			Name:   group.Name(),
			Leader: userSummaryFromMap(residentNames, group.LeaderID()),
		})
	}

	return dto.GroupListResponse{Groups: items}, nil
}

func (s *Service) GetGroupByID(ctx context.Context, id uuid.UUID) (*structure.Group, error) {
	return s.groupRepo.FindByID(ctx, id)
}

func (s *Service) GetGroupDetails(ctx context.Context, id uuid.UUID) (*dto.GroupDetails, error) {
	group, err := s.groupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load group: %w", err)
	}
	if group == nil {
		return nil, nil
	}

	residentNames, err := s.dormitoryUserNames(ctx, group.DormitoryID())
	if err != nil {
		return nil, fmt.Errorf("load group leaders: %w", err)
	}

	return groupDetailsFromDomain(group, residentNames), nil
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

func (s *Service) GetDormitoryUserOptionsResponse(ctx context.Context, dormitoryID int64) (dto.UserOptionsResponse, error) {
	options, err := s.GetDormitoryUserOptions(ctx, dormitoryID)
	if err != nil {
		return dto.UserOptionsResponse{}, fmt.Errorf("load dormitory user options: %w", err)
	}

	items := make([]dto.UserOptionItem, 0, len(options))
	for _, option := range options {
		items = append(items, dto.UserOptionItem{
			ID:   option.ID.String(),
			Name: option.FullName,
		})
	}

	return dto.UserOptionsResponse{
		Users: items,
	}, nil
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

func (s *Service) CreateGroupDetails(ctx context.Context, req dto.CreateGroupRequest) (*dto.GroupDetails, error) {
	input, err := s.normalizeGroupInput(ctx, req.Name, req.LeaderID, req.DormitoryID)
	if err != nil {
		return nil, err
	}

	group := structure.NewGroup(input.name, input.leaderID, "", input.dormitoryID)
	if err := s.groupRepo.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}

	residentNames, err := s.dormitoryUserNames(ctx, input.dormitoryID)
	if err != nil {
		return nil, fmt.Errorf("load group leaders: %w", err)
	}

	return groupDetailsFromDomain(group, residentNames), nil
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

func (s *Service) UpdateGroupDetails(ctx context.Context, id uuid.UUID, req dto.UpdateGroupRequest) (*dto.GroupDetails, error) {
	current, err := s.groupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load group: %w", err)
	}
	if current == nil {
		return nil, nil
	}
	if req.DormitoryID != current.DormitoryID() {
		return nil, fmt.Errorf("группа принадлежит другому общежитию")
	}

	input, err := s.normalizeGroupInput(ctx, req.Name, req.LeaderID, req.DormitoryID)
	if err != nil {
		return nil, err
	}

	updated := structure.RestoreGroup(id, input.leaderID, input.name, current.SpreadsheetID(), current.DormitoryID(), current.NextDutyTeam())
	if err := s.groupRepo.Save(ctx, updated); err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}

	residentNames, err := s.dormitoryUserNames(ctx, current.DormitoryID())
	if err != nil {
		return nil, fmt.Errorf("load group leaders: %w", err)
	}

	return groupDetailsFromDomain(updated, residentNames), nil
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

type normalizedGroupInput struct {
	name        string
	leaderID    *uuid.UUID
	dormitoryID int64
}

func (s *Service) normalizeGroupInput(
	ctx context.Context,
	name string,
	leaderID *string,
	dormitoryID int64,
) (*normalizedGroupInput, error) {
	trimmedName := strings.TrimSpace(name)
	if err := validateRequiredDormitoryField(trimmedName, "Введите название", 255, "Название не должно превышать 255 символов"); err != nil {
		return nil, err
	}

	if _, err := s.requireDormitory(ctx, dormitoryID); err != nil {
		return nil, err
	}

	parsedLeaderID, err := parseOptionalStringUUIDPointer(leaderID)
	if err != nil {
		return nil, err
	}
	if err := s.requireUserInDormitory(ctx, parsedLeaderID, dormitoryID); err != nil {
		return nil, err
	}

	return &normalizedGroupInput{
		name:        trimmedName,
		leaderID:    parsedLeaderID,
		dormitoryID: dormitoryID,
	}, nil
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

func userSummaryFromMap(names map[uuid.UUID]string, leaderID *uuid.UUID) *dto.UserSummary {
	if leaderID == nil {
		return nil
	}

	name, ok := names[*leaderID]
	if !ok || strings.TrimSpace(name) == "" {
		return nil
	}

	return &dto.UserSummary{
		ID:   leaderID.String(),
		Name: name,
	}
}

type normalizedDormitoryInput struct {
	name        string
	city        string
	streetType  string
	streetName  string
	houseNumber string
	leaderID    *uuid.UUID
}

func (s *Service) normalizeDormitoryInput(
	ctx context.Context,
	name string,
	city string,
	streetType string,
	streetName string,
	houseNumber string,
	leaderID *string,
) (*normalizedDormitoryInput, error) {
	result := &normalizedDormitoryInput{
		name:        strings.TrimSpace(name),
		city:        strings.TrimSpace(city),
		streetType:  strings.TrimSpace(streetType),
		streetName:  strings.TrimSpace(streetName),
		houseNumber: strings.TrimSpace(houseNumber),
	}

	if err := validateRequiredDormitoryField(result.name, "Введите название", 255, "Название не должно превышать 255 символов"); err != nil {
		return nil, err
	}
	if err := validateRequiredDormitoryField(result.city, "Введите город", 255, "Город не должен превышать 255 символов"); err != nil {
		return nil, err
	}
	if err := validateRequiredDormitoryField(result.streetType, "Введите тип улицы", 100, "Тип улицы не должен превышать 100 символов"); err != nil {
		return nil, err
	}
	if err := validateRequiredDormitoryField(result.streetName, "Введите название улицы", 100, "Название улицы не должно превышать 100 символов"); err != nil {
		return nil, err
	}
	if err := validateRequiredDormitoryField(result.houseNumber, "Введите номер дома", 50, "Номер дома не должен превышать 50 символов"); err != nil {
		return nil, err
	}

	parsedLeaderID, err := parseOptionalStringUUIDPointer(leaderID)
	if err != nil {
		return nil, err
	}
	if err := s.requireUser(ctx, parsedLeaderID); err != nil {
		return nil, err
	}

	result.leaderID = parsedLeaderID
	return result, nil
}

func validateRequiredDormitoryField(value string, requiredMessage string, maxLen int, maxLenMessage string) error {
	if value == "" {
		return fmt.Errorf("%s", requiredMessage)
	}
	if utf8.RuneCountInString(value) > maxLen {
		return fmt.Errorf("%s", maxLenMessage)
	}
	return nil
}

func parseOptionalStringUUIDPointer(rawID *string) (*uuid.UUID, error) {
	if rawID == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*rawID)
	if trimmed == "" {
		return nil, nil
	}

	id, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	return &id, nil
}

func dormitoryDetailsFromDomain(dormitory *structure.Dormitory, names map[uuid.UUID]string) *dto.DormitoryDetails {
	if dormitory == nil {
		return nil
	}

	return &dto.DormitoryDetails{
		ID:          dormitory.ID(),
		Name:        dormitory.Name(),
		City:        dormitory.City(),
		StreetType:  dormitory.StreetType(),
		StreetName:  dormitory.StreetName(),
		HouseNumber: dormitory.HouseNumber(),
		Leader:      userSummaryFromMap(names, dormitory.LeaderID()),
	}
}

func groupDetailsFromDomain(group *structure.Group, names map[uuid.UUID]string) *dto.GroupDetails {
	if group == nil {
		return nil
	}

	return &dto.GroupDetails{
		ID:     group.ID().String(),
		Name:   group.Name(),
		Leader: userSummaryFromMap(names, group.LeaderID()),
	}
}
