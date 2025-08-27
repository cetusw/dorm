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

func (s *TaskService) GetTasksByAreaID(areaID int) ([]model.Task, error) {
	return s.taskRepository.FindTasksByAreaID(areaID)
}

func (s *TaskService) GetTaskTitleByTaskID(taskID uuid.UUID) (*model.Task, error) {
	return s.taskRepository.Find(taskID)
}

func (s *TaskService) GetAllTaskIDsByFrequency(freq int) ([]uuid.UUID, error) {
	tasks, err := s.taskRepository.FindAllByFrequency(freq)
	if err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, t.TaskID)
	}
	return out, nil
}
