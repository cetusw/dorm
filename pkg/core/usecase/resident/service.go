package resident

import (
	"context"
	"fmt"
	"sort"
	"time"

	"dorm/pkg/core/domain/catalog"
	dutydomain "dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type currentDutyContext struct {
	resident      *user.User
	residentTeam  *structure.Team
	residentGroup *structure.Group
	selectedGroup *structure.Group
	dutyTeam      *structure.Team
	duty          *dutydomain.Duty
	groups        []*structure.Group
}

type currentDutyLookups struct {
	taskDefinitions map[uuid.UUID]*catalog.TaskDefinition
	areas           map[int]*catalog.Area
	userNames       map[uuid.UUID]string
}

type Service struct {
	userRepo   user.Repository
	teamRepo   structure.TeamRepository
	groupRepo  structure.GroupRepository
	dutyRepo   dutydomain.Repository
	taskRepo   catalog.TaskDefinitionRepository
	areaRepo   catalog.AreaRepository
	cleaningUC ports.CleaningUseCase
	now        func() time.Time
}

func NewResidentDutyService(
	userRepo user.Repository,
	teamRepo structure.TeamRepository,
	groupRepo structure.GroupRepository,
	dutyRepo dutydomain.Repository,
	taskRepo catalog.TaskDefinitionRepository,
	areaRepo catalog.AreaRepository,
	cleaningUC ports.CleaningUseCase,
) *Service {
	return &Service{
		userRepo:   userRepo,
		teamRepo:   teamRepo,
		groupRepo:  groupRepo,
		dutyRepo:   dutyRepo,
		taskRepo:   taskRepo,
		areaRepo:   areaRepo,
		cleaningUC: cleaningUC,
		now:        time.Now,
	}
}

func (s *Service) GetCurrentDuty(
	ctx context.Context,
	userID uuid.UUID,
	groupID *uuid.UUID,
) (*dto.ResidentCurrentDutyResponse, error) {
	currentDuty, err := s.loadCurrentDutyContext(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}

	if currentDuty.duty == nil || currentDuty.dutyTeam == nil {
		return buildResidentCurrentDutyResponse(currentDuty, nil, 0), nil
	}

	lookups, err := s.loadCurrentDutyLookups(ctx, currentDuty.dutyTeam.ID())
	if err != nil {
		return nil, err
	}

	canManageTasks := currentDuty.resident.TeamID() != nil &&
		*currentDuty.resident.TeamID() == currentDuty.dutyTeam.ID()

	tasks := buildResidentDutyTasks(currentDuty.duty.Tasks(), userID, lookups, canManageTasks)
	sortResidentDutyTasks(tasks)

	return buildResidentCurrentDutyResponse(
		currentDuty,
		tasks,
		len(lookups.userNames),
	), nil
}

func (s *Service) TakeTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	return s.cleaningUC.AssignTask(ctx, taskID, userID)
}

func (s *Service) ReturnTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	return s.cleaningUC.UnassignTask(ctx, taskID, userID)
}

func (s *Service) CompleteTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	return s.cleaningUC.CompleteTask(ctx, taskID, userID)
}

func (s *Service) OpenTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	return s.cleaningUC.OpenTask(ctx, taskID, userID)
}

func (s *Service) VerifyTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	return s.cleaningUC.VerifyTask(ctx, taskID, userID)
}

func (s *Service) loadCurrentDutyContext(
	ctx context.Context,
	userID uuid.UUID,
	groupID *uuid.UUID,
) (*currentDutyContext, error) {
	resident, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load resident: %w", err)
	}
	if resident == nil {
		return nil, fmt.Errorf("resident not found")
	}
	if resident.TeamID() == nil {
		return nil, fmt.Errorf("resident is not assigned to a team")
	}
	if resident.DormitoryID() == nil {
		return nil, fmt.Errorf("resident is not assigned to a dormitory")
	}

	residentTeam, err := s.teamRepo.FindByID(ctx, *resident.TeamID())
	if err != nil {
		return nil, fmt.Errorf("load resident team: %w", err)
	}
	if residentTeam == nil {
		return nil, fmt.Errorf("team not found")
	}

	residentGroup, err := s.groupRepo.FindByID(ctx, residentTeam.GroupID())
	if err != nil {
		return nil, fmt.Errorf("load resident group: %w", err)
	}
	if residentGroup == nil {
		return nil, fmt.Errorf("group not found")
	}

	groups, err := s.groupRepo.FindByDormitoryID(ctx, *resident.DormitoryID())
	if err != nil {
		return nil, fmt.Errorf("load dormitory groups: %w", err)
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("no groups found for resident dormitory")
	}

	selectedGroup, err := resolveSelectedGroup(groups, residentGroup, groupID)
	if err != nil {
		return nil, err
	}

	dutyTeam, duty, err := s.loadActiveDutyForGroup(ctx, selectedGroup.ID())
	if err != nil {
		return nil, fmt.Errorf("load active duty for group: %w", err)
	}

	return &currentDutyContext{
		resident:      resident,
		residentTeam:  residentTeam,
		residentGroup: residentGroup,
		selectedGroup: selectedGroup,
		dutyTeam:      dutyTeam,
		duty:          duty,
		groups:        groups,
	}, nil
}

func resolveSelectedGroup(
	groups []*structure.Group,
	residentGroup *structure.Group,
	groupID *uuid.UUID,
) (*structure.Group, error) {
	if groupID == nil {
		return residentGroup, nil
	}

	for _, group := range groups {
		if group.ID() == *groupID {
			return group, nil
		}
	}

	return nil, fmt.Errorf("group does not belong to resident dormitory")
}

func (s *Service) loadActiveDutyForGroup(
	ctx context.Context,
	groupID uuid.UUID,
) (*structure.Team, *dutydomain.Duty, error) {
	teams, err := s.teamRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, nil, fmt.Errorf("load group teams: %w", err)
	}

	for _, team := range teams {
		duty, err := s.dutyRepo.FindActiveByTeamID(ctx, team.ID(), s.now())
		if err != nil {
			return nil, nil, fmt.Errorf("load active duty by team: %w", err)
		}
		if duty != nil {
			return team, duty, nil
		}
	}

	return nil, nil, nil
}

func (s *Service) loadCurrentDutyLookups(
	ctx context.Context,
	teamID uuid.UUID,
) (*currentDutyLookups, error) {
	taskDefinitions, err := s.taskDefinitionsByID(ctx)
	if err != nil {
		return nil, err
	}

	areas, err := s.areasByID(ctx)
	if err != nil {
		return nil, err
	}

	userNames, err := s.userNamesByID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	return &currentDutyLookups{
		taskDefinitions: taskDefinitions,
		areas:           areas,
		userNames:       userNames,
	}, nil
}

func buildResidentDutyTasks(
	dutyTasks []*dutydomain.DutyTask,
	userID uuid.UUID,
	lookups *currentDutyLookups,
	canManageTasks bool,
) []dto.ResidentDutyTask {
	tasks := make([]dto.ResidentDutyTask, 0, len(dutyTasks))

	for _, dutyTask := range dutyTasks {
		taskDefinition, area, ok := resolveResidentDutyTaskDetails(dutyTask, lookups)
		if !ok {
			continue
		}

		tasks = append(tasks, buildResidentDutyTask(
			dutyTask,
			taskDefinition,
			area,
			userID,
			lookups.userNames,
			canManageTasks,
		))
	}

	return tasks
}

func resolveResidentDutyTaskDetails(
	dutyTask *dutydomain.DutyTask,
	lookups *currentDutyLookups,
) (*catalog.TaskDefinition, *catalog.Area, bool) {
	taskDefinition, ok := lookups.taskDefinitions[dutyTask.TaskDefID()]
	if !ok {
		return nil, nil, false
	}

	area, ok := lookups.areas[taskDefinition.AreaID()]
	if !ok {
		return nil, nil, false
	}

	return taskDefinition, area, true
}

func buildResidentDutyTask(
	dutyTask *dutydomain.DutyTask,
	taskDefinition *catalog.TaskDefinition,
	area *catalog.Area,
	userID uuid.UUID,
	userNames map[uuid.UUID]string,
	canManageTasks bool,
) dto.ResidentDutyTask {
	status := resolveTaskStatus(dutyTask)
	isMine := dutyTask.AssigneeID() != nil && *dutyTask.AssigneeID() == userID
	canTake, canReturn, canComplete, canOpen, canVerify, canReviewOpen := buildTaskPermissions(
		status,
		isMine,
		canManageTasks,
	)
	assigneeID, assigneeName := resolveAssignee(dutyTask, userNames)

	return dto.ResidentDutyTask{
		ID:            dutyTask.ID().String(),
		AreaName:      area.Name(),
		AreaFloor:     area.Floor(),
		Title:         taskDefinition.Title(),
		Cost:          taskDefinition.Cost(),
		Status:        status,
		AssigneeID:    assigneeID,
		AssigneeName:  assigneeName,
		IsMine:        isMine,
		CanTake:       canTake,
		CanReturn:     canReturn,
		CanComplete:   canComplete,
		CanOpen:       canOpen,
		CanVerify:     canVerify,
		CanReviewOpen: canReviewOpen,
	}
}

func resolveAssignee(
	dutyTask *dutydomain.DutyTask,
	userNames map[uuid.UUID]string,
) (*string, *string) {
	if dutyTask.AssigneeID() == nil {
		return nil, nil
	}

	rawAssigneeID := dutyTask.AssigneeID().String()
	assigneeName, ok := userNames[*dutyTask.AssigneeID()]
	if !ok {
		return &rawAssigneeID, nil
	}

	return &rawAssigneeID, &assigneeName
}

func buildResidentCurrentDutyResponse(
	currentDuty *currentDutyContext,
	tasks []dto.ResidentDutyTask,
	residentCount int,
) *dto.ResidentCurrentDutyResponse {
	response := &dto.ResidentCurrentDutyResponse{
		SelectedGroupID: currentDuty.selectedGroup.ID().String(),
		Groups:          buildResidentDutyGroupOptions(currentDuty.groups),
		Group:           currentDuty.selectedGroup.Name(),
		CanManageTasks: currentDuty.duty != nil &&
			currentDuty.resident.TeamID() != nil &&
			currentDuty.dutyTeam != nil &&
			*currentDuty.resident.TeamID() == currentDuty.dutyTeam.ID(),
		Tasks: tasks,
	}

	if currentDuty.duty == nil || currentDuty.dutyTeam == nil {
		return response
	}

	return &dto.ResidentCurrentDutyResponse{
		SelectedGroupID:     response.SelectedGroupID,
		HasActiveDuty:       true,
		CanManageTasks:      response.CanManageTasks,
		Groups:              response.Groups,
		DutyID:              currentDuty.duty.ID().String(),
		Group:               currentDuty.selectedGroup.Name(),
		Team:                currentDuty.dutyTeam.Name(),
		StartDate:           currentDuty.duty.Start().Format("2006-01-02"),
		EndDate:             currentDuty.duty.End().Format("2006-01-02"),
		CostPerResidentGoal: calculateCostPerResidentGoal(tasks, residentCount),
		MyTakenCostSum:      countMyTakenCost(tasks),
		Tasks:               response.Tasks,
	}
}

func buildResidentDutyGroupOptions(groups []*structure.Group) []dto.ResidentDutyGroupOption {
	options := make([]dto.ResidentDutyGroupOption, 0, len(groups))
	for _, group := range groups {
		options = append(options, dto.ResidentDutyGroupOption{
			ID:   group.ID().String(),
			Name: group.Name(),
		})
	}

	return options
}

func (s *Service) taskDefinitionsByID(
	ctx context.Context,
) (map[uuid.UUID]*catalog.TaskDefinition, error) {
	taskDefinitions, err := s.taskRepo.GetAllTaskDefinitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("load task definitions: %w", err)
	}

	result := make(map[uuid.UUID]*catalog.TaskDefinition, len(taskDefinitions))
	for _, taskDefinition := range taskDefinitions {
		result[taskDefinition.ID()] = taskDefinition
	}

	return result, nil
}

func (s *Service) areasByID(ctx context.Context) (map[int]*catalog.Area, error) {
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, fmt.Errorf("load areas: %w", err)
	}

	result := make(map[int]*catalog.Area, len(areas))
	for _, area := range areas {
		result[area.ID()] = area
	}

	return result, nil
}

func (s *Service) userNamesByID(
	ctx context.Context,
	teamID uuid.UUID,
) (map[uuid.UUID]string, error) {
	users, err := s.userRepo.FindByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("load team users: %w", err)
	}

	result := make(map[uuid.UUID]string, len(users))
	for _, resident := range users {
		result[resident.ID()] = formatUserName(resident)
	}

	return result, nil
}

func resolveTaskStatus(task *dutydomain.DutyTask) string {
	if task.VerificationDate() != nil {
		return dto.ResidentDutyTaskStatusVerified
	}
	if task.CompletionDate() != nil {
		return dto.ResidentDutyTaskStatusCompleted
	}
	if task.AssigneeID() != nil {
		return dto.ResidentDutyTaskStatusAssigned
	}
	return dto.ResidentDutyTaskStatusFree
}

func buildTaskPermissions(
	status string,
	isMine bool,
	canManageTasks bool,
) (canTake bool, canReturn bool, canComplete bool, canOpen bool, canVerify bool, canReviewOpen bool) {
	if !canManageTasks {
		return false, false, false, false, false, false
	}

	switch status {
	case dto.ResidentDutyTaskStatusFree:
		return true, false, false, false, false, false
	case dto.ResidentDutyTaskStatusAssigned:
		if isMine {
			return false, true, true, false, false, false
		}
		return false, false, false, false, false, false
	case dto.ResidentDutyTaskStatusCompleted:
		if isMine {
			return false, false, false, true, true, true
		}
		return false, false, false, false, true, true
	default:
		return false, false, false, false, false, false
	}
}

func countMyTakenCost(tasks []dto.ResidentDutyTask) int {
	total := 0
	for _, task := range tasks {
		if task.IsMine {
			total += task.Cost
		}
	}

	return total
}

func calculateCostPerResidentGoal(tasks []dto.ResidentDutyTask, residentCount int) int {
	if len(tasks) == 0 || residentCount == 0 {
		return 0
	}

	totalCost := 0
	for _, task := range tasks {
		totalCost += task.Cost
	}

	return (totalCost + residentCount - 1) / residentCount
}

func sortResidentDutyTasks(tasks []dto.ResidentDutyTask) {
	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].AreaFloor != tasks[j].AreaFloor {
			return tasks[i].AreaFloor > tasks[j].AreaFloor
		}
		if tasks[i].AreaName != tasks[j].AreaName {
			return tasks[i].AreaName < tasks[j].AreaName
		}
		if tasks[i].Cost != tasks[j].Cost {
			return tasks[i].Cost > tasks[j].Cost
		}
		return tasks[i].Title < tasks[j].Title
	})
}

func formatUserName(resident *user.User) string {
	return fmt.Sprintf("%s %s", resident.FirstName(), resident.LastName())
}
