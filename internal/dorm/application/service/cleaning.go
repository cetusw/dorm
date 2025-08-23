package service

import (
	"dorm/internal/dorm/application/model"
	"fmt"
	"time"
)

type CleaningService struct {
	sheetsService   *SheetsService
	userService     *UserService
	dutyService     *DutyService
	dutyTaskService *DutyTaskService
	teamService     *TeamService
	taskService     *TaskService
}

func NewCleaningService(
	sheetsService *SheetsService,
	userService *UserService,
	dutyService *DutyService,
	dutyTaskService *DutyTaskService,
	teamService *TeamService,
	taskService *TaskService,
) *CleaningService {
	return &CleaningService{
		sheetsService:   sheetsService,
		userService:     userService,
		dutyService:     dutyService,
		dutyTaskService: dutyTaskService,
		teamService:     teamService,
		taskService:     taskService,
	}
}

func (s *CleaningService) StartNewWeek() error {
	startTime := time.Now()
	endTime := startTime.AddDate(0, 0, 6)
	sheetTitle := fmt.Sprintf("%s-%s", startTime.Format("02.01"), endTime.Format("02.01"))

	newDutyTeamID, err := s.getNewDutyTeamID()
	if err != nil {
		return err
	}
	var duty *model.Duty
	duty, err = s.dutyService.CreateNewDuty(newDutyTeamID, startTime, endTime)
	if err != nil {
		return err
	}
	team, err := s.teamService.GetTeam(duty.TeamID)
	if err != nil {
		return err
	}
	taskIDs, err := s.taskService.GetAllTaskIDs()
	if err != nil {
		return err
	}
	err = s.dutyTaskService.SetDutyTasks(duty.DutyID, taskIDs, team.TeamLeaderID)
	if err != nil {
		return err
	}
	var teamColor string
	teamColor, err = s.teamService.GetTeamColor(newDutyTeamID)
	if err != nil {
		return err
	}
	var dutyTasksReadable []model.DutyTaskView
	dutyTasksReadable, err = s.dutyTaskService.GetDutyTasksReadable(duty.DutyID)
	if err != nil {
		return err
	}
	users, err := s.userService.GetAllUsers()
	if err != nil {
		return err
	}
	err = s.sheetsService.CreateWeeklySheet(sheetTitle, teamColor, newDutyTeamID, dutyTasksReadable, users)
	if err != nil {
		return err
	}
	return nil
}

func (s *CleaningService) UpdateCurrentSheet(duty *model.Duty) error {
	sheetTitle := fmt.Sprintf("%s-%s", duty.Start.Format("02.01"), duty.End.Format("02.01"))

	dutyTasksReadable, err := s.dutyTaskService.GetDutyTasksReadable(duty.DutyID)
	if err != nil {
		return fmt.Errorf("failed to get readable duty tasks for sheet update: %w", err)
	}

	err = s.sheetsService.UpdateWeeklySheet(sheetTitle, dutyTasksReadable)
	if err != nil {
		return fmt.Errorf("failed to update weekly sheet: %w", err)
	}

	return nil
}

func (s *CleaningService) getNewDutyTeamID() (int, error) {
	teamIDs, err := s.teamService.GetSortedTeamIDs()
	if err != nil {
		return 0, err
	}
	var lastDutyTeamID int
	lastDutyTeamID, err = s.dutyService.GetLastDutyTeamID()
	if err != nil {
		return 0, err
	}

	nextTeamIndex := 0

	for i, id := range teamIDs {
		if id == lastDutyTeamID {
			nextTeamIndex = (i + 1) % len(teamIDs)
			break
		}
	}

	return teamIDs[nextTeamIndex], nil
}
