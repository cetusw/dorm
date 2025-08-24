package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"

	"github.com/google/uuid"
)

type DutyTaskService struct {
	dutyTaskRepository *repository.DutyTaskRepository
}

func NewDutyTaskService(repository *repository.DutyTaskRepository) *DutyTaskService {
	return &DutyTaskService{
		dutyTaskRepository: repository,
	}
}

func (s *DutyTaskService) SetDutyTasks(dutyID uuid.UUID, taskIDs []uuid.UUID, reviewerID uuid.UUID) error {
	var dutyTasksToCreate []model.DutyTask
	for _, taskID := range taskIDs {
		dutyTask := model.DutyTask{
			DutyTaskID: uuid.New(),
			DutyID:     dutyID,
			TaskID:     taskID,
			ReviewerID: reviewerID,
		}
		dutyTasksToCreate = append(dutyTasksToCreate, dutyTask)
	}

	err := s.dutyTaskRepository.StoreBatch(dutyTasksToCreate)
	if err != nil {
		return err
	}

	return nil
}

func (s *DutyTaskService) GetDutyTasksReadable(dutyID uuid.UUID) ([]model.DutyTaskView, error) {
	return s.dutyTaskRepository.FindDutyTasksReadableByDutyID(dutyID)
}

func (s *DutyTaskService) GetUncompletedDutyTasksView(
	assigneeID uuid.UUID,
	dutyID uuid.UUID,
) ([]model.DutyTaskView, error) {
	return s.dutyTaskRepository.FindUncompletedDutyTasksView(assigneeID, dutyID)
}

func (s *DutyTaskService) GetUnassignedDutyTasksView(
	areaID int,
	dutyID uuid.UUID,
) ([]model.DutyTaskView, error) {
	return s.dutyTaskRepository.FindUnassignedDutyTasksView(areaID, dutyID)
}

func (s *DutyTaskService) GetUserPoints(assigneeID uuid.UUID, dutyID uuid.UUID) (int, error) {
	dutyTasks, err := s.dutyTaskRepository.FindDutyTasksReadableByDutyID(dutyID)
	if err != nil {
		return 0, err
	}
	sum := 0
	for _, task := range dutyTasks {
		if task.AssigneeID != nil && *task.AssigneeID == assigneeID {
			sum += task.TaskCost
		}
	}

	return sum, nil
}

func (s *DutyTaskService) GetUserConfirmedPoints(assigneeID uuid.UUID, dutyID uuid.UUID) (int, error) {
	dutyTasks, err := s.dutyTaskRepository.FindDutyTasksReadableByDutyID(dutyID)
	if err != nil {
		return 0, err
	}
	sum := 0
	for _, task := range dutyTasks {
		if task.AssigneeID != nil && *task.AssigneeID == assigneeID && task.CompletionDate != nil {
			sum += task.TaskCost
		}
	}

	return sum, nil
}

func (s *DutyTaskService) GetDutyPoints(dutyID uuid.UUID) (int, error) {
	dutyTasks, err := s.dutyTaskRepository.FindDutyTasksReadableByDutyID(dutyID)
	if err != nil {
		return 0, err
	}
	sum := 0
	for _, task := range dutyTasks {
		sum += task.TaskCost
	}

	return sum, nil
}

func (s *DutyTaskService) SetDutyTaskAssigneeID(
	assigneeID *uuid.UUID,
	taskID uuid.UUID,
	dutyID uuid.UUID,
) error {
	err := s.dutyTaskRepository.UpdateDutyTaskAssigneeID(assigneeID, dutyID, taskID)
	if err != nil {
		return err
	}

	return nil
}

func (s *DutyTaskService) CompleteDutyTaskByDutyIDAndTaskID(dutyID uuid.UUID, taskID uuid.UUID) error {
	err := s.dutyTaskRepository.UpdateDutyTaskCompletionDate(dutyID, taskID)
	if err != nil {
		return err
	}

	return nil
}
