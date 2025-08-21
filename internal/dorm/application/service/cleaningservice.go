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
	var dutyTasks []model.DutyTaskReadable
	dutyTasks, err = s.dutyTaskService.GetDutyTasks(newDutyID)
	if err != nil {
		return err
	}
	users, err := s.userService.GetAllUsers()
	if err != nil {
		return err
	}
	err = s.sheetsService.CreateWeeklySheet(sheetTitle, teamColor, newDutyTeamID, dutyTasks, users)
	if err != nil {
		return err
	}
	return nil
}

func (s *CleaningService) HandleTaskBooking(memberID int64, taskID int) {
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
