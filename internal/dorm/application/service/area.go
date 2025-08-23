package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"

	"github.com/google/uuid"
)

func NewAreaService(repository *repository.AreaRepository) *AreaService {
	return &AreaService{
		areaRepository: repository,
	}
}

type AreaService struct {
	areaRepository *repository.AreaRepository
}

func (s *AreaService) GetAllAreas() ([]model.Area, error) {
	return s.areaRepository.FindAll()
}

func (s *AreaService) GetUnassignedAreasByDutyID(dutyID uuid.UUID) ([]model.Area, error) {
	return s.areaRepository.FindUnassignedAreasByDutyID(dutyID)
}
