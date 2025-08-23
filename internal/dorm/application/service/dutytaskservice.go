package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"
	"fmt"

	"github.com/google/uuid"
)

type DutyTaskService struct {
	dutyTaskRepository *repository.DutyTaskRepository
	taskRepository     *repository.TaskRepository
	dutyRepository     *repository.DutyRepository
	teamRepository     *repository.TeamRepository
}

func NewDutyTaskService(
	dutyTaskRepository *repository.DutyTaskRepository,
	taskRepository *repository.TaskRepository,
	dutyRepository *repository.DutyRepository,
	teamRepository *repository.TeamRepository,
) *DutyTaskService {
	return &DutyTaskService{
		dutyTaskRepository: dutyTaskRepository,
		taskRepository:     taskRepository,
		dutyRepository:     dutyRepository,
		teamRepository:     teamRepository,
	}
}

func (s *DutyTaskService) SetDutyTasks(dutyID uuid.UUID) error {
	duty, err := s.dutyRepository.Find(dutyID)
	if err != nil {
		return err
	}
	if duty == nil {
		return fmt.Errorf("duty ID not found: %s", dutyID)
	}

	team, err := s.teamRepository.Find(duty.TeamID)
	if err != nil {
		return err
	}
	reviewerID := team.TeamLeaderID

	taskIDs, err := s.taskRepository.FindAllIDs()
	if err != nil {
		return err
	}
	if len(taskIDs) == 0 {
		return nil
	}

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

	err = s.dutyTaskRepository.StoreBatch(dutyTasksToCreate)
	if err != nil {
		return err
	}

	return nil
}

func (s *DutyTaskService) GetDutyTasksReadable(dutyID uuid.UUID) ([]model.DutyTaskReadable, error) {
	return s.dutyTaskRepository.FindDutyTasksReadableByDutyID(dutyID)
}

func (s *DutyTaskService) GetUncompletedDutyTasksReadableByAssigneeIDAndDutyID(assigneeID uuid.UUID, dutyID uuid.UUID) ([]model.DutyTaskReadable, error) {
	return s.dutyTaskRepository.FindUncompletedDutyTasksReadableByAssigneeIDAndDutyID(assigneeID, dutyID)
}

func (s *DutyTaskService) GetUnassignedDutyTasksReadableByAreaIDAndDutyID(areaID int, dutyID uuid.UUID) ([]model.DutyTaskReadable, error) {
	return s.dutyTaskRepository.FindUnassignedDutyTasksReadableByAreaIDAndDutyID(areaID, dutyID)
}

func (s *DutyTaskService) GetUserPointsByDutyID(assigneeID uuid.UUID, dutyID uuid.UUID) (int, error) {
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

func (s *DutyTaskService) GetAllPointsByDutyID(dutyID uuid.UUID) (int, error) {
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

func (s *DutyTaskService) SetDutyTaskAssigneeIDByDutyID(assigneeID *uuid.UUID, taskID uuid.UUID, dutyID uuid.UUID) error {
	err := s.dutyTaskRepository.UpdateCurrentDutyTaskAssigneeIDByTaskIDAndDutyID(assigneeID, taskID, dutyID)
	if err != nil {
		return err
	}

	return nil
}

func (s *DutyTaskService) CompleteDutyTaskByDutyIDAndTaskID(dutyID uuid.UUID, taskID uuid.UUID) error {
	err := s.dutyTaskRepository.UpdateDutyTaskCompletionDateByDutyIDAndTaskID(dutyID, taskID)
	if err != nil {
		return err
	}

	return nil
}
