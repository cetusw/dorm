package service

import (
	"dorm/pkg/dorm/application/model"
	"dorm/pkg/dorm/infrastructure/mysql/repository"
	"errors"
	"time"

	"github.com/google/uuid"
)

type DutyService struct {
	dutyRepository *repository.DutyRepository
}

func NewDutyService(repository *repository.DutyRepository) *DutyService {
	return &DutyService{
		dutyRepository: repository,
	}
}

func (s *DutyService) CreateNewDuties(teams []model.Team, start time.Time, end time.Time) ([]model.Duty, error) {
	var duties []model.Duty
	for _, team := range teams {
		duty := &model.Duty{
			DutyID: uuid.New(),
			TeamID: team.TeamID,
			Start:  start,
			End:    end,
		}
		duties = append(duties, *duty)
	}

	return duties, s.dutyRepository.StoreBatch(duties)
}

func (s *DutyService) GetCurrentWeek() (int, error) {
	return s.dutyRepository.CountDistinctStartDates()
}

func (s *DutyService) GetGroupLastDuty(groupID uuid.UUID) (*model.Duty, error) {
	lastDuty, err := s.dutyRepository.FindLastDutyByGroupID(groupID)
	if err != nil {
		return nil, err
	}
	if lastDuty == nil {
		return nil, nil
	}
	return lastDuty, nil
}

func (s *DutyService) GetUserLastDuty(userID uuid.UUID) (*model.Duty, error) {
	lastDuty, err := s.dutyRepository.FindLastDutyByUserID(userID)
	if err != nil {
		return nil, err
	}
	if lastDuty == nil {
		return nil, errors.New("last user duty not found")
	}
	return lastDuty, nil
}
