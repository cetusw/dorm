package cleaning

import (
	"context"
	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
	queryports "dorm/pkg/core/ports/query"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	userRepo     user.Repository
	teamRepo     structure.TeamRepository
	groupRepo    structure.GroupRepository
	dutyRepo     duty.DutyRepository
	dutyTaskRepo duty.DutyTaskRepository
	taskRepo     catalog.TaskDefinitionRepository
	overrideRepo catalog.DutyTaskOverrideRepository
	areaRepo     catalog.AreaRepository
	eventBus     ports.EventBus
	now          func() time.Time
}

func NewCleaningService(
	userRepo user.Repository,
	teamRepo structure.TeamRepository,
	groupRepo structure.GroupRepository,
	dutyRepo duty.DutyRepository,
	dutyTaskRepo duty.DutyTaskRepository,
	taskRepo catalog.TaskDefinitionRepository,
	overrideRepo catalog.DutyTaskOverrideRepository,
	areaRepo catalog.AreaRepository,
	eventBus ports.EventBus,
) *Service {
	return &Service{
		userRepo:     userRepo,
		teamRepo:     teamRepo,
		groupRepo:    groupRepo,
		dutyRepo:     dutyRepo,
		dutyTaskRepo: dutyTaskRepo,
		taskRepo:     taskRepo,
		overrideRepo: overrideRepo,
		areaRepo:     areaRepo,
		eventBus:     eventBus,
		now:          time.Now,
	}
}

func (s *Service) AssignTask(ctx context.Context, taskID, userID uuid.UUID) error {
	data, err := s.loadTaskActionContext(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if err := s.canAssignTask(data); err != nil {
		return err
	}
	return s.applyTaskAssignment(ctx, data)
}

func (s *Service) CompleteTask(ctx context.Context, taskID, userID uuid.UUID) error {
	data, err := s.loadTaskActionContext(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if err := s.canCompleteTask(data); err != nil {
		return err
	}
	return s.applyTaskCompletion(ctx, data)
}

func (s *Service) UnassignTask(ctx context.Context, taskID, userID uuid.UUID) error {
	data, err := s.loadTaskActionContext(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if err := s.canUnassignTask(data); err != nil {
		return err
	}
	return s.applyTaskUnassignment(ctx, data)
}

func (s *Service) OpenTask(ctx context.Context, taskID, userID uuid.UUID) error {
	return s.CancelCompletion(ctx, taskID, userID)
}

func (s *Service) CancelCompletion(ctx context.Context, taskID, userID uuid.UUID) error {
	data, err := s.loadTaskActionContext(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if err := s.canCancelCompletion(data); err != nil {
		return err
	}
	return s.applyCompletionCancellation(ctx, data)
}

func (s *Service) VerifyTask(ctx context.Context, taskID, userID uuid.UUID) error {
	data, err := s.loadTaskActionContext(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if err := s.canVerifyTask(ctx, data); err != nil {
		return err
	}
	return s.applyTaskVerification(ctx, data)
}

func (s *Service) ReopenTask(ctx context.Context, taskID, userID uuid.UUID) error {
	data, err := s.loadTaskActionContext(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if err := s.canReopenTask(ctx, data); err != nil {
		return err
	}
	return s.applyTaskReopen(ctx, data)
}

type taskActionContext struct {
	user *user.User
	task *duty.DutyTask
	duty *duty.Duty
	now  time.Time
}

func (s *Service) loadTaskActionContext(
	ctx context.Context,
	taskID uuid.UUID,
	userID uuid.UUID,
) (*taskActionContext, error) {
	actor, err := s.loadTaskActor(ctx, userID)
	if err != nil {
		return nil, err
	}
	task, err := s.loadDutyTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	currentDuty, err := s.loadDuty(ctx, task.DutyID())
	if err != nil {
		return nil, err
	}
	return &taskActionContext{user: actor, task: task, duty: currentDuty, now: s.currentTime()}, nil
}

func (s *Service) loadTaskActor(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	actor, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}
	if actor == nil {
		return nil, fmt.Errorf("user not found")
	}
	return actor, nil
}

func (s *Service) loadDutyTask(ctx context.Context, taskID uuid.UUID) (*duty.DutyTask, error) {
	task, err := s.dutyTaskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("load duty task: %w", err)
	}
	return task, nil
}

func (s *Service) loadDuty(ctx context.Context, dutyID uuid.UUID) (*duty.Duty, error) {
	currentDuty, err := s.dutyRepo.FindByID(ctx, dutyID)
	if err != nil {
		return nil, fmt.Errorf("load duty: %w", err)
	}
	if currentDuty == nil {
		return nil, fmt.Errorf("duty not found")
	}
	return currentDuty, nil
}

func (s *Service) canAssignTask(data *taskActionContext) error {
	if err := s.requireActiveTeamDuty(data); err != nil {
		return err
	}
	return data.task.CanAssign(data.user.ID())
}

func (s *Service) canUnassignTask(data *taskActionContext) error {
	if err := s.requireActiveTeamDuty(data); err != nil {
		return err
	}
	return data.task.CanUnassign(data.user.ID())
}

func (s *Service) canCompleteTask(data *taskActionContext) error {
	if err := s.requireActiveTeamDuty(data); err != nil {
		return err
	}
	return data.task.CanComplete(data.user.ID())
}

func (s *Service) canCancelCompletion(data *taskActionContext) error {
	if err := s.requireActiveTeamDuty(data); err != nil {
		return err
	}
	return data.task.CanCancelCompletion(data.user.ID())
}

func (s *Service) canVerifyTask(ctx context.Context, data *taskActionContext) error {
	if err := s.requireTaskReviewer(ctx, data); err != nil {
		return err
	}
	return data.task.CanVerify(data.user.ID())
}

func (s *Service) canReopenTask(ctx context.Context, data *taskActionContext) error {
	if err := s.requireTaskReviewer(ctx, data); err != nil {
		return err
	}
	return data.task.CanReopen(data.user.ID())
}

func (s *Service) requireActiveTeamDuty(data *taskActionContext) error {
	if data.user.TeamID() == nil {
		return duty.ErrTaskAccessDenied
	}
	if !data.duty.BelongsToTeam(*data.user.TeamID()) {
		return duty.ErrTaskAccessDenied
	}
	if !data.duty.IsActiveAt(data.now) {
		return fmt.Errorf("duty is not active")
	}
	return nil
}

func (s *Service) requireTaskReviewer(ctx context.Context, data *taskActionContext) error {
	if err := s.requireActiveTeamDuty(data); err != nil {
		return err
	}
	team, err := s.teamRepo.FindByID(ctx, data.duty.TeamID())
	if err != nil {
		return fmt.Errorf("load duty team: %w", err)
	}
	if team == nil || team.LeaderID() == nil || *team.LeaderID() != data.user.ID() {
		return duty.ErrTaskAccessDenied
	}
	return nil
}

func (s *Service) applyTaskAssignment(ctx context.Context, data *taskActionContext) error {
	err := s.dutyTaskRepo.Assign(ctx, data.task.ID(), data.user.ID(), data.now)
	if err != nil {
		return fmt.Errorf("assign duty task: %w", err)
	}
	userID := data.user.ID()
	s.publishTaskAssigned(ctx, data.task.ID(), &userID)
	return nil
}

func (s *Service) applyTaskUnassignment(ctx context.Context, data *taskActionContext) error {
	err := s.dutyTaskRepo.Unassign(ctx, data.task.ID(), data.user.ID())
	if err != nil {
		return fmt.Errorf("unassign duty task: %w", err)
	}
	s.publishTaskAssigned(ctx, data.task.ID(), nil)
	return nil
}

func (s *Service) applyTaskCompletion(ctx context.Context, data *taskActionContext) error {
	err := s.dutyTaskRepo.Complete(ctx, data.task.ID(), data.user.ID(), data.now)
	if err != nil {
		return fmt.Errorf("complete duty task: %w", err)
	}
	s.publishTasksReadyForReviewIfNeeded(ctx, data)
	_ = s.eventBus.Publish(ctx, events.TopicTaskCompleted, events.TaskCompletedEvent{
		TaskID: data.task.ID(),
		UserID: data.user.ID(),
		Time:   data.now,
	})
	return nil
}

func (s *Service) applyCompletionCancellation(ctx context.Context, data *taskActionContext) error {
	err := s.dutyTaskRepo.CancelCompletion(ctx, data.task.ID(), data.user.ID())
	if err != nil {
		return fmt.Errorf("cancel duty task completion: %w", err)
	}
	s.publishTaskUncompleted(ctx, data)
	return nil
}

func (s *Service) applyTaskVerification(ctx context.Context, data *taskActionContext) error {
	err := s.dutyTaskRepo.Verify(ctx, data.task.ID(), data.user.ID(), data.now)
	if err != nil {
		return fmt.Errorf("verify duty task: %w", err)
	}
	return nil
}

func (s *Service) applyTaskReopen(ctx context.Context, data *taskActionContext) error {
	err := s.dutyTaskRepo.Reopen(ctx, data.task.ID(), data.user.ID())
	if err != nil {
		return fmt.Errorf("reopen duty task: %w", err)
	}
	s.publishTaskUncompleted(ctx, data)
	return nil
}

func (s *Service) publishTaskAssigned(ctx context.Context, taskID uuid.UUID, assigneeID *uuid.UUID) {
	_ = s.eventBus.Publish(ctx, events.TopicTaskAssigned, events.TaskAssignedEvent{
		TaskID:     taskID,
		AssigneeID: assigneeID,
	})
}

func (s *Service) publishTaskUncompleted(ctx context.Context, data *taskActionContext) {
	_ = s.eventBus.Publish(ctx, events.TopicTaskUncompleted, events.TaskUncompletedEvent{
		TaskID: data.task.ID(),
		UserID: data.user.ID(),
		Time:   data.now,
	})
}

func (s *Service) currentTime() time.Time {
	if s.now == nil {
		return time.Now()
	}
	return s.now()
}

func (s *Service) publishTasksReadyForReviewIfNeeded(ctx context.Context, data *taskActionContext) {
	currentDuty, err := s.dutyRepo.FindByID(ctx, data.duty.ID())
	if err != nil {
		log.Printf("load duty for tasks ready event: %v", err)
		return
	}
	if currentDuty == nil {
		return
	}

	progress := calculateDutyTaskProgress(currentDuty.Tasks())
	if !progress.IsReadyForReview() {
		return
	}

	team, err := s.teamRepo.FindByID(ctx, currentDuty.TeamID())
	if err != nil {
		log.Printf("load team for tasks ready event: %v", err)
		return
	}
	if team == nil || team.LeaderID() == nil {
		log.Printf("skip tasks ready notification: team %s has no leader", currentDuty.TeamID())
		return
	}

	_ = s.eventBus.Publish(ctx, events.TopicTasksReadyForReview, events.TasksReadyForReviewEvent{
		DutyID:     currentDuty.ID(),
		TeamID:     currentDuty.TeamID(),
		TeamHeadID: *team.LeaderID(),
		OccurredAt: data.now,
	})
}

func calculateDutyTaskProgress(tasks []*duty.DutyTask) queryports.DutyTaskProgress {
	progress := queryports.DutyTaskProgress{
		TotalCount: len(tasks),
	}

	for _, task := range tasks {
		if task.CompletionDate() != nil {
			progress.CompletedCount++
		}
		if task.VerificationDate() != nil {
			progress.VerifiedCount++
		}
	}

	return progress
}
