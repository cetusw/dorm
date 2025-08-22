package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"

	"github.com/google/uuid"
)

type TaskService struct {
	taskRepository *repository.TaskRepository
}

func NewTaskService(taskRepository *repository.TaskRepository) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
	}
}

func (s *TaskService) GetTasksByAreaID(areaID int) ([]model.Task, error) {
	return s.taskRepository.FindTasksByAreaID(areaID)
}

func (s *TaskService) GetTaskTitleByTaskID(taskID uuid.UUID) (*model.Task, error) {
	return s.taskRepository.Find(taskID)
}
