package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"

	"github.com/google/uuid"
)

type TaskService struct {
	taskRepository *repository.TaskRepository
}

func NewTaskService(repository *repository.TaskRepository) *TaskService {
	return &TaskService{
		taskRepository: repository,
	}
}

func (s *TaskService) GetPublicTasks() ([]model.Task, error) {
	return s.taskRepository.FindPublicTasks()
}

func (s *TaskService) GetGroupTasks(groupID uuid.UUID) ([]model.Task, error) {
	return s.taskRepository.FindGroupTasks(groupID)
}
