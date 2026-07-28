package cleaning

import (
	"context"
	"errors"
	"testing"
	"time"

	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAssignTask_UsesAtomicRepositoryMethod(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	teamID := uuid.New()
	taskID := uuid.New()
	taskDefID := uuid.New()

	userRepo := new(MockUserRepo)
	dutyRepo := new(MockDutyRepo)
	dutyTaskRepo := new(MockDutyTaskRepo)
	teamRepo := new(MockTeamRepo)
	eventBus := new(MockEventBus)

	svc := &Service{
		userRepo:     userRepo,
		teamRepo:     teamRepo,
		dutyRepo:     dutyRepo,
		dutyTaskRepo: dutyTaskRepo,
		eventBus:     eventBus,
	}

	resident, err := user.NewManualUser("Иван", "Иванов", "ivan", "hash")
	require.NoError(t, err)
	resident.JoinTeam(teamID)

	currentDuty := duty.NewDuty(teamID, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), 1)
	currentDuty.AddTask(taskID, taskDefID)
	task := duty.RestoreDutyTask(duty.RestoreDutyTaskParams{
		ID: taskID, DutyID: currentDuty.ID(), TaskDefID: taskDefID,
	})

	userRepo.On("FindByID", ctx, userID).Return(resident, nil).Once()
	dutyTaskRepo.On("FindByID", ctx, taskID).Return(task, nil).Once()
	dutyRepo.On("FindByID", ctx, currentDuty.ID()).Return(currentDuty, nil).Once()
	dutyTaskRepo.On("Assign", ctx, taskID, resident.ID(), mock.AnythingOfType("time.Time")).Return(nil).Once()
	eventBus.On("Publish", ctx, events.TopicTaskAssigned, mock.AnythingOfType("events.TaskAssignedEvent")).Return(nil).Once()

	err = svc.AssignTask(ctx, taskID, userID)
	require.NoError(t, err)

	dutyRepo.AssertNotCalled(t, "CreateWithTasks", mock.Anything, mock.Anything, mock.Anything)
}

func TestAssignTask_DoesNotPublishEventOnAtomicConflict(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	teamID := uuid.New()
	taskID := uuid.New()
	taskDefID := uuid.New()

	userRepo := new(MockUserRepo)
	dutyRepo := new(MockDutyRepo)
	dutyTaskRepo := new(MockDutyTaskRepo)
	eventBus := new(MockEventBus)

	svc := &Service{
		userRepo:     userRepo,
		dutyRepo:     dutyRepo,
		dutyTaskRepo: dutyTaskRepo,
		eventBus:     eventBus,
	}

	resident, err := user.NewManualUser("Петр", "Петров", "petr", "hash")
	require.NoError(t, err)
	resident.JoinTeam(teamID)

	currentDuty := duty.NewDuty(teamID, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), 1)
	currentDuty.AddTask(taskID, taskDefID)
	task := duty.RestoreDutyTask(duty.RestoreDutyTaskParams{
		ID: taskID, DutyID: currentDuty.ID(), TaskDefID: taskDefID,
	})

	userRepo.On("FindByID", ctx, userID).Return(resident, nil).Once()
	dutyTaskRepo.On("FindByID", ctx, taskID).Return(task, nil).Once()
	dutyRepo.On("FindByID", ctx, currentDuty.ID()).Return(currentDuty, nil).Once()
	dutyTaskRepo.On("Assign", ctx, taskID, resident.ID(), mock.AnythingOfType("time.Time")).Return(duty.ErrTaskAssigned).Once()

	err = svc.AssignTask(ctx, taskID, userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, duty.ErrTaskAssigned)

	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
	dutyRepo.AssertNotCalled(t, "CreateWithTasks", mock.Anything, mock.Anything, mock.Anything)
}

func TestCompleteTask_UsesAtomicRepositoryMethod(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	teamID := uuid.New()
	taskID := uuid.New()
	taskDefID := uuid.New()

	userRepo := new(MockUserRepo)
	dutyRepo := new(MockDutyRepo)
	dutyTaskRepo := new(MockDutyTaskRepo)
	eventBus := new(MockEventBus)

	svc := &Service{
		userRepo:     userRepo,
		dutyRepo:     dutyRepo,
		dutyTaskRepo: dutyTaskRepo,
		eventBus:     eventBus,
	}

	resident, err := user.NewManualUser("Мария", "Иванова", "maria", "hash")
	require.NoError(t, err)
	resident.JoinTeam(teamID)

	currentDuty := duty.NewDuty(teamID, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), 1)
	currentDuty.AddTask(taskID, taskDefID)
	assigneeID := resident.ID()
	task := duty.RestoreDutyTask(duty.RestoreDutyTaskParams{
		ID: taskID, DutyID: currentDuty.ID(), TaskDefID: taskDefID, AssigneeID: &assigneeID,
	})

	userRepo.On("FindByID", ctx, userID).Return(resident, nil).Once()
	dutyTaskRepo.On("FindByID", ctx, taskID).Return(task, nil).Once()
	dutyRepo.On("FindByID", ctx, currentDuty.ID()).Return(currentDuty, nil).Twice()
	dutyTaskRepo.On("Complete", ctx, taskID, resident.ID(), mock.AnythingOfType("time.Time")).Return(nil).Once()
	eventBus.On("Publish", ctx, events.TopicTaskCompleted, mock.AnythingOfType("events.TaskCompletedEvent")).Return(nil).Once()

	err = svc.CompleteTask(ctx, taskID, userID)
	require.NoError(t, err)

	dutyRepo.AssertNotCalled(t, "CreateWithTasks", mock.Anything, mock.Anything, mock.Anything)
}

func TestCompleteTask_PublishesTasksReadyForReviewEvent(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	teamID := uuid.New()
	taskID := uuid.New()
	taskDefID := uuid.New()

	userRepo := new(MockUserRepo)
	dutyRepo := new(MockDutyRepo)
	dutyTaskRepo := new(MockDutyTaskRepo)
	teamRepo := new(MockTeamRepo)
	eventBus := new(MockEventBus)

	svc := &Service{
		userRepo:     userRepo,
		teamRepo:     teamRepo,
		dutyRepo:     dutyRepo,
		dutyTaskRepo: dutyTaskRepo,
		eventBus:     eventBus,
	}

	resident, err := user.NewManualUser("Анна", "Смирнова", "anna", "hash")
	require.NoError(t, err)
	resident.JoinTeam(teamID)

	currentDuty := duty.NewDuty(teamID, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), 1)
	currentDuty.AddTask(taskID, taskDefID)
	assigneeID := resident.ID()
	task := duty.RestoreDutyTask(duty.RestoreDutyTaskParams{
		ID: taskID, DutyID: currentDuty.ID(), TaskDefID: taskDefID, AssigneeID: &assigneeID,
	})

	completedAt := time.Now()
	refreshedDutyTask := duty.RestoreDutyTask(duty.RestoreDutyTaskParams{
		ID:             taskID,
		DutyID:         currentDuty.ID(),
		TaskDefID:      taskDefID,
		AssigneeID:     &assigneeID,
		CompletionDate: &completedAt,
	})
	refreshedDuty := duty.RestoreDuty(
		currentDuty.ID(),
		currentDuty.TeamID(),
		currentDuty.StartDate(),
		currentDuty.EndDate(),
		currentDuty.SequenceNumber(),
		[]*duty.DutyTask{refreshedDutyTask},
	)

	leaderID := uuid.New()
	team := structure.RestoreTeam(teamID, "Команда 1", uuid.New(), &leaderID, "blue", 1)

	userRepo.On("FindByID", ctx, userID).Return(resident, nil).Once()
	dutyTaskRepo.On("FindByID", ctx, taskID).Return(task, nil).Once()
	dutyRepo.On("FindByID", ctx, currentDuty.ID()).Return(currentDuty, nil).Once()
	dutyTaskRepo.On("Complete", ctx, taskID, resident.ID(), mock.AnythingOfType("time.Time")).Return(nil).Once()
	dutyRepo.On("FindByID", ctx, currentDuty.ID()).Return(refreshedDuty, nil).Once()
	teamRepo.On("FindByID", ctx, teamID).Return(team, nil).Once()
	eventBus.On("Publish", ctx, events.TopicTasksReadyForReview, mock.AnythingOfType("events.TasksReadyForReviewEvent")).Return(nil).Once()
	eventBus.On("Publish", ctx, events.TopicTaskCompleted, mock.AnythingOfType("events.TaskCompletedEvent")).Return(nil).Once()

	err = svc.CompleteTask(ctx, taskID, userID)
	require.NoError(t, err)
}

func TestAssignTask_WrapsRepositoryError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	teamID := uuid.New()
	taskID := uuid.New()
	taskDefID := uuid.New()
	repoErr := errors.New("db failure")

	userRepo := new(MockUserRepo)
	dutyRepo := new(MockDutyRepo)
	dutyTaskRepo := new(MockDutyTaskRepo)
	eventBus := new(MockEventBus)

	svc := &Service{
		userRepo:     userRepo,
		dutyRepo:     dutyRepo,
		dutyTaskRepo: dutyTaskRepo,
		eventBus:     eventBus,
	}

	resident, err := user.NewManualUser("Олег", "Сидоров", "oleg", "hash")
	require.NoError(t, err)
	resident.JoinTeam(teamID)

	currentDuty := duty.NewDuty(teamID, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), 1)
	currentDuty.AddTask(taskID, taskDefID)
	task := duty.RestoreDutyTask(duty.RestoreDutyTaskParams{
		ID: taskID, DutyID: currentDuty.ID(), TaskDefID: taskDefID,
	})

	userRepo.On("FindByID", ctx, userID).Return(resident, nil).Once()
	dutyTaskRepo.On("FindByID", ctx, taskID).Return(task, nil).Once()
	dutyRepo.On("FindByID", ctx, currentDuty.ID()).Return(currentDuty, nil).Once()
	dutyTaskRepo.On("Assign", ctx, taskID, resident.ID(), mock.AnythingOfType("time.Time")).Return(repoErr).Once()

	err = svc.AssignTask(ctx, taskID, userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)

	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}
