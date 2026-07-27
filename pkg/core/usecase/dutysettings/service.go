package dutysettings

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

var ErrAccessDenied = errors.New("duty settings access denied")

type Service struct {
	groupRepo structure.GroupRepository
	teamRepo  structure.TeamRepository
	areaRepo  catalog.AreaRepository
	taskRepo  catalog.TaskDefinitionRepository
	dutyRepo  duty.DutyRepository
	userRepo  user.Repository
}

func NewDutySettingsService(
	groupRepo structure.GroupRepository,
	teamRepo structure.TeamRepository,
	areaRepo catalog.AreaRepository,
	taskRepo catalog.TaskDefinitionRepository,
	dutyRepo duty.DutyRepository,
	userRepo user.Repository,
) *Service {
	return &Service{
		groupRepo: groupRepo,
		teamRepo:  teamRepo,
		areaRepo:  areaRepo,
		taskRepo:  taskRepo,
		dutyRepo:  dutyRepo,
		userRepo:  userRepo,
	}
}

func (s *Service) GetDutySettings(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID) (*dto.DutySettingsResponse, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return nil, err
	}

	areas, err := s.listGroupAreas(ctx, group.ID())
	if err != nil {
		return nil, err
	}

	tasks, err := s.taskRepo.FindByGroupID(ctx, group.ID())
	if err != nil {
		return nil, fmt.Errorf("load group tasks: %w", err)
	}

	taskIDs := make([]uuid.UUID, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID())
	}

	lastCompletionDates, err := s.taskRepo.FindLastCompletionDates(ctx, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("load task completion dates: %w", err)
	}

	tasksByArea := make(map[int][]dto.DutySettingsTask, len(areas))
	for _, task := range tasks {
		tasksByArea[task.AreaID()] = append(tasksByArea[task.AreaID()], dto.DutySettingsTask{
			ID:              task.ID().String(),
			Title:           task.Title(),
			Cost:            task.Cost(),
			Frequency:       task.Frequency(),
			LastCompletedAt: lastCompletionDates[task.ID()],
		})
	}

	responseAreas := make([]dto.DutySettingsArea, 0, len(areas))
	for _, area := range areas {
		areaTasks := tasksByArea[area.ID()]
		sort.Slice(areaTasks, func(i, j int) bool {
			if areaTasks[i].Cost != areaTasks[j].Cost {
				return areaTasks[i].Cost > areaTasks[j].Cost
			}
			return areaTasks[i].Title < areaTasks[j].Title
		})

		responseAreas = append(responseAreas, dto.DutySettingsArea{
			ID:    area.ID(),
			Name:  area.Name(),
			Floor: areaFloor(area),
			Tasks: areaTasks,
		})
	}

	sort.Slice(responseAreas, func(i, j int) bool {
		leftFloor := responseAreas[i].Floor
		rightFloor := responseAreas[j].Floor
		switch {
		case leftFloor == nil && rightFloor != nil:
			return false
		case leftFloor != nil && rightFloor == nil:
			return true
		case leftFloor != nil && rightFloor != nil && *leftFloor != *rightFloor:
			return *leftFloor > *rightFloor
		default:
			return responseAreas[i].Name < responseAreas[j].Name
		}
	})

	responseTeams, activeDutyTeamID, err := s.buildDutySettingsTeams(ctx, group)
	if err != nil {
		return nil, err
	}

	return &dto.DutySettingsResponse{
		Group: dto.DutySettingsGroup{
			ID:          group.ID().String(),
			Name:        group.Name(),
			DormitoryID: group.DormitoryID(),
		},
		Areas:            responseAreas,
		Teams:            responseTeams,
		ActiveDutyTeamID: activeDutyTeamID,
	}, nil
}

func (s *Service) GetTeamDetails(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamID uuid.UUID) (*dto.TeamDetails, error) {
	group, team, err := s.requireManagedGroupTeam(ctx, currentUserID, groupID, teamID)
	if err != nil {
		return nil, err
	}

	residents, err := s.userRepo.FindByDormitoryID(ctx, group.DormitoryID())
	if err != nil {
		return nil, fmt.Errorf("load residents: %w", err)
	}
	residentNames := userNames(residents)

	members, err := s.userRepo.FindByTeamID(ctx, team.ID())
	if err != nil {
		return nil, fmt.Errorf("load team members: %w", err)
	}

	memberIDs := make([]string, 0, len(members))
	for _, member := range members {
		memberIDs = append(memberIDs, member.ID().String())
	}

	return &dto.TeamDetails{
		ID:        team.ID().String(),
		Name:      team.Name(),
		GroupID:   group.ID().String(),
		GroupName: group.Name(),
		Leader:    userSummaryFromMap(residentNames, team.LeaderID()),
		MemberIDs: memberIDs,
	}, nil
}

func (s *Service) GetTeamMemberOptions(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamID *uuid.UUID) (dto.TeamMemberOptionsResponse, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return dto.TeamMemberOptionsResponse{}, err
	}

	currentTeamID := uuid.Nil
	if teamID != nil {
		team, err := s.teamRepo.FindByID(ctx, *teamID)
		if err != nil {
			return dto.TeamMemberOptionsResponse{}, fmt.Errorf("load team: %w", err)
		}
		if team == nil || team.GroupID() != group.ID() {
			return dto.TeamMemberOptionsResponse{}, fmt.Errorf("команда не найдена")
		}
		currentTeamID = *teamID
	}

	items, err := s.getTeamMemberItems(ctx, group.DormitoryID(), currentTeamID)
	if err != nil {
		return dto.TeamMemberOptionsResponse{}, err
	}

	responseItems := make([]dto.TeamMemberOptionItem, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, dto.TeamMemberOptionItem{
			ID:              item.UserID.String(),
			Name:            item.FullName,
			CurrentTeamID:   optionalUUIDString(item.CurrentTeamID),
			CurrentTeamName: item.CurrentTeamName,
			IsInCurrentTeam: item.IsInCurrentTeam,
		})
	}

	return dto.TeamMemberOptionsResponse{Members: responseItems}, nil
}

func (s *Service) CreateTeam(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, req dto.CreateResidentTeamRequest) (*dto.TeamDetails, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return nil, err
	}
	if parsedGroupID, err := uuid.Parse(req.GroupID); err != nil || parsedGroupID != groupID {
		return nil, fmt.Errorf("invalid group id")
	}

	memberIDs, err := parseMemberIDs(req.MemberIDs)
	if err != nil {
		return nil, err
	}
	leaderID, err := parseOptionalStringUUIDPointer(req.LeaderID)
	if err != nil {
		return nil, err
	}
	memberIDs = appendUniqueOptionalUUID(memberIDs, leaderID)
	if err := s.requireUsersInDormitory(ctx, group.DormitoryID(), memberIDs); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if err := validateResidentTeamName(name); err != nil {
		return nil, err
	}

	position, err := s.nextTeamOrder(ctx, groupID)
	if err != nil {
		return nil, err
	}

	team := structure.NewTeam(name, groupID, "", position)
	if leaderID != nil {
		team = structure.RestoreTeam(team.ID(), team.Name(), team.GroupID(), leaderID, team.Color(), team.RotationPosition())
	}
	if err := s.teamRepo.Save(ctx, team); err != nil {
		return nil, err
	}
	if err := s.moveMembers(ctx, team.ID(), memberIDs); err != nil {
		return nil, err
	}

	return s.GetTeamDetails(ctx, currentUserID, groupID, team.ID())
}

func (s *Service) UpdateTeam(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamID uuid.UUID, req dto.UpdateResidentTeamRequest) (*dto.TeamDetails, error) {
	group, team, err := s.requireManagedGroupTeam(ctx, currentUserID, groupID, teamID)
	if err != nil {
		return nil, err
	}
	if parsedGroupID, err := uuid.Parse(req.GroupID); err != nil || parsedGroupID != groupID {
		return nil, fmt.Errorf("invalid group id")
	}

	memberIDs, err := parseMemberIDs(req.MemberIDs)
	if err != nil {
		return nil, err
	}
	leaderID, err := parseOptionalStringUUIDPointer(req.LeaderID)
	if err != nil {
		return nil, err
	}
	memberIDs = appendUniqueOptionalUUID(memberIDs, leaderID)
	if err := s.requireUsersInDormitory(ctx, group.DormitoryID(), memberIDs); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if err := validateResidentTeamName(name); err != nil {
		return nil, err
	}

	updated := structure.RestoreTeam(teamID, name, groupID, leaderID, team.Color(), team.RotationPosition())
	if err := s.teamRepo.Save(ctx, updated); err != nil {
		return nil, err
	}
	if err := s.syncMembers(ctx, teamID, memberIDs); err != nil {
		return nil, err
	}

	return s.GetTeamDetails(ctx, currentUserID, groupID, teamID)
}

func (s *Service) DeleteTeam(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamID uuid.UUID) error {
	_, _, err := s.requireManagedGroupTeam(ctx, currentUserID, groupID, teamID)
	if err != nil {
		return err
	}
	return s.teamRepo.Delete(ctx, teamID)
}

func (s *Service) ReorderTeams(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamIDs []uuid.UUID) error {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return err
	}

	teams, err := s.teamRepo.FindByGroupID(ctx, group.ID())
	if err != nil {
		return fmt.Errorf("load teams: %w", err)
	}
	if len(teams) != len(teamIDs) {
		return fmt.Errorf("invalid team order")
	}

	expected := make(map[uuid.UUID]struct{}, len(teams))
	for _, team := range teams {
		expected[team.ID()] = struct{}{}
	}

	seen := make(map[uuid.UUID]struct{}, len(teamIDs))
	for _, teamID := range teamIDs {
		if _, ok := expected[teamID]; !ok {
			return fmt.Errorf("invalid team order")
		}
		if _, ok := seen[teamID]; ok {
			return fmt.Errorf("invalid team order")
		}
		seen[teamID] = struct{}{}
	}

	return s.teamRepo.UpdateRotationPositions(ctx, group.ID(), teamIDs)
}

func (s *Service) CreateArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, req dto.CreateAreaRequest) (*dto.AreaDetails, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeAreaInput(req.Name, req.Floor)
	if err != nil {
		return nil, err
	}

	area := catalog.NewArea(normalized.name, normalized.floorValue(), &groupID)
	if err := s.areaRepo.Save(ctx, area); err != nil {
		return nil, fmt.Errorf("create area: %w", err)
	}

	return buildAreaDetails(area, group), nil
}

func (s *Service) UpdateArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int, req dto.UpdateAreaRequest) (*dto.AreaDetails, error) {
	group, area, err := s.requireManagedGroupArea(ctx, currentUserID, groupID, areaID)
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeAreaInput(req.Name, req.Floor)
	if err != nil {
		return nil, err
	}

	updated := catalog.RestoreArea(area.ID(), normalized.name, normalized.floorValue(), area.GroupID())
	if err := s.areaRepo.Save(ctx, updated); err != nil {
		return nil, fmt.Errorf("update area: %w", err)
	}

	return buildAreaDetails(updated, group), nil
}

func (s *Service) DeleteArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int) error {
	_, _, err := s.requireManagedGroupArea(ctx, currentUserID, groupID, areaID)
	if err != nil {
		return err
	}

	if err := s.areaRepo.Delete(ctx, areaID); err != nil {
		return fmt.Errorf("delete area: %w", err)
	}

	return nil
}

func (s *Service) CreateTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int, req dto.CreateTaskRequest) (*dto.TaskDetails, error) {
	group, area, err := s.requireManagedGroupArea(ctx, currentUserID, groupID, areaID)
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeTaskInput(req.Title, req.Cost, req.Frequency)
	if err != nil {
		return nil, err
	}

	task, err := catalog.NewTaskDefinition(area.ID(), normalized.title, normalized.cost, normalized.frequency)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return buildTaskDetails(task, group, area), nil
}

func (s *Service) UpdateTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, taskID uuid.UUID, req dto.UpdateTaskRequest) (*dto.TaskDetails, error) {
	group, task, _, err := s.requireManagedGroupTask(ctx, currentUserID, groupID, taskID)
	if err != nil {
		return nil, err
	}

	targetArea, err := s.areaRepo.FindByID(ctx, req.AreaID)
	if err != nil {
		return nil, fmt.Errorf("load target area: %w", err)
	}
	if targetArea == nil {
		return nil, fmt.Errorf("территория не найдена")
	}
	if err := requireAreaInGroup(targetArea, group.ID()); err != nil {
		return nil, err
	}

	normalized, err := normalizeTaskInput(req.Title, req.Cost, req.Frequency)
	if err != nil {
		return nil, err
	}

	updated := catalog.RestoreTaskDefinition(task.ID(), targetArea.ID(), normalized.title, normalized.cost, normalized.frequency)
	if err := s.taskRepo.Save(ctx, updated); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	return buildTaskDetails(updated, group, targetArea), nil
}

func (s *Service) DeleteTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, taskID uuid.UUID) error {
	_, _, _, err := s.requireManagedGroupTask(ctx, currentUserID, groupID, taskID)
	if err != nil {
		return err
	}

	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (s *Service) buildDutySettingsTeams(ctx context.Context, group *structure.Group) ([]dto.DutySettingsTeam, *string, error) {
	teams, err := s.teamRepo.FindByGroupID(ctx, group.ID())
	if err != nil {
		return nil, nil, fmt.Errorf("load teams: %w", err)
	}

	residents, err := s.userRepo.FindByDormitoryID(ctx, group.DormitoryID())
	if err != nil {
		return nil, nil, fmt.Errorf("load residents: %w", err)
	}

	memberCounts := make(map[uuid.UUID]int, len(teams))
	residentNames := make(map[uuid.UUID]string, len(residents))
	for _, resident := range residents {
		residentNames[resident.ID()] = strings.TrimSpace(fmt.Sprintf("%s %s", resident.LastName(), resident.FirstName()))
		if resident.TeamID() != nil {
			memberCounts[*resident.TeamID()]++
		}
	}

	sort.Slice(teams, func(i, j int) bool {
		return teams[i].RotationPosition() < teams[j].RotationPosition()
	})

	responseTeams := make([]dto.DutySettingsTeam, 0, len(teams))
	for _, team := range teams {
		responseTeams = append(responseTeams, dto.DutySettingsTeam{
			ID:               team.ID().String(),
			Name:             team.Name(),
			RotationPosition: team.RotationPosition(),
			Leader:           userSummaryFromMap(residentNames, team.LeaderID()),
			MembersCount:     memberCounts[team.ID()],
		})
	}

	latestDuty, err := s.dutyRepo.FindLatestByGroupID(ctx, group.ID())
	if err != nil {
		return nil, nil, fmt.Errorf("load latest duty: %w", err)
	}
	if latestDuty == nil {
		return responseTeams, nil, nil
	}

	activeDutyTeamID := latestDuty.TeamID().String()
	return responseTeams, &activeDutyTeamID, nil
}

func (s *Service) requireManagedGroup(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID) (*structure.Group, error) {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("load group: %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("группа не найдена")
	}
	if group.LeaderID() == nil || *group.LeaderID() != currentUserID {
		return nil, ErrAccessDenied
	}

	return group, nil
}

func (s *Service) requireManagedGroupArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int) (*structure.Group, *catalog.Area, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return nil, nil, err
	}

	area, err := s.areaRepo.FindByID(ctx, areaID)
	if err != nil {
		return nil, nil, fmt.Errorf("load area: %w", err)
	}
	if area == nil {
		return nil, nil, fmt.Errorf("территория не найдена")
	}
	if err := requireAreaInGroup(area, group.ID()); err != nil {
		return nil, nil, err
	}

	return group, area, nil
}

func (s *Service) requireManagedGroupTeam(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, teamID uuid.UUID) (*structure.Group, *structure.Team, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return nil, nil, err
	}

	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, nil, fmt.Errorf("load team: %w", err)
	}
	if team == nil || team.GroupID() != group.ID() {
		return nil, nil, fmt.Errorf("команда не найдена")
	}

	return group, team, nil
}

func parseMemberIDs(rawIDs []string) ([]uuid.UUID, error) {
	memberIDs := make([]uuid.UUID, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		userID, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("invalid member id")
		}
		memberIDs = append(memberIDs, userID)
	}
	return memberIDs, nil
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

func appendUniqueOptionalUUID(ids []uuid.UUID, optionalID *uuid.UUID) []uuid.UUID {
	if optionalID == nil {
		return ids
	}

	for _, id := range ids {
		if id == *optionalID {
			return ids
		}
	}

	return append(ids, *optionalID)
}

func optionalUUIDString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}

	stringValue := value.String()
	return &stringValue
}

func validateResidentTeamName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("Введите название")
	}
	if utf8.RuneCountInString(strings.TrimSpace(name)) > 255 {
		return fmt.Errorf("Название не должно превышать 255 символов")
	}
	return nil
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

func userNames(users []*user.User) map[uuid.UUID]string {
	names := make(map[uuid.UUID]string, len(users))
	for _, resident := range users {
		names[resident.ID()] = strings.TrimSpace(fmt.Sprintf("%s %s", resident.LastName(), resident.FirstName()))
	}
	return names
}

func (s *Service) nextTeamOrder(ctx context.Context, groupID uuid.UUID) (int, error) {
	teams, err := s.teamRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return 0, err
	}

	maxOrder := 0
	for _, team := range teams {
		if team.RotationPosition() > maxOrder {
			maxOrder = team.RotationPosition()
		}
	}

	return maxOrder + 1, nil
}

func (s *Service) moveMembers(ctx context.Context, teamID uuid.UUID, memberIDs []uuid.UUID) error {
	for _, userID := range memberIDs {
		if err := s.userRepo.MoveUserToTeam(ctx, userID, &teamID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) syncMembers(ctx context.Context, teamID uuid.UUID, memberIDs []uuid.UUID) error {
	selected := make(map[uuid.UUID]struct{}, len(memberIDs))
	for _, userID := range memberIDs {
		selected[userID] = struct{}{}
	}

	currentMembers, err := s.userRepo.FindByTeamID(ctx, teamID)
	if err != nil {
		return err
	}
	for _, resident := range currentMembers {
		if _, ok := selected[resident.ID()]; ok {
			continue
		}
		if err := s.userRepo.MoveUserToTeam(ctx, resident.ID(), nil); err != nil {
			return err
		}
	}

	for _, userID := range memberIDs {
		if err := s.userRepo.MoveUserToTeam(ctx, userID, &teamID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) requireUsersInDormitory(ctx context.Context, dormitoryID int64, userIDs []uuid.UUID) error {
	for _, userID := range userIDs {
		resident, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		if resident == nil {
			return fmt.Errorf("user not found")
		}
		if resident.DormitoryID() == nil || *resident.DormitoryID() != dormitoryID {
			return fmt.Errorf("user does not belong to dormitory")
		}
	}
	return nil
}

func (s *Service) getTeamMemberItems(ctx context.Context, dormitoryID int64, currentTeamID uuid.UUID) ([]dto.TeamMemberItem, error) {
	users, err := s.userRepo.FindByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return nil, err
	}

	allGroups, err := s.groupRepo.FindByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return nil, err
	}

	teamNames := make(map[uuid.UUID]string)
	for _, group := range allGroups {
		groupTeams, err := s.teamRepo.FindByGroupID(ctx, group.ID())
		if err != nil {
			return nil, err
		}
		for _, team := range groupTeams {
			teamNames[team.ID()] = team.Name()
		}
	}

	items := make([]dto.TeamMemberItem, 0, len(users))
	for _, resident := range users {
		fullName := strings.TrimSpace(fmt.Sprintf("%s %s", resident.LastName(), resident.FirstName()))
		var currentTeamName string
		if resident.TeamID() != nil {
			currentTeamName = teamNames[*resident.TeamID()]
		}

		items = append(items, dto.TeamMemberItem{
			UserID:          resident.ID(),
			FullName:        fullName,
			CurrentTeamID:   resident.TeamID(),
			CurrentTeamName: currentTeamName,
			IsInCurrentTeam: currentTeamID != uuid.Nil && resident.TeamID() != nil && *resident.TeamID() == currentTeamID,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsInCurrentTeam != items[j].IsInCurrentTeam {
			return items[i].IsInCurrentTeam
		}
		return strings.ToLower(items[i].FullName) < strings.ToLower(items[j].FullName)
	})

	return items, nil
}

func (s *Service) requireManagedGroupTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, taskID uuid.UUID) (*structure.Group, *catalog.TaskDefinition, *catalog.Area, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return nil, nil, nil, err
	}

	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load task: %w", err)
	}
	if task == nil {
		return nil, nil, nil, fmt.Errorf("задача не найдена")
	}

	area, err := s.areaRepo.FindByID(ctx, task.AreaID())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load task area: %w", err)
	}
	if area == nil {
		return nil, nil, nil, fmt.Errorf("территория не найдена")
	}
	if err := requireAreaInGroup(area, group.ID()); err != nil {
		return nil, nil, nil, err
	}

	return group, task, area, nil
}

func (s *Service) listGroupAreas(ctx context.Context, groupID uuid.UUID) ([]*catalog.Area, error) {
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, fmt.Errorf("load areas: %w", err)
	}

	result := make([]*catalog.Area, 0, len(areas))
	for _, area := range areas {
		if area.GroupID() != nil && *area.GroupID() == groupID {
			result = append(result, area)
		}
	}

	return result, nil
}

type normalizedAreaInput struct {
	name  string
	floor *int
}

func (i normalizedAreaInput) floorValue() int {
	if i.floor == nil {
		return 0
	}
	return *i.floor
}

func normalizeAreaInput(name string, floor *int) (*normalizedAreaInput, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, fmt.Errorf("Введите название")
	}
	if utf8.RuneCountInString(trimmedName) > 255 {
		return nil, fmt.Errorf("Название не должно превышать 255 символов")
	}

	return &normalizedAreaInput{name: trimmedName, floor: floor}, nil
}

type normalizedTaskInput struct {
	title     string
	cost      int
	frequency int
}

func normalizeTaskInput(title string, cost int, frequency int) (*normalizedTaskInput, error) {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return nil, fmt.Errorf("Введите название")
	}
	if utf8.RuneCountInString(trimmedTitle) > 255 {
		return nil, fmt.Errorf("Название не должно превышать 255 символов")
	}
	if cost <= 0 {
		return nil, fmt.Errorf("Стоимость должна быть больше нуля")
	}
	if frequency <= 0 {
		return nil, fmt.Errorf("Частота должна быть больше нуля")
	}

	return &normalizedTaskInput{
		title:     trimmedTitle,
		cost:      cost,
		frequency: frequency,
	}, nil
}

func requireAreaInGroup(area *catalog.Area, groupID uuid.UUID) error {
	if area.GroupID() == nil || *area.GroupID() != groupID {
		return fmt.Errorf("территория принадлежит другой группе")
	}

	return nil
}

func areaFloor(area *catalog.Area) *int {
	if area == nil || area.Floor() == 0 {
		return nil
	}

	floor := area.Floor()
	return &floor
}

func buildAreaDetails(area *catalog.Area, group *structure.Group) *dto.AreaDetails {
	return &dto.AreaDetails{
		ID:    area.ID(),
		Name:  area.Name(),
		Floor: areaFloor(area),
		Group: &dto.GroupOption{
			ID:   group.ID().String(),
			Name: group.Name(),
		},
	}
}

func buildTaskDetails(task *catalog.TaskDefinition, group *structure.Group, area *catalog.Area) *dto.TaskDetails {
	return &dto.TaskDetails{
		ID:        task.ID().String(),
		Title:     task.Title(),
		Cost:      task.Cost(),
		Frequency: task.Frequency(),
		Area: dto.AreaSummary{
			ID:   area.ID(),
			Name: area.Name(),
		},
	}
}
