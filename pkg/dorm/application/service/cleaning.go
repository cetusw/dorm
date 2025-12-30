package service

import (
	"fmt"
	"sort"
	"time"

	"dorm/pkg/common/utils"
	"dorm/pkg/dorm/application/model"
	"dorm/pkg/dorm/application/service/sheets"

	"github.com/google/uuid"
)

type CleaningService struct {
	sheetsService   sheets.SheetsService
	userService     *UserService
	dutyService     *DutyService
	dutyTaskService *DutyTaskService
	teamService     *TeamService
	taskService     *TaskService
	groupService    *GroupService
	areaService     *AreaService
}

func NewCleaningService(
	sheetsService sheets.SheetsService,
	userService *UserService,
	dutyService *DutyService,
	dutyTaskService *DutyTaskService,
	teamService *TeamService,
	taskService *TaskService,
	groupService *GroupService,
	areaService *AreaService,
) *CleaningService {
	return &CleaningService{
		sheetsService:   sheetsService,
		userService:     userService,
		dutyService:     dutyService,
		dutyTaskService: dutyTaskService,
		teamService:     teamService,
		taskService:     taskService,
		groupService:    groupService,
		areaService:     areaService,
	}
}

func (s *CleaningService) StartNewWeek() error {
	startTime := utils.NowMoscow()
	endTime := startTime.AddDate(0, 0, 6)
	newDutyTeams, err := s.getNewDutyTeams()
	if err != nil {
		return err
	}
	duties, err := s.dutyService.CreateNewDuties(newDutyTeams, startTime, endTime)
	if err != nil {
		return err
	}
	err = s.assignPrivateAreaTasksToDuties(duties)
	if err != nil {
		return err
	}
	err = s.assignPublicAreaTasksToDuties(duties)
	if err != nil {
		return err
	}
	err = s.createDutySheets(duties)
	if err != nil {
		return err
	}
	return nil
}

func (s *CleaningService) UpdateCurrentSheet(user model.User) error {
	team, err := s.teamService.GetTeam(*user.TeamID)
	if err != nil {
		return err
	}
	group, err := s.groupService.GetGroup(team.GroupID)
	if err != nil {
		return err
	}
	duty, err := s.dutyService.GetUserLastDuty(user.ID)
	if err != nil {
		return err
	}

	dutyTasksView, err := s.dutyTaskService.GetDutyTasksView(duty.ID)
	if err != nil {
		return fmt.Errorf("failed to get readable duty tasks for sheet update: %w", err)
	}

	users, err := s.userService.GetUsersByTeamID(*user.TeamID)
	if err != nil {
		return fmt.Errorf("failed to get users for sheet update: %w", err)
	}

	sheetData := model.SheetData{
		SpreadsheetID: group.SpreadsheetID,
		Title:         s.createSheetTitle(*duty),
		Tasks:         dutyTasksView,
		Order:         team.Order,
		TeamColor:     "",
		Users:         users,
	}

	err = s.sheetsService.UpdateDutySheet(sheetData)
	if err != nil {
		return fmt.Errorf("failed to update weekly sheet: %w", err)
	}

	return nil
}

func (s *CleaningService) getNewDutyTeams() ([]model.Team, error) {
	var dutyTeams []model.Team
	groups, err := s.groupService.GetAllGroups()
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		nextDutyTeam, err := s.getNextDutyTeamInGroup(group)
		if err != nil {
			return nil, fmt.Errorf("failed to get next duty team for group %s: %w", group.Name, err)
		}

		if nextDutyTeam != nil {
			dutyTeams = append(dutyTeams, *nextDutyTeam)
		}
	}

	return dutyTeams, nil
}

func (s *CleaningService) getNextDutyTeamInGroup(group model.Group) (*model.Team, error) {
	var teamID uuid.UUID
	teams, err := s.teamService.GetGroupTeams(group.ID)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, fmt.Errorf("no teams found in team group %s", group.ID)
	}
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Order < teams[j].Order
	})
	lastDuty, err := s.dutyService.GetGroupLastDuty(group.ID)
	if err != nil {
		return nil, err
	}
	if lastDuty == nil {
		firstTeamInGroup, err := s.teamService.GetFirstGroupTeam(group.ID)
		if err != nil {
			return nil, err
		}
		teamID = firstTeamInGroup.ID
	} else {
		teamID = lastDuty.TeamID
	}
	lastDutyTeam, err := s.teamService.GetTeam(teamID)
	if err != nil {
		return nil, err
	}
	nextOrder, err := s.getNextOrder(lastDutyTeam.Order, len(teams))
	if err != nil {
		return nil, err
	}

	return s.teamService.GetTeamByGroupIDAndOrder(group.ID, nextOrder)
}

func (s *CleaningService) getNextOrder(currentOrder int, totalTeams int) (int, error) {
	if totalTeams <= 0 {
		return 0, fmt.Errorf("total teams must be positive")
	}
	if currentOrder < 1 || currentOrder > totalTeams {
		return 0, fmt.Errorf("current order %d is out of bounds [1, %d]", currentOrder, totalTeams)
	}

	nextOrder := (currentOrder % totalTeams) + 1

	return nextOrder, nil
}

func (s *CleaningService) assignPrivateAreaTasksToDuties(duties []model.Duty) error {
	if len(duties) == 0 {
		return nil
	}
	var dutyTasks []model.DutyTask
	for _, duty := range duties {
		team, err := s.teamService.GetTeam(duty.TeamID)
		if err != nil {
			return err
		}
		privateTasks, err := s.taskService.GetGroupTasks(team.GroupID)
		if err != nil {
			return err
		}
		for _, task := range privateTasks {
			if !s.isTaskDue(task) {
				continue
			}
			dutyTask := model.DutyTask{
				ID:         uuid.New(),
				DutyID:     duty.ID,
				TaskID:     task.ID,
				ReviewerID: team.LeaderID,
			}
			dutyTasks = append(dutyTasks, dutyTask)
		}
	}

	return s.dutyTaskService.SetDutyTasksBatch(dutyTasks)
}

func (s *CleaningService) groupPublicTasksByArea() (map[int][]model.Task, error) {
	publicTasks, err := s.taskService.GetPublicTasks()
	if err != nil {
		return nil, err
	}

	tasksByArea := make(map[int][]model.Task)
	for _, task := range publicTasks {
		tasksByArea[task.AreaID] = append(tasksByArea[task.AreaID], task)
	}

	return tasksByArea, nil
}

func (s *CleaningService) buildDutyToTeamMap(duties []model.Duty) (map[uuid.UUID]model.Team, error) {
	dutyToTeamMap := make(map[uuid.UUID]model.Team)
	for _, duty := range duties {
		team, err := s.teamService.GetTeam(duty.TeamID)
		if err != nil {
			return nil, fmt.Errorf("failed to get team %s: %w", duty.TeamID, err)
		}
		dutyToTeamMap[duty.ID] = *team
	}
	return dutyToTeamMap, nil
}

func (s *CleaningService) assignPublicAreaTasksToDuties(duties []model.Duty) error {
	if len(duties) == 0 {
		return nil
	}
	publicAreas, err := s.areaService.GetPublicAreas()
	if err != nil || len(publicAreas) == 0 {
		return err
	}
	tasksByArea, err := s.groupPublicTasksByArea()
	if err != nil {
		return err
	}
	teamForDuty, err := s.buildDutyToTeamMap(duties)
	if err != nil {
		return err
	}

	dutyTasks := s.assignAreasRoundRobin(duties, publicAreas, teamForDuty, tasksByArea)

	return s.dutyTaskService.SetDutyTasksBatch(dutyTasks)
}

func (s *CleaningService) assignAreasRoundRobin(
	duties []model.Duty,
	publicAreas []model.Area,
	teamForDuty map[uuid.UUID]model.Team,
	tasksByArea map[int][]model.Task,
) []model.DutyTask {
	sort.Slice(publicAreas, func(i, j int) bool {
		return publicAreas[i].ID < publicAreas[j].ID
	})

	currentWeek, err := s.dutyService.GetCurrentWeek()
	if err != nil {
		return nil
	}

	var allDutyTasks []model.DutyTask
	dutyIndex := currentWeek % len(duties)
	for _, area := range publicAreas {
		duty := duties[dutyIndex]
		team := teamForDuty[duty.ID]

		for _, task := range tasksByArea[area.ID] {
			if !s.isTaskDue(task) {
				continue
			}
			allDutyTasks = append(allDutyTasks, model.DutyTask{
				ID:         uuid.New(),
				DutyID:     duty.ID,
				TaskID:     task.ID,
				ReviewerID: team.LeaderID,
			})
		}
		dutyIndex = (dutyIndex + 1) % len(duties)
	}
	return allDutyTasks
}

func (s *CleaningService) createDutySheets(duties []model.Duty) error {
	for _, duty := range duties {
		team, err := s.teamService.GetTeam(duty.TeamID)
		if err != nil {
			return err
		}
		group, err := s.groupService.GetGroup(team.GroupID)
		if err != nil {
			return err
		}
		color, err := s.teamService.GetTeamColor(duty.TeamID)
		if err != nil {
			return err
		}
		dutyTasksView, err := s.dutyTaskService.GetDutyTasksView(duty.ID)
		if err != nil {
			return err
		}
		users, err := s.userService.GetUsersByTeamID(duty.TeamID)
		if err != nil {
			return err
		}
		dutySheet := model.SheetData{
			SpreadsheetID: group.SpreadsheetID,
			Title:         s.createSheetTitle(duty),
			TeamColor:     color,
			Order:         team.Order,
			Tasks:         dutyTasksView,
			Users:         users,
		}
		err = s.sheetsService.CreateDutySheet(dutySheet)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *CleaningService) isTaskDue(task model.Task) bool {
	lastTaskDuty, err := s.dutyService.GetTaskLastDuty(task.ID)
	if err != nil {
		return true
	}

	nextAllowedDate := lastTaskDuty.Start.AddDate(0, 0, task.Frequency)

	return !time.Now().Before(nextAllowedDate)
}

func (s *CleaningService) createSheetTitle(duty model.Duty) string {
	// TODO: исправить часовые зоны DORM-17
	return fmt.Sprintf("%s-%s", duty.Start.Format("02.01"), duty.End.Format("02.01"))
}
