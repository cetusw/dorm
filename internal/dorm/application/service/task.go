package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"
)

type TaskService struct {
	taskRepository *repository.TaskRepository
}

func NewTaskService(repository *repository.TaskRepository) *TaskService {
	return &TaskService{
		taskRepository: repository,
	}
}

func (s *TaskService) GetTasksByScope(isPublic bool) ([]model.Task, error) {
	return s.taskRepository.FindTasksByScope(isPublic)
}
