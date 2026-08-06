package cleaning

import (
	"context"
	"errors"
	"testing"
	"time"

	"dorm/pkg/core/domain/catalog"
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

	resident, err := user.NewUser("Иван", "Иванов", "ivan", "hash")
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

	resident, err := user.NewUser("Петр", "Петров", "petr", "hash")
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

	resident, err := user.NewUser("Мария", "Иванова", "maria", "hash")
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

	resident, err := user.NewUser("Анна", "Смирнова", "anna", "hash")
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

	resident, err := user.NewUser("Олег", "Сидоров", "oleg", "hash")
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

func TestAssignTask_AutoAssignsRemainingFreeTasksToLastUnderGoalMember(t *testing.T) {
	ctx := context.Background()
	teamID := uuid.New()
	userAID := uuid.New()
	userBID := uuid.New()

	takenTaskID := uuid.New()
	takenTaskDefID := uuid.New()
	freeTaskID := uuid.New()
	freeTaskDefID := uuid.New()

	userRepo := new(MockUserRepo)
	dutyRepo := new(MockDutyRepo)
	dutyTaskRepo := new(MockDutyTaskRepo)
	taskRepo := new(MockCatalogRepo)
	eventBus := new(MockEventBus)

	svc := &Service{
		userRepo:     userRepo,
		dutyRepo:     dutyRepo,
		dutyTaskRepo: dutyTaskRepo,
		taskRepo:     taskRepo,
		eventBus:     eventBus,
	}

	userA, err := user.NewUser("Иван", "Иванов", "ivan", "hash")
	require.NoError(t, err)
	userA.JoinTeam(teamID)
	userA = user.RestoreUser(userAID, userA.Login(), userA.PasswordHash(), userA.FirstName(), userA.MiddleName(), userA.LastName(), userA.TeamID(), userA.RoomNumber(), userA.FloorNumber(), userA.DormitoryID(), userA.CreatedAt())

	userB, err := user.NewUser("Петр", "Петров", "petr", "hash")
	require.NoError(t, err)
	userB.JoinTeam(teamID)
	userB = user.RestoreUser(userBID, userB.Login(), userB.PasswordHash(), userB.FirstName(), userB.MiddleName(), userB.LastName(), userB.TeamID(), userB.RoomNumber(), userB.FloorNumber(), userB.DormitoryID(), userB.CreatedAt())

	currentDuty := duty.NewDuty(teamID, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), 1)
	currentDuty.AddTask(takenTaskID, takenTaskDefID)
	currentDuty.AddTask(freeTaskID, freeTaskDefID)

	taskToTake := duty.RestoreDutyTask(duty.RestoreDutyTaskParams{
		ID: takenTaskID, DutyID: currentDuty.ID(), TaskDefID: takenTaskDefID,
	})

	assignedToA := userA.ID()
	refreshedTakenTask := duty.RestoreDutyTask(duty.RestoreDutyTaskParams{
		ID:         takenTaskID,
		DutyID:     currentDuty.ID(),
		TaskDefID:  takenTaskDefID,
		AssigneeID: &assignedToA,
	})
	refreshedFreeTask := duty.RestoreDutyTask(duty.RestoreDutyTaskParams{
		ID: freeTaskID, DutyID: currentDuty.ID(), TaskDefID: freeTaskDefID,
	})
	refreshedDuty := duty.RestoreDuty(
		currentDuty.ID(),
		currentDuty.TeamID(),
		currentDuty.StartDate(),
		currentDuty.EndDate(),
		currentDuty.SequenceNumber(),
		[]*duty.DutyTask{refreshedTakenTask, refreshedFreeTask},
	)

	taskDef40, err := catalog.NewTaskDefinition(1, "Задача 40", 40, 1, 1)
	require.NoError(t, err)
	taskDef40 = catalog.RestoreTaskDefinition(takenTaskDefID, taskDef40.AreaID(), taskDef40.Title(), taskDef40.Cost(), taskDef40.RecurrenceInterval(), taskDef40.StartSequence())

	taskDef20, err := catalog.NewTaskDefinition(1, "Задача 20", 20, 1, 1)
	require.NoError(t, err)
	taskDef20 = catalog.RestoreTaskDefinition(freeTaskDefID, taskDef20.AreaID(), taskDef20.Title(), taskDef20.Cost(), taskDef20.RecurrenceInterval(), taskDef20.StartSequence())

	userRepo.On("FindByID", ctx, userA.ID()).Return(userA, nil).Once()
	userRepo.On("FindByTeamID", ctx, teamID).Return([]*user.User{userA, userB}, nil).Once()

	dutyTaskRepo.On("FindByID", ctx, takenTaskID).Return(taskToTake, nil).Once()
	dutyTaskRepo.On("Assign", ctx, takenTaskID, userA.ID(), mock.AnythingOfType("time.Time")).Return(nil).Once()
	dutyTaskRepo.On("Assign", ctx, freeTaskID, userB.ID(), mock.AnythingOfType("time.Time")).Return(nil).Once()

	dutyRepo.On("FindByID", ctx, currentDuty.ID()).Return(currentDuty, nil).Once()
	dutyRepo.On("FindByID", ctx, currentDuty.ID()).Return(refreshedDuty, nil).Once()

	taskRepo.On("FindByID", ctx, takenTaskDefID).Return(taskDef40, nil).Once()
	taskRepo.On("FindByID", ctx, takenTaskDefID).Return(taskDef40, nil).Once()
	taskRepo.On("FindByID", ctx, freeTaskDefID).Return(taskDef20, nil).Once()

	eventBus.On("Publish", ctx, events.TopicTaskAssigned, mock.AnythingOfType("events.TaskAssignedEvent")).Return(nil).Twice()

	err = svc.AssignTask(ctx, takenTaskID, userA.ID())
	require.NoError(t, err)
}
