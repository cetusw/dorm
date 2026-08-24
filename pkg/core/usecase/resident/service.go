package resident

import (
	"context"
	"errors"
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

var (
	ErrResidentDutyGroupNotFound = errors.New("resident duty group not found")
	ErrResidentDutyNotFound      = errors.New("resident duty not found")
	ErrResidentDutyAccessDenied  = errors.New("resident duty access denied")
)

const (
	dutyPeriodStatusPast   = "past"
	dutyPeriodStatusActive = "active"
	dutyPeriodStatusFuture = "future"
)

type residentAccessContext struct {
	resident          *user.User
	residentTeam      *structure.Team
	residentGroup     *structure.Group
	dormitory         *structure.Dormitory
	myDormitoryID     int64
	myResidentTeam    *structure.Team
	myResidentGroup   *structure.Group
	availableGroups   []*structure.Group
	canSelectAnyGroup bool
}

type currentDutyContext struct {
	resident          *user.User
	residentTeam      *structure.Team
	residentGroup     *structure.Group
	dormitory         *structure.Dormitory
	myDormitoryID     int64
	myResidentTeam    *structure.Team
	myResidentGroup   *structure.Group
	canSelectAnyGroup bool
	selectedGroup     *structure.Group
	activeDutyTeam    *structure.Team
	activeDuty        *dutydomain.Duty
	dutyTeam          *structure.Team
	duty              *dutydomain.Duty
	groups            []*structure.Group
}

type residentDutyView struct {
	showGroupSelect bool
	visibleTabs     []string
	readOnly        bool
	canViewTasks    bool
	canManageTasks  bool
	canVerifyTasks  bool
	noticeMessage   string
	noticeTone      string
}

type currentDutyLookups struct {
	taskDefinitions map[uuid.UUID]*catalog.TaskDefinition
	areas           map[int]*catalog.Area
	userNames       map[uuid.UUID]string
	teamMembers     []dto.ResidentDutyTeamMember
}

type Service struct {
	userRepo        user.Repository
	teamRepo        structure.TeamRepository
	groupRepo       structure.GroupRepository
	dormRepo        structure.DormitoryRepository
	dutyRepo        dutydomain.DutyRepository
	participantRepo dutydomain.ParticipantRepository
	taskRepo        catalog.TaskDefinitionRepository
	areaRepo        catalog.AreaRepository
	cleaningUC      ports.CleaningUseCase
	now             func() time.Time
	location        *time.Location
}

func (s *Service) SetParticipantRepository(repo dutydomain.ParticipantRepository) {
	s.participantRepo = repo
}

func NewResidentDutyService(
	userRepo user.Repository,
	teamRepo structure.TeamRepository,
	groupRepo structure.GroupRepository,
	dormRepo structure.DormitoryRepository,
	dutyRepo dutydomain.DutyRepository,
	taskRepo catalog.TaskDefinitionRepository,
	areaRepo catalog.AreaRepository,
	cleaningUC ports.CleaningUseCase,
) *Service {
	return &Service{
		userRepo:   userRepo,
		teamRepo:   teamRepo,
		groupRepo:  groupRepo,
		dormRepo:   dormRepo,
		dutyRepo:   dutyRepo,
		taskRepo:   taskRepo,
		areaRepo:   areaRepo,
		cleaningUC: cleaningUC,
		now:        time.Now,
		location:   time.Local,
	}
}

func (s *Service) SetLocation(location *time.Location) {
	if location != nil {
		s.location = location
	}
}

func (s *Service) GetCurrentDuty(
	ctx context.Context,
	userID uuid.UUID,
	groupID *uuid.UUID,
	dormitoryID *int64,
) (*dto.ResidentCurrentDutyResponse, error) {
	currentDuty, err := s.loadCurrentDutyContext(ctx, userID, groupID, dormitoryID)
	if err != nil {
		return nil, err
	}

	if currentDuty.duty == nil || currentDuty.dutyTeam == nil {
		view, err := s.resolveResidentDutyView(ctx, currentDuty)
		if err != nil {
			return nil, err
		}
		return buildResidentCurrentDutyResponse(currentDuty, view, nil, nil, 0, s.currentTime()), nil
	}

	view, err := s.resolveResidentDutyView(ctx, currentDuty)
	if err != nil {
		return nil, err
	}
	if !view.canViewTasks {
		return buildResidentCurrentDutyResponse(currentDuty, view, nil, nil, 0, s.currentTime()), nil
	}

	lookups, err := s.loadCurrentDutyLookups(ctx, currentDuty.duty.ID(), currentDuty.dutyTeam.ID())
	if err != nil {
		return nil, err
	}

	tasks := buildResidentDutyTasks(currentDuty.duty.Tasks(), userID, lookups, view)
	sortResidentDutyTasks(tasks)

	response := buildResidentCurrentDutyResponse(
		currentDuty,
		view,
		tasks,
		lookups.teamMembers,
		len(lookups.userNames),
		s.currentTime(),
	)
	response.Team = s.loadTeamLeaderName(ctx, currentDuty.dutyTeam)
	return response, nil
}

func (s *Service) GetDutyHistory(
	ctx context.Context,
	userID uuid.UUID,
	groupID *uuid.UUID,
) (*dto.ResidentDutyHistoryResponse, error) {
	access, err := s.loadResidentAccessContext(ctx, userID, nil)
	if err != nil {
		return nil, err
	}

	selectedGroup, err := s.resolveResidentSelectedGroup(ctx, access, groupID)
	if err != nil {
		return nil, err
	}
	if !isGroupLeader(selectedGroup, access.resident.ID()) {
		return nil, ErrResidentDutyAccessDenied
	}

	history, err := s.dutyRepo.FindHistoryByGroupID(ctx, selectedGroup.ID())
	if err != nil {
		return nil, fmt.Errorf("load duty history: %w", err)
	}

	items := make([]dto.ResidentDutyHistoryItem, 0, len(history))
	for _, item := range history {
		items = append(items, dto.ResidentDutyHistoryItem{
			ID:             item.DutyID.String(),
			StartDate:      item.Start.Format("2006-01-02"),
			EndDate:        item.End.Format("2006-01-02"),
			TeamLeaderName: item.TeamLeaderName,
			PeriodStatus:   resolveDutyPeriodStatus(item.Start, item.End, s.currentTime()),
			Progress: dto.ResidentDutyProgressSummary{
				TotalCostSum:        item.TotalCostSum,
				TakenCostSum:        item.TakenCostSum,
				TotalTasksCount:     item.TotalTasksCount,
				TakenTasksCount:     item.TakenTasksCount,
				CompletedTasksCount: item.CompletedTasksCount,
				VerifiedTasksCount:  item.VerifiedTasksCount,
			},
		})
	}

	return &dto.ResidentDutyHistoryResponse{
		SelectedGroupID: selectedGroup.ID().String(),
		ShowGroupSelect: access.canSelectAnyGroup && len(access.availableGroups) > 1,
		Groups:          buildResidentDutyGroupOptions(access.availableGroups, access.canSelectAnyGroup),
		Duties:          items,
	}, nil
}

func (s *Service) GetDutyDetails(
	ctx context.Context,
	userID uuid.UUID,
	dutyID uuid.UUID,
) (*dto.ResidentDutyDetailsResponse, error) {
	access, err := s.loadResidentAccessContext(ctx, userID, nil)
	if err != nil {
		return nil, err
	}

	selectedDuty, err := s.dutyRepo.FindByID(ctx, dutyID)
	if err != nil {
		return nil, fmt.Errorf("load duty: %w", err)
	}
	if selectedDuty == nil {
		return nil, ErrResidentDutyNotFound
	}

	dutyTeam, err := s.teamRepo.FindByID(ctx, selectedDuty.TeamID())
	if err != nil {
		return nil, fmt.Errorf("load duty team: %w", err)
	}
	if dutyTeam == nil {
		return nil, ErrResidentDutyNotFound
	}

	selectedGroup, err := s.groupRepo.FindByID(ctx, dutyTeam.GroupID())
	if err != nil {
		return nil, fmt.Errorf("load duty group: %w", err)
	}
	if selectedGroup == nil || !groupIsAccessible(access.availableGroups, selectedGroup.ID()) {
		return nil, ErrResidentDutyNotFound
	}
	if !isGroupLeader(selectedGroup, access.resident.ID()) {
		return nil, ErrResidentDutyAccessDenied
	}

	dutyContext := &currentDutyContext{
		resident:      access.resident,
		residentTeam:  access.residentTeam,
		residentGroup: access.residentGroup,
		dormitory:     access.dormitory,
		selectedGroup: selectedGroup,
		dutyTeam:      dutyTeam,
		duty:          selectedDuty,
		groups:        access.availableGroups,
	}

	lookups, err := s.loadCurrentDutyLookups(ctx, selectedDuty.ID(), dutyTeam.ID())
	if err != nil {
		return nil, err
	}

	view := buildReadonlyDutyDetailsView()
	tasks := buildResidentDutyTasks(selectedDuty.Tasks(), userID, lookups, view)
	sortResidentDutyTasks(tasks)

	response := buildResidentCurrentDutyResponse(
		dutyContext,
		view,
		tasks,
		lookups.teamMembers,
		len(lookups.userNames),
		s.currentTime(),
	)
	response.Team = s.loadTeamLeaderName(ctx, dutyTeam)
	return response, nil
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

func (s *Service) ReopenTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	return s.cleaningUC.ReopenTask(ctx, taskID, userID)
}

func (s *Service) loadCurrentDutyContext(
	ctx context.Context,
	userID uuid.UUID,
	groupID *uuid.UUID,
	dormitoryID *int64,
) (*currentDutyContext, error) {
	access, err := s.loadResidentAccessContext(ctx, userID, dormitoryID)
	if err != nil {
		return nil, err
	}

	var selectedGroup *structure.Group
	if access.residentGroup == nil && !access.canSelectAnyGroup && groupID == nil {
		selectedGroup, err = s.resolveObserverGroupFallback(ctx, access.availableGroups)
		if err != nil {
			return nil, err
		}
	} else {
		selectedGroup, err = s.resolveResidentSelectedGroup(ctx, access, groupID)
		if err != nil {
			return nil, err
		}
	}

	activeDutyTeam, activeDuty, err := s.loadActiveDutyForGroup(ctx, selectedGroup.ID())
	if err != nil {
		return nil, fmt.Errorf("load active duty for group: %w", err)
	}

	displayDutyTeam := activeDutyTeam
	displayDuty := activeDuty
	if access.residentTeam != nil && access.residentGroup != nil && access.residentGroup.ID() == selectedGroup.ID() {
		showResidentUnfinishedDuty := activeDutyTeam == nil || activeDutyTeam.ID() != access.residentTeam.ID()
		if showResidentUnfinishedDuty {
			unfinishedDuty, err := s.loadLatestUnfinishedPastDutyForTeam(ctx, access.residentTeam.ID(), selectedGroup.ID())
			if err != nil {
				return nil, fmt.Errorf("load unfinished team duty: %w", err)
			}
			if unfinishedDuty != nil {
				displayDutyTeam = access.residentTeam
				displayDuty = unfinishedDuty
			}
		}
	}
	// Team soft deletion clears current user.team_id but does not remove the
	// participant snapshot of an outstanding past Duty. Keep that Duty reachable.
	if displayDuty == activeDuty && s.participantRepo != nil {
		unfinishedDuty, err := s.loadLatestUnfinishedPastDutyForParticipant(ctx, userID, selectedGroup.ID())
		if err != nil {
			return nil, fmt.Errorf("load unfinished participant duty: %w", err)
		}
		if unfinishedDuty != nil {
			leaderID := uuid.Nil
			if unfinishedDuty.LeaderID() != nil {
				leaderID = *unfinishedDuty.LeaderID()
			}
			displayDuty = unfinishedDuty
			displayDutyTeam = structure.RestoreTeam(unfinishedDuty.TeamID(), selectedGroup.ID(), leaderID, "", 1)
		}
	}

	return &currentDutyContext{
		resident:          access.resident,
		residentTeam:      access.residentTeam,
		residentGroup:     access.residentGroup,
		dormitory:         access.dormitory,
		myDormitoryID:     access.myDormitoryID,
		myResidentTeam:    access.myResidentTeam,
		myResidentGroup:   access.myResidentGroup,
		canSelectAnyGroup: access.canSelectAnyGroup,
		selectedGroup:     selectedGroup,
		activeDutyTeam:    activeDutyTeam,
		activeDuty:        activeDuty,
		dutyTeam:          displayDutyTeam,
		duty:              displayDuty,
		groups:            access.availableGroups,
	}, nil
}

func (s *Service) loadResidentAccessContext(ctx context.Context, userID uuid.UUID, requestedDormitoryID *int64) (*residentAccessContext, error) {
	resident, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load resident: %w", err)
	}
	if resident == nil {
		return nil, fmt.Errorf("resident not found")
	}
	if resident.DormitoryID() == nil {
		return nil, fmt.Errorf("resident is not assigned to a dormitory")
	}

	canSelectAnyGroup, err := s.dormRepo.ExistsByLeaderID(ctx, resident.ID())
	if err != nil {
		return nil, fmt.Errorf("check dormitory leadership: %w", err)
	}

	myDormitory, err := s.dormRepo.FindByID(ctx, *resident.DormitoryID())
	if err != nil {
		return nil, fmt.Errorf("load dormitory: %w", err)
	}
	if myDormitory == nil {
		return nil, fmt.Errorf("dormitory not found")
	}

	myGroups, err := s.groupRepo.FindByDormitoryID(ctx, *resident.DormitoryID())
	if err != nil {
		return nil, fmt.Errorf("load dormitory groups: %w", err)
	}
	if len(myGroups) == 0 {
		return nil, fmt.Errorf("no groups found for resident dormitory")
	}
	sortGroups(myGroups)

	myResidentTeam, myResidentGroup, err := s.resolveResidentAffiliation(ctx, resident, myGroups, canSelectAnyGroup)
	if err != nil {
		return nil, err
	}

	effectiveDormitoryID := *resident.DormitoryID()
	if requestedDormitoryID != nil && canSelectAnyGroup {
		effectiveDormitoryID = *requestedDormitoryID
	}

	dormitory := myDormitory
	groups := myGroups
	residentTeam := myResidentTeam
	residentGroup := myResidentGroup
	if effectiveDormitoryID != *resident.DormitoryID() {
		dormitory, err = s.dormRepo.FindByID(ctx, effectiveDormitoryID)
		if err != nil {
			return nil, fmt.Errorf("load selected dormitory: %w", err)
		}
		if dormitory == nil {
			return nil, ErrResidentDutyAccessDenied
		}

		groups, err = s.groupRepo.FindByDormitoryID(ctx, effectiveDormitoryID)
		if err != nil {
			return nil, fmt.Errorf("load selected dormitory groups: %w", err)
		}
		sortGroups(groups)
		residentTeam, residentGroup, err = s.resolveResidentAffiliation(ctx, resident, groups, canSelectAnyGroup)
		if err != nil {
			return nil, err
		}
	}

	return &residentAccessContext{
		resident:          resident,
		residentTeam:      residentTeam,
		residentGroup:     residentGroup,
		dormitory:         dormitory,
		myDormitoryID:     *resident.DormitoryID(),
		myResidentTeam:    myResidentTeam,
		myResidentGroup:   myResidentGroup,
		availableGroups:   groups,
		canSelectAnyGroup: canSelectAnyGroup,
	}, nil
}

func (s *Service) resolveResidentSelectedGroup(
	ctx context.Context,
	access *residentAccessContext,
	groupID *uuid.UUID,
) (*structure.Group, error) {
	if access == nil {
		return nil, ErrResidentDutyGroupNotFound
	}

	if access.residentGroup == nil && !access.canSelectAnyGroup && groupID == nil {
		return s.resolveObserverGroupFallback(ctx, access.availableGroups)
	}

	selectedGroup, err := resolveSelectedGroup(
		access.availableGroups,
		access.residentGroup,
		groupID,
		access.canSelectAnyGroup,
	)
	if err != nil {
		return nil, ErrResidentDutyGroupNotFound
	}

	return selectedGroup, nil
}

func (s *Service) resolveResidentAffiliation(
	ctx context.Context,
	resident *user.User,
	groups []*structure.Group,
	isDormitoryLeader bool,
) (*structure.Team, *structure.Group, error) {
	if resident.TeamID() != nil {
		residentTeam, err := s.teamRepo.FindByID(ctx, *resident.TeamID())
		if err != nil {
			return nil, nil, fmt.Errorf("load resident team: %w", err)
		}
		if residentTeam != nil {
			residentGroup, err := s.groupRepo.FindByID(ctx, residentTeam.GroupID())
			if err != nil {
				return nil, nil, fmt.Errorf("load resident group: %w", err)
			}
			if residentGroup != nil {
				return residentTeam, residentGroup, nil
			}
		}
	}

	for _, group := range groups {
		teams, err := s.teamRepo.FindByGroupID(ctx, group.ID())
		if err != nil {
			return nil, nil, fmt.Errorf("load group teams: %w", err)
		}
		for _, team := range teams {
			if isTeamLeader(team, resident.ID()) {
				return team, group, nil
			}
		}
	}

	for _, group := range groups {
		if isGroupLeader(group, resident.ID()) {
			return nil, group, nil
		}
	}

	if isDormitoryLeader {
		return nil, nil, nil
	}

	return nil, nil, nil
}

func resolveSelectedGroup(
	groups []*structure.Group,
	residentGroup *structure.Group,
	groupID *uuid.UUID,
	canSelectGroup bool,
) (*structure.Group, error) {
	if !canSelectGroup || groupID == nil {
		if residentGroup != nil {
			return residentGroup, nil
		}
		if canSelectGroup && len(groups) > 0 {
			return groups[0], nil
		}
		return nil, fmt.Errorf("resident is not assigned to a group")
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

func (s *Service) resolveObserverGroupFallback(
	ctx context.Context,
	groups []*structure.Group,
) (*structure.Group, error) {
	for _, group := range groups {
		dutyTeam, duty, err := s.loadActiveDutyForGroup(ctx, group.ID())
		if err != nil {
			return nil, fmt.Errorf("resolve observer group fallback: %w", err)
		}
		if dutyTeam != nil && duty != nil {
			return group, nil
		}
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("resident is not assigned to a group")
	}

	return groups[0], nil
}

func (s *Service) loadLatestUnfinishedPastDutyForTeam(
	ctx context.Context,
	teamID uuid.UUID,
	groupID uuid.UUID,
) (*dutydomain.Duty, error) {
	duties, err := s.dutyRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("load duties by group: %w", err)
	}

	now := s.currentTime()
	var selected *dutydomain.Duty
	for _, currentDuty := range duties {
		if currentDuty.TeamID() != teamID {
			continue
		}
		if resolveDutyPeriodStatus(currentDuty.Start(), currentDuty.End(), now) != dutyPeriodStatusPast {
			continue
		}
		if !dutyHasOutstandingTasks(currentDuty) {
			continue
		}
		if selected == nil || dutyIsLaterThan(currentDuty, selected) {
			selected = currentDuty
		}
	}

	return selected, nil
}

func (s *Service) loadLatestUnfinishedPastDutyForParticipant(
	ctx context.Context,
	participantID uuid.UUID,
	groupID uuid.UUID,
) (*dutydomain.Duty, error) {
	duties, err := s.dutyRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("load duties by group: %w", err)
	}

	now := s.currentTime()
	var selected *dutydomain.Duty
	for _, currentDuty := range duties {
		if resolveDutyPeriodStatus(currentDuty.Start(), currentDuty.End(), now) != dutyPeriodStatusPast || !dutyHasOutstandingTasks(currentDuty) {
			continue
		}
		participants, err := s.participantRepo.List(ctx, currentDuty.ID(), false)
		if err != nil {
			return nil, fmt.Errorf("load duty participants: %w", err)
		}
		isParticipant := false
		for _, participant := range participants {
			if participant.ParticipantID == participantID {
				isParticipant = true
				break
			}
		}
		if isParticipant && (selected == nil || dutyIsLaterThan(currentDuty, selected)) {
			selected = currentDuty
		}
	}
	return selected, nil
}

func (s *Service) loadCurrentDutyLookups(
	ctx context.Context,
	dutyID uuid.UUID,
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

	userNames, err := s.userNamesByDutyID(ctx, dutyID, teamID)
	if err != nil {
		return nil, err
	}

	return &currentDutyLookups{
		taskDefinitions: taskDefinitions,
		areas:           areas,
		userNames:       userNames,
		teamMembers:     teamMembersFromNames(userNames),
	}, nil
}

func (s *Service) userNamesByDutyID(ctx context.Context, dutyID, teamID uuid.UUID) (map[uuid.UUID]string, error) {
	if s.participantRepo == nil {
		return s.userNamesByID(ctx, teamID)
	}
	participants, err := s.participantRepo.List(ctx, dutyID, false)
	if err != nil {
		return nil, fmt.Errorf("load duty participants: %w", err)
	}
	result := make(map[uuid.UUID]string, len(participants))
	for _, participant := range participants {
		resident, err := s.userRepo.FindByID(ctx, participant.ParticipantID)
		if err != nil {
			return nil, fmt.Errorf("load duty participant: %w", err)
		}
		if resident != nil {
			result[resident.ID()] = formatUserName(resident)
		}
	}
	return result, nil
}

func buildResidentDutyTasks(
	dutyTasks []*dutydomain.DutyTask,
	userID uuid.UUID,
	lookups *currentDutyLookups,
	view residentDutyView,
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
			view,
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
	view residentDutyView,
) dto.ResidentDutyTask {
	status := resolveTaskStatus(dutyTask)
	isMine := dutyTask.AssigneeID() != nil && *dutyTask.AssigneeID() == userID
	canTake, canReturn, canComplete, canOpen, canVerify, canReviewOpen := buildTaskPermissions(
		status,
		isMine,
		view.canManageTasks,
		view.canVerifyTasks,
	)
	assigneeID, assigneeName := resolveAssignee(dutyTask, userNames)

	return dto.ResidentDutyTask{
		ID:            dutyTask.ID().String(),
		AreaID:        area.ID(),
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
	view residentDutyView,
	tasks []dto.ResidentDutyTask,
	teamMembers []dto.ResidentDutyTeamMember,
	residentCount int,
	now time.Time,
) *dto.ResidentCurrentDutyResponse {
	response := &dto.ResidentCurrentDutyResponse{
		DormitoryID:           currentDuty.selectedGroup.DormitoryID(),
		MyDormitoryID:         &currentDuty.myDormitoryID,
		SelectedGroupID:       currentDuty.selectedGroup.ID().String(),
		Groups:                buildResidentDutyGroupOptions(currentDuty.groups, view.showGroupSelect),
		MyGroup:               buildResidentMyGroupOption(currentDuty.myResidentGroup),
		Group:                 currentDuty.selectedGroup.Name(),
		CanManageTasks:        view.canManageTasks,
		CanManageDutySettings: canOpenDutySettings(currentDuty, now),
		ReadOnly:              view.readOnly,
		ShowGroupSelect:       view.showGroupSelect,
		VisibleTabs:           view.visibleTabs,
		NoticeMessage:         view.noticeMessage,
		NoticeTone:            view.noticeTone,
		TeamMembers:           teamMembers,
		Tasks:                 tasks,
	}

	if currentDuty.duty == nil || currentDuty.dutyTeam == nil {
		return response
	}

	return &dto.ResidentCurrentDutyResponse{
		DormitoryID:           response.DormitoryID,
		MyDormitoryID:         response.MyDormitoryID,
		SelectedGroupID:       response.SelectedGroupID,
		HasActiveDuty:         currentDuty.duty != nil,
		CanManageTasks:        response.CanManageTasks,
		CanManageDutySettings: response.CanManageDutySettings,
		ReadOnly:              response.ReadOnly,
		ShowGroupSelect:       response.ShowGroupSelect,
		VisibleTabs:           response.VisibleTabs,
		NoticeMessage:         response.NoticeMessage,
		NoticeTone:            response.NoticeTone,
		PeriodStatus:          resolveDutyPeriodStatus(currentDuty.duty.Start(), currentDuty.duty.End(), now),
		Groups:                response.Groups,
		MyGroup:               response.MyGroup,
		DutyID:                currentDuty.duty.ID().String(),
		Group:                 currentDuty.selectedGroup.Name(),
		Team:                  teamLeaderName(currentDuty.dutyTeam.LeaderID(), teamMembers),
		StartDate:             currentDuty.duty.Start().Format("2006-01-02"),
		EndDate:               currentDuty.duty.End().Format("2006-01-02"),
		CostPerResidentGoal:   calculateCostPerResidentGoal(tasks, residentCount),
		MyTakenCostSum:        countMyTakenCost(tasks),
		TeamMembers:           response.TeamMembers,
		Tasks:                 response.Tasks,
	}
}

func canOpenDutySettings(currentDuty *currentDutyContext, now time.Time) bool {
	if currentDuty == nil || currentDuty.resident == nil {
		return false
	}
	actorID := currentDuty.resident.ID()
	if isGroupLeader(currentDuty.selectedGroup, actorID) || isDormitoryLeader(currentDuty.dormitory, actorID) {
		return true
	}
	return currentDuty.duty != nil && currentDuty.duty.IsActiveAt(now) && currentDuty.duty.LeaderID() != nil && *currentDuty.duty.LeaderID() == actorID
}

func teamLeaderName(leaderID uuid.UUID, members []dto.ResidentDutyTeamMember) string {
	for _, member := range members {
		if member.ID == leaderID.String() {
			return member.Name
		}
	}
	return "Глава команды не назначен"
}

func buildResidentDutyGroupOptions(groups []*structure.Group, enabled bool) []dto.ResidentDutyGroupOption {
	if !enabled {
		return nil
	}

	options := make([]dto.ResidentDutyGroupOption, 0, len(groups))
	for _, group := range groups {
		options = append(options, dto.ResidentDutyGroupOption{
			ID:   group.ID().String(),
			Name: group.Name(),
		})
	}

	return options
}

func buildResidentMyGroupOption(group *structure.Group) *dto.ResidentDutyGroupOption {
	if group == nil {
		return nil
	}

	return &dto.ResidentDutyGroupOption{
		ID:   group.ID().String(),
		Name: group.Name(),
	}
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

func teamMembersFromNames(userNames map[uuid.UUID]string) []dto.ResidentDutyTeamMember {
	members := make([]dto.ResidentDutyTeamMember, 0, len(userNames))
	for userID, name := range userNames {
		members = append(members, dto.ResidentDutyTeamMember{
			ID:   userID.String(),
			Name: name,
		})
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].Name < members[j].Name
	})

	return members
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
	canVerifyTasks bool,
) (canTake bool, canReturn bool, canComplete bool, canOpen bool, canVerify bool, canReviewOpen bool) {
	if !canManageTasks {
		if canVerifyTasks && status == dto.ResidentDutyTaskStatusCompleted {
			return false, false, false, false, true, true
		}
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
			return false, false, false, true, canVerifyTasks, canVerifyTasks
		}
		return false, false, false, false, canVerifyTasks, canVerifyTasks
	default:
		return false, false, false, false, false, false
	}
}

func (s *Service) resolveResidentDutyView(
	ctx context.Context,
	currentDuty *currentDutyContext,
) (residentDutyView, error) {
	if currentDuty == nil {
		return residentDutyView{}, nil
	}

	isOnDutyTeam := false
	if currentDuty.duty != nil {
		if s.participantRepo != nil {
			active, err := s.participantRepo.IsActive(ctx, currentDuty.duty.ID(), currentDuty.resident.ID())
			if err != nil {
				return residentDutyView{}, fmt.Errorf("check duty participant: %w", err)
			}
			isOnDutyTeam = active
		} else {
			isOnDutyTeam = currentDuty.dutyTeam != nil && currentDuty.resident.TeamID() != nil && *currentDuty.resident.TeamID() == currentDuty.dutyTeam.ID()
		}
	}

	isTeamLeader := currentDuty.duty != nil && currentDuty.duty.LeaderID() != nil && *currentDuty.duty.LeaderID() == currentDuty.resident.ID()
	canManageTasks := false
	canVerifyTasks := false
	readOnly := true
	periodStatus := ""
	hasOutstandingTasks := false

	if currentDuty.duty != nil {
		periodStatus = resolveDutyPeriodStatus(currentDuty.duty.Start(), currentDuty.duty.End(), s.currentTime())
		hasOutstandingTasks = dutyHasOutstandingTasks(currentDuty.duty)
	}

	if isOnDutyTeam && currentDuty.duty != nil {
		allowed, err := s.canInteractWithDuty(ctx, currentDuty)
		if err != nil {
			return residentDutyView{}, err
		}
		if allowed {
			canManageTasks = true
			canVerifyTasks = isTeamLeader
			readOnly = false
		}
	}

	visibleTabs := []string{}
	if currentDuty.duty != nil && currentDuty.dutyTeam != nil {
		if canManageTasks {
			visibleTabs = []string{"all", "mine", "team"}
			if canVerifyTasks {
				visibleTabs = append(visibleTabs, "verification")
			}
		} else {
			visibleTabs = []string{"all", "team"}
		}
	}

	if currentDuty.canSelectAnyGroup {
		return residentDutyView{
			showGroupSelect: true,
			visibleTabs:     visibleTabs,
			readOnly:        readOnly,
			canViewTasks:    currentDuty.duty != nil && currentDuty.dutyTeam != nil,
			canManageTasks:  canManageTasks,
			canVerifyTasks:  canVerifyTasks,
			noticeMessage:   s.resolveCurrentDutyNotice(ctx, currentDuty, isOnDutyTeam, periodStatus, hasOutstandingTasks),
			noticeTone:      s.resolveCurrentDutyNoticeTone(isOnDutyTeam, periodStatus, hasOutstandingTasks),
		}, nil
	}

	if canManageTasks {
		return residentDutyView{
			visibleTabs:    visibleTabs,
			canViewTasks:   true,
			canManageTasks: true,
			canVerifyTasks: canVerifyTasks,
			noticeMessage:  s.resolveCurrentDutyNotice(ctx, currentDuty, isOnDutyTeam, periodStatus, hasOutstandingTasks),
			noticeTone:     s.resolveCurrentDutyNoticeTone(isOnDutyTeam, periodStatus, hasOutstandingTasks),
		}, nil
	}

	return residentDutyView{
		readOnly:      true,
		visibleTabs:   visibleTabs,
		canViewTasks:  currentDuty.duty != nil && currentDuty.dutyTeam != nil,
		noticeMessage: s.resolveCurrentDutyNotice(ctx, currentDuty, false, periodStatus, hasOutstandingTasks),
		noticeTone:    s.resolveCurrentDutyNoticeTone(false, periodStatus, hasOutstandingTasks),
	}, nil
}

func (s *Service) resolveCurrentDutyNotice(
	ctx context.Context,
	currentDuty *currentDutyContext,
	isOnDutyTeam bool,
	periodStatus string,
	hasOutstandingTasks bool,
) string {
	if currentDuty == nil {
		return ""
	}

	if isOnDutyTeam && periodStatus == dutyPeriodStatusPast && hasOutstandingTasks && currentDuty.activeDutyTeam != nil && currentDuty.activeDuty != nil {
		return "Завершите задачи вовремя, чтобы избежать предупреждения."
	}

	if isOnDutyTeam || currentDuty.activeDutyTeam == nil {
		return ""
	}

	return fmt.Sprintf("На этой неделе ответственный за дежурство — %s", s.loadTeamLeaderName(ctx, currentDuty.activeDutyTeam))
}

func (s *Service) resolveCurrentDutyNoticeTone(
	isOnDutyTeam bool,
	periodStatus string,
	hasOutstandingTasks bool,
) string {
	if isOnDutyTeam && periodStatus == dutyPeriodStatusPast && hasOutstandingTasks {
		return "warning"
	}

	return "info"
}

func (s *Service) canInteractWithDuty(ctx context.Context, currentDuty *currentDutyContext) (bool, error) {
	if currentDuty == nil || currentDuty.duty == nil || currentDuty.dutyTeam == nil {
		return false, nil
	}
	if s.participantRepo != nil {
		active, err := s.participantRepo.IsActive(ctx, currentDuty.duty.ID(), currentDuty.resident.ID())
		if err != nil {
			return false, err
		}
		if !active {
			return false, nil
		}
	} else if currentDuty.resident.TeamID() == nil || *currentDuty.resident.TeamID() != currentDuty.dutyTeam.ID() {
		return false, nil
	}

	now := s.currentTime()
	periodStatus := resolveDutyPeriodStatus(currentDuty.duty.Start(), currentDuty.duty.End(), now)
	switch periodStatus {
	case dutyPeriodStatusActive:
		return true, nil
	case dutyPeriodStatusFuture:
		return false, nil
	}

	return dutyHasOutstandingTasks(currentDuty.duty), nil
}

func buildReadonlyDutyDetailsView() residentDutyView {
	return residentDutyView{
		visibleTabs:    []string{"all", "team"},
		readOnly:       true,
		canViewTasks:   true,
		canManageTasks: false,
		canVerifyTasks: false,
	}
}

func resolveDutyPeriodStatus(startDate, endDate, now time.Time) string {
	if startDate.After(now) {
		return dutyPeriodStatusFuture
	}
	if !endDate.After(now) {
		return dutyPeriodStatusPast
	}
	return dutyPeriodStatusActive
}

func (s *Service) currentTime() time.Time {
	if s.now == nil {
		s.now = time.Now
	}
	now := s.now()
	if s.location == nil {
		return now
	}
	return now.In(s.location)
}

func sortGroups(groups []*structure.Group) {
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].Name() != groups[j].Name() {
			return groups[i].Name() < groups[j].Name()
		}
		return groups[i].ID().String() < groups[j].ID().String()
	})
}

func groupIsAccessible(groups []*structure.Group, groupID uuid.UUID) bool {
	for _, group := range groups {
		if group.ID() == groupID {
			return true
		}
	}
	return false
}

func dutyHasOutstandingTasks(currentDuty *dutydomain.Duty) bool {
	if currentDuty == nil {
		return false
	}

	for _, task := range currentDuty.Tasks() {
		if task.VerificationDate() == nil {
			return true
		}
	}

	return false
}

func dutyIsLaterThan(left *dutydomain.Duty, right *dutydomain.Duty) bool {
	if left.SequenceNumber() != right.SequenceNumber() {
		return left.SequenceNumber() > right.SequenceNumber()
	}
	if !left.Start().Equal(right.Start()) {
		return left.Start().After(right.Start())
	}
	return left.ID().String() > right.ID().String()
}

func (s *Service) loadTeamLeaderName(ctx context.Context, dutyTeam *structure.Team) string {
	if dutyTeam == nil {
		return "Глава команды не назначен"
	}

	leader, err := s.userRepo.FindByID(ctx, dutyTeam.LeaderID())
	if err != nil || leader == nil {
		return "Глава команды не назначен"
	}

	return formatUserName(leader)
}

func isDormitoryLeader(dormitory *structure.Dormitory, userID uuid.UUID) bool {
	return dormitory != nil && dormitory.LeaderID() != nil && *dormitory.LeaderID() == userID
}

func isGroupLeader(group *structure.Group, userID uuid.UUID) bool {
	return group != nil && group.LeaderID() != nil && *group.LeaderID() == userID
}

func isTeamLeader(team *structure.Team, userID uuid.UUID) bool {
	return team != nil && team.LeaderID() == userID
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
