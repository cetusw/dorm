package service

import (
	"dorm/internal/dorm/infrastructure/mysql/repository"

	"github.com/google/uuid"
)

type SpecialTaskAssignmentService struct {
	repo *repository.SpecialTaskAssignmentRepository
}

func NewSpecialTaskAssignmentService(repo *repository.SpecialTaskAssignmentRepository) *SpecialTaskAssignmentService {
	return &SpecialTaskAssignmentService{repo: repo}
}

func (s *SpecialTaskAssignmentService) FindLatestByTaskID(taskID uuid.UUID) (*repository.SpecialTaskAssignmentRow, error) {
	return s.repo.FindLatestByTaskID(taskID)
}

func (s *SpecialTaskAssignmentService) StoreBatch(rows []repository.SpecialTaskAssignmentRow) error {
	return s.repo.StoreBatch(rows)
}
