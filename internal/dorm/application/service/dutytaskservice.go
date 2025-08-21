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

func (s *DutyTaskService) GetDutyTasks(dutyID uuid.UUID) ([]model.DutyTaskReadable, error) {
	rows, err := s.dutyTaskRepository.GetDutyTaskRows(dutyID)
	if err != nil {
		return nil, err
	}

	var dutyTasksReadable []model.DutyTaskReadable

	for rows.Next() {
		var dutyTaskReadable model.DutyTaskReadable
		if err := rows.Scan(
			&dutyTaskReadable.AreaFloor,
			&dutyTaskReadable.AreaName,
			&dutyTaskReadable.TaskTitle,
			&dutyTaskReadable.TaskCost,
			&dutyTaskReadable.AssigneeFirstName,
			&dutyTaskReadable.AssigneeLastName,
			&dutyTaskReadable.CompletionDate,
			&dutyTaskReadable.VerificationDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tasksReadable row: %w", err)
		}

		dutyTasksReadable = append(dutyTasksReadable, dutyTaskReadable)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team rows: %w", err)
	}

	return dutyTasksReadable, nil
}
