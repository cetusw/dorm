package dormitory

import (
	"context"

	"dorm/pkg/core/domain/structure"
)

type Service struct {
	dormitoryRepo structure.DormitoryRepository
}

func NewDormitoryService(
	dormitoryRepo structure.DormitoryRepository,
) *Service {
	return &Service{
		dormitoryRepo: dormitoryRepo,
	}
}

func (s *Service) GetDormitories(ctx context.Context) ([]*structure.Dormitory, error) {
	return s.dormitoryRepo.FindAll(ctx)
}
