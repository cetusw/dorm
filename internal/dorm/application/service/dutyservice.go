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

func (s *DutyService) CreateNewDuty(teamID int, start time.Time, end time.Time) error {
	duty := &model.Duty{
		DutyID: uuid.New(),
		TeamID: teamID,
		Start:  start,
		End:    end,
	}

	return s.dutyRepository.Store(duty)
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
