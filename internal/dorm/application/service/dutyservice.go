package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"
	"time"

	"github.com/google/uuid"
)

type DutyService struct {
	dutyRepository *repository.DutyRepository
}

func NewDutyService(dutyRepository *repository.DutyRepository) *DutyService {
	return &DutyService{
		dutyRepository: dutyRepository,
	}
}

func (s *DutyService) CreateNewDuty(teamID int, start time.Time, end time.Time) (uuid.UUID, error) {
	newDutyID := uuid.New()
	duty := &model.Duty{
		DutyID: newDutyID,
		TeamID: teamID,
		Start:  start,
		End:    end,
	}

	return newDutyID, s.dutyRepository.Store(duty)
}

func (s *DutyService) GetLastDutyTeamID() (int, error) {
	lastDuty, err := s.dutyRepository.FindLast()
	if err != nil {
		return 0, err
	}
	if lastDuty == nil {
		return 0, nil
	}
	return lastDuty.TeamID, nil
}

func (s *DutyService) GetLastDuty() (*model.Duty, error) {
	return s.dutyRepository.FindLast()
}
