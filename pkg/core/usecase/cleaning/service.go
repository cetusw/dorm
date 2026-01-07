package cleaning

import (
	"context"
	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	userRepo  user.Repository
	teamRepo  structure.TeamRepository
	groupRepo structure.GroupRepository
	dutyRepo  duty.Repository
	taskRepo  catalog.TaskDefinitionRepository
	areaRepo  catalog.AreaRepository
	eventBus  ports.EventBus
}

func NewCleaningService(
	userRepo user.Repository,
	teamRepo structure.TeamRepository,
	groupRepo structure.GroupRepository,
	dutyRepo duty.Repository,
	taskRepo catalog.TaskDefinitionRepository,
	areaRepo catalog.AreaRepository,
	eventBus ports.EventBus,
) *Service {
	return &Service{
		userRepo:  userRepo,
		teamRepo:  teamRepo,
		groupRepo: groupRepo,
		dutyRepo:  dutyRepo,
		taskRepo:  taskRepo,
		areaRepo:  areaRepo,
		eventBus:  eventBus,
	}
}

func (s *Service) AssignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if u.TeamID() == nil {
		return fmt.Errorf("user is not in a team")
	}

	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil {
		return fmt.Errorf("failed to find active duty: %w", err)
	}
	if d == nil {
		return fmt.Errorf("no active duty for this team")
	}

	if err := d.AssignTask(taskID, u.ID()); err != nil {
		return err
	}

	if err := s.dutyRepo.Save(ctx, d); err != nil {
		return fmt.Errorf("failed to save duty: %w", err)
	}

	_ = s.eventBus.Publish(ctx, events.TopicTaskAssigned, events.TaskAssignedEvent{
		TaskID:     taskID,
		AssigneeID: &userID,
	})

	return nil
}

func (s *Service) CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.TeamID() == nil {
		return fmt.Errorf("user not in team")
	}

	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil || d == nil {
		return fmt.Errorf("duty not found")
	}

	if err := d.CompleteTask(taskID); err != nil {
		return err
	}

	if err := s.dutyRepo.Save(ctx, d); err != nil {
		return err
	}

	_ = s.eventBus.Publish(ctx, events.TopicTaskCompleted, events.TaskCompletedEvent{
		TaskID: taskID,
		UserID: userID,
		Time:   time.Now(),
	})

	return nil
}

func (s *Service) UnassignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if u.TeamID() == nil {
		return fmt.Errorf("user is not in a team")
	}

	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil {
		return fmt.Errorf("failed to find active duty: %w", err)
	}
	if d == nil {
		return fmt.Errorf("no active duty for this team")
	}

	if err := d.UnassignTask(taskID); err != nil {
		return err
	}

	if err := s.dutyRepo.Save(ctx, d); err != nil {
		return fmt.Errorf("failed to save duty: %w", err)
	}

	_ = s.eventBus.Publish(ctx, events.TopicTaskAssigned, events.TaskAssignedEvent{
		TaskID:     taskID,
		AssigneeID: nil,
	})
	return nil
}

func (s *Service) OpenTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.TeamID() == nil {
		return fmt.Errorf("user not in team")
	}

	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil || d == nil {
		return fmt.Errorf("duty not found")
	}

	if err := d.OpenTask(taskID); err != nil {
		return err
	}

	if err := s.dutyRepo.Save(ctx, d); err != nil {
		return err
	}

	_ = s.eventBus.Publish(ctx, events.TopicTaskUncompleted, events.TaskUncompletedEvent{
		TaskID: taskID,
		UserID: userID,
		Time:   time.Now(),
	})

	return nil
}

func (s *Service) GetTeamActiveDuty(ctx context.Context, teamID uuid.UUID) (*duty.Duty, error) {
	return s.dutyRepo.FindCurrentByTeamID(ctx, teamID)
}
