package service

import (
	"dorm/internal/dorm/application/model"
	"fmt"
	"time"
)

type CleaningService struct {
	sheetsService *SheetsService
	dutyService   *DutyService
	teamService   *TeamService
}

func NewCleaningService(
	sheetsService *SheetsService,
	dutyService *DutyService,
	teamService *TeamService,
) *CleaningService {
	return &CleaningService{
		sheetsService: sheetsService,
		dutyService:   dutyService,
		teamService:   teamService,
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
	err = s.dutyService.CreateNewDuty(newDutyTeamID, startTime, endTime)
	if err != nil {
		return err
	}
	var teamColor int
	teamColor, err = s.teamService.GetTeamColor(newDutyTeamID)
	if err != nil {
		return err
	}
	var tasks []model.Task
	err = s.sheetsService.CreateWeeklySheet(sheetTitle, teamColor)
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
