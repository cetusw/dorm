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
	"sort"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	userRepo        user.Repository
	teamRepo        structure.TeamRepository
	groupRepo       structure.GroupRepository
	dutyRepo        duty.DutyRepository
	dutyTaskRepo    duty.DutyTaskRepository
	participantRepo duty.ParticipantRepository
	taskRepo        catalog.TaskDefinitionRepository
	overrideRepo    catalog.DutyTaskOverrideRepository
	areaRepo        catalog.AreaRepository
	eventBus        ports.EventBus
	now             func() time.Time
	location        *time.Location
}

// SetParticipantRepository enables duty-specific membership checks. It is kept
// separate from construction to preserve the existing scheduler/test wiring.
func (s *Service) SetParticipantRepository(repo duty.ParticipantRepository) { s.participantRepo = repo }

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
		location:     time.Local,
	}
}

func (s *Service) SetLocation(location *time.Location) {
	if location != nil {
		s.location = location
	}
}

func (s *Service) AssignTask(ctx context.Context, taskID, userID uuid.UUID) error {
	data, err := s.loadTaskActionContext(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if err := s.canAssignTask(ctx, data); err != nil {
		return err
	}
	return s.applyTaskAssignment(ctx, data)
}

func (s *Service) CompleteTask(ctx context.Context, taskID, userID uuid.UUID) error {
	data, err := s.loadTaskActionContext(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if err := s.canCompleteTask(ctx, data); err != nil {
		return err
	}
	return s.applyTaskCompletion(ctx, data)
}

func (s *Service) UnassignTask(ctx context.Context, taskID, userID uuid.UUID) error {
	data, err := s.loadTaskActionContext(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if err := s.canUnassignTask(ctx, data); err != nil {
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
	if err := s.canCancelCompletion(ctx, data); err != nil {
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

func (s *Service) canAssignTask(ctx context.Context, data *taskActionContext) error {
	if err := s.requireActionableTeamDuty(ctx, data); err != nil {
		return err
	}
	return data.task.CanAssign(data.user.ID())
}

func (s *Service) canUnassignTask(ctx context.Context, data *taskActionContext) error {
	if err := s.requireActionableTeamDuty(ctx, data); err != nil {
		return err
	}
	return data.task.CanUnassign(data.user.ID())
}

func (s *Service) canCompleteTask(ctx context.Context, data *taskActionContext) error {
	if err := s.requireActionableTeamDuty(ctx, data); err != nil {
		return err
	}
	return data.task.CanComplete(data.user.ID())
}

func (s *Service) canCancelCompletion(ctx context.Context, data *taskActionContext) error {
	if err := s.requireActionableTeamDuty(ctx, data); err != nil {
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

func (s *Service) requireActionableTeamDuty(ctx context.Context, data *taskActionContext) error {
	if s.participantRepo != nil {
		active, err := s.participantRepo.IsActive(ctx, data.duty.ID(), data.user.ID())
		if err != nil {
			return fmt.Errorf("check duty participant: %w", err)
		}
		if !active {
			return duty.ErrTaskAccessDenied
		}
	} else {
		if data.user.TeamID() == nil {
			return duty.ErrTaskAccessDenied
		}
		if !data.duty.BelongsToTeam(*data.user.TeamID()) {
			return duty.ErrTaskAccessDenied
		}
	}

	at := s.currentTime()
	if data.duty.IsActiveAt(at) {
		return nil
	}
	if data.duty.Start().After(at) {
		return duty.ErrDutyActionsUnavailable
	}
	if !dutyHasOutstandingTasks(data.duty) {
		return duty.ErrDutyActionsUnavailable
	}

	return nil
}

func (s *Service) requireTaskReviewer(ctx context.Context, data *taskActionContext) error {
	if err := s.requireActionableTeamDuty(ctx, data); err != nil {
		return err
	}
	if data.duty.LeaderID() == nil || *data.duty.LeaderID() != data.user.ID() {
		return duty.ErrTaskAccessDenied
	}
	return nil
}

func dutyHasOutstandingTasks(currentDuty *duty.Duty) bool {
	if currentDuty == nil {
		return false
	}

	for _, task := range currentDuty.Tasks() {
		if task.VerificationDate() == nil {
			return true
		}
	}

	return false
}

func (s *Service) currentTime() time.Time {
	if s.now == nil {
		s.now = time.Now
	}
	now := s.now()
	if s.location == nil {
		return now
	}
	return now.In(s.location)
}

func (s *Service) applyTaskAssignment(ctx context.Context, data *taskActionContext) error {
	err := s.dutyTaskRepo.Assign(ctx, data.task.ID(), data.user.ID(), data.now)
	if err != nil {
		return fmt.Errorf("assign duty task: %w", err)
	}
	userID := data.user.ID()
	s.publishTaskAssigned(ctx, data.task.ID(), &userID)
	if err := s.assignRemainingFreeTasksToLastUnderGoalMember(ctx, data); err != nil {
		return err
	}
	return nil
}

func (s *Service) assignRemainingFreeTasksToLastUnderGoalMember(ctx context.Context, data *taskActionContext) error {
	if !hasOtherFreeTasks(data.duty.Tasks(), data.task.ID()) {
		return nil
	}

	refreshedDuty, err := s.loadDuty(ctx, data.task.DutyID())
	if err != nil {
		return fmt.Errorf("reload duty after assignment: %w", err)
	}

	teamMembers, err := s.dutyMembers(ctx, refreshedDuty)
	if err != nil {
		return err
	}
	if len(teamMembers) == 0 {
		return nil
	}

	targetCost, err := s.calculateCostPerResidentGoal(ctx, refreshedDuty.Tasks(), len(teamMembers))
	if err != nil {
		return fmt.Errorf("calculate duty cost goal: %w", err)
	}
	if targetCost == 0 {
		return nil
	}

	memberTakenCost := make(map[uuid.UUID]int, len(teamMembers))
	freeTasks := make([]*duty.DutyTask, 0)
	for _, dutyTask := range refreshedDuty.Tasks() {
		if dutyTask.AssigneeID() == nil {
			freeTasks = append(freeTasks, dutyTask)
			continue
		}

		taskDefinition, err := s.taskRepo.FindByID(ctx, dutyTask.TaskDefID())
		if err != nil {
			return fmt.Errorf("load task definition %s: %w", dutyTask.TaskDefID(), err)
		}
		if taskDefinition == nil {
			return fmt.Errorf("task definition %s not found", dutyTask.TaskDefID())
		}

		memberTakenCost[*dutyTask.AssigneeID()] += taskDefinition.Cost()
	}

	if len(freeTasks) == 0 {
		return nil
	}

	var membersBelowGoal []*user.User
	for _, teamMember := range teamMembers {
		if memberTakenCost[teamMember.ID()] < targetCost {
			membersBelowGoal = append(membersBelowGoal, teamMember)
		}
	}

	if len(membersBelowGoal) != 1 {
		return nil
	}

	lastMember := membersBelowGoal[0]
	sort.SliceStable(freeTasks, func(i, j int) bool {
		return freeTasks[i].ID().String() < freeTasks[j].ID().String()
	})

	for _, freeTask := range freeTasks {
		err := s.dutyTaskRepo.Assign(ctx, freeTask.ID(), lastMember.ID(), data.now)
		if err != nil {
			if err == duty.ErrTaskAssigned {
				continue
			}
			return fmt.Errorf("auto-assign remaining duty task %s: %w", freeTask.ID(), err)
		}

		lastMemberID := lastMember.ID()
		s.publishTaskAssigned(ctx, freeTask.ID(), &lastMemberID)
	}

	return nil
}

func (s *Service) dutyMembers(ctx context.Context, currentDuty *duty.Duty) ([]*user.User, error) {
	if s.participantRepo == nil {
		return s.userRepo.FindByTeamID(ctx, currentDuty.TeamID())
	}
	participants, err := s.participantRepo.List(ctx, currentDuty.ID(), false)
	if err != nil {
		return nil, fmt.Errorf("load duty participants: %w", err)
	}
	members := make([]*user.User, 0, len(participants))
	for _, participant := range participants {
		member, err := s.userRepo.FindByID(ctx, participant.ParticipantID)
		if err != nil {
			return nil, fmt.Errorf("load duty participant: %w", err)
		}
		if member != nil {
			members = append(members, member)
		}
	}
	return members, nil
}

func hasOtherFreeTasks(tasks []*duty.DutyTask, currentTaskID uuid.UUID) bool {
	for _, dutyTask := range tasks {
		if dutyTask.ID() == currentTaskID {
			continue
		}
		if dutyTask.AssigneeID() == nil {
			return true
		}
	}

	return false
}

func (s *Service) calculateCostPerResidentGoal(
	ctx context.Context,
	tasks []*duty.DutyTask,
	residentCount int,
) (int, error) {
	if len(tasks) == 0 || residentCount == 0 {
		return 0, nil
	}

	totalCost := 0
	for _, dutyTask := range tasks {
		taskDefinition, err := s.taskRepo.FindByID(ctx, dutyTask.TaskDefID())
		if err != nil {
			return 0, fmt.Errorf("load task definition %s: %w", dutyTask.TaskDefID(), err)
		}
		if taskDefinition == nil {
			return 0, fmt.Errorf("task definition %s not found", dutyTask.TaskDefID())
		}

		totalCost += taskDefinition.Cost()
	}

	return (totalCost + residentCount - 1) / residentCount, nil
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

	if currentDuty.LeaderID() == nil {
		log.Printf("skip tasks ready notification: duty %s has no leader", currentDuty.ID())
		return
	}

	_ = s.eventBus.Publish(ctx, events.TopicTasksReadyForReview, events.TasksReadyForReviewEvent{
		DutyID:     currentDuty.ID(),
		TeamID:     currentDuty.TeamID(),
		TeamHeadID: *currentDuty.LeaderID(),
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
