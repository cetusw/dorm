package service

import (
	"dorm/internal/dorm/application/model"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CleaningService struct {
	sheetsService   *SheetsService
	userService     *UserService
	dutyService     *DutyService
	dutyTaskService *DutyTaskService
	teamService     *TeamService
}

func NewCleaningService(
	sheetsService *SheetsService,
	userService *UserService,
	dutyService *DutyService,
	dutyTaskService *DutyTaskService,
	teamService *TeamService,
) *CleaningService {
	return &CleaningService{
		sheetsService:   sheetsService,
		userService:     userService,
		dutyService:     dutyService,
		dutyTaskService: dutyTaskService,
		teamService:     teamService,
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
	var newDutyID uuid.UUID
	newDutyID, err = s.dutyService.CreateNewDuty(newDutyTeamID, startTime, endTime)
	if err != nil {
		return err
	}
	err = s.dutyTaskService.SetDutyTasks(newDutyID)
	if err != nil {
		return err
	}
	var teamColor string
	teamColor, err = s.teamService.GetTeamColor(newDutyTeamID)
	if err != nil {
		return err
	}
	var dutyTasksReadable []model.DutyTaskReadable
	dutyTasksReadable, err = s.dutyTaskService.GetDutyTasksReadable(newDutyID)
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

func (s *CleaningService) HandleTaskAssignment(chatID int64, taskID uuid.UUID, duty *model.Duty) error {
	user, err := s.userService.GetUser(chatID)
	if err != nil {
		return fmt.Errorf("ERROR: failed to get user while getting duty tasks: %v", err)
	}
	err = s.dutyTaskService.SetDutyTaskAssigneeIDByDutyID(user.UserID, taskID, duty.DutyID)
	if err != nil {
		return fmt.Errorf("ERROR: failed to set current duty task assignee: %v", err)
	}

	return s.UpdateCurrentSheet(duty)
}
func (s *CleaningService) HandleTaskCompletion(memberID int64, taskID int) {
}

func (s *CleaningService) HandleTaskVerification(memberID int64, taskID int) {
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
