package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"
)

type AreaService struct {
	areaRepository *repository.AreaRepository
}

func NewAreaService(areaRepository *repository.AreaRepository) *AreaService {
	return &AreaService{
		areaRepository: areaRepository,
	}
}

func (s *AreaService) GetAllAreas() ([]model.Area, error) {
	return s.areaRepository.FindAll()
}
