package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "dorm/pkg/core/domain/notification"
	"dorm/pkg/core/ports"
	queryports "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type activeDutyQueryStub struct {
	result []queryports.ActiveDuty
	err    error
}

func (s *activeDutyQueryStub) FindAllActive(context.Context, time.Time) ([]queryports.ActiveDuty, error) {
	return s.result, s.err
}

type dutyTeamMembersQueryStub struct {
	result map[uuid.UUID][]uuid.UUID
	err    error
}

func (s *dutyTeamMembersQueryStub) FindUserIDsByTeamID(_ context.Context, teamID uuid.UUID) ([]uuid.UUID, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.result[teamID], nil
}

type dutyUsersWithoutAssignedTasksQueryStub struct {
	result map[uuid.UUID][]uuid.UUID
	err    error
}

func (s *dutyUsersWithoutAssignedTasksQueryStub) FindUserIDs(_ context.Context, dutyID uuid.UUID, _ uuid.UUID) ([]uuid.UUID, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.result[dutyID], nil
}

type dutyFinishReminderQueryStub struct {
	result map[uuid.UUID]queryports.DutyFinishReminderState
	err    error
}

func (s *dutyFinishReminderQueryStub) GetByDutyID(_ context.Context, dutyID uuid.UUID, _ uuid.UUID) (queryports.DutyFinishReminderState, error) {
	if s.err != nil {
		return queryports.DutyFinishReminderState{}, s.err
	}
	return s.result[dutyID], nil
}

type userNotificationUseCaseStub struct {
	calls []ports.NotifyUserCommand
	err   error
}

func (s *userNotificationUseCaseStub) NotifyUser(_ context.Context, command ports.NotifyUserCommand) error {
	s.calls = append(s.calls, command)
	return s.err
}

func TestDutyReminderServiceSendDutyStartedReminders(t *testing.T) {
	dutyID := uuid.New()
	teamID := uuid.New()
	userA := uuid.New()
	userB := uuid.New()

	service := NewDutyReminderService(
		&activeDutyQueryStub{result: []queryports.ActiveDuty{{ID: dutyID, TeamID: teamID}}},
		&dutyTeamMembersQueryStub{result: map[uuid.UUID][]uuid.UUID{teamID: {userA, userB}}},
		&dutyFinishReminderQueryStub{},
		&userNotificationUseCaseStub{},
	)

	notifications := service.notifications.(*userNotificationUseCaseStub)

	err := service.SendDutyStartedReminders(context.Background(), time.Now())
	require.NoError(t, err)
	require.Len(t, notifications.calls, 2)
	assert.Equal(t, domain.NotificationDutyStarted, notifications.calls[0].Type)
	assert.Equal(t, "/app/current-duty", notifications.calls[0].TargetURL)
}

func TestDutyReminderServiceSendSundayTakeTaskRemindersFiltersRecipients(t *testing.T) {
	dutyID := uuid.New()
	teamID := uuid.New()
	userA := uuid.New()

	service := NewDutyReminderService(
		&activeDutyQueryStub{result: []queryports.ActiveDuty{{ID: dutyID, TeamID: teamID}}},
		&dutyTeamMembersQueryStub{},
		&dutyFinishReminderQueryStub{result: map[uuid.UUID]queryports.DutyFinishReminderState{
			dutyID: {
				FreeTaskCount:             1,
				UsersBelowAssignedGoalIDs: []uuid.UUID{userA},
			},
		}},
		&userNotificationUseCaseStub{},
	)

	notifications := service.notifications.(*userNotificationUseCaseStub)

	err := service.SendSundayTakeTaskReminders(context.Background(), time.Now())
	require.NoError(t, err)
	require.Len(t, notifications.calls, 1)
	assert.Equal(t, userA, notifications.calls[0].UserID)
	assert.Equal(t, domain.NotificationTakeTasksSundayReminder, notifications.calls[0].Type)
}

func TestDutyReminderServiceSendSundayFinishTaskRemindersIncludesTeamWhenFreeTasksExist(t *testing.T) {
	dutyID := uuid.New()
	teamID := uuid.New()
	userA := uuid.New()
	userB := uuid.New()

	service := NewDutyReminderService(
		&activeDutyQueryStub{result: []queryports.ActiveDuty{{ID: dutyID, TeamID: teamID}}},
		&dutyTeamMembersQueryStub{result: map[uuid.UUID][]uuid.UUID{teamID: {userA, userB}}},
		&dutyFinishReminderQueryStub{result: map[uuid.UUID]queryports.DutyFinishReminderState{
			dutyID: {
				FreeTaskCount:              1,
				UsersBelowAssignedGoalIDs:  []uuid.UUID{userB},
				UsersWithIncompleteTaskIDs: []uuid.UUID{userA},
			},
		}},
		&userNotificationUseCaseStub{},
	)

	notifications := service.notifications.(*userNotificationUseCaseStub)

	err := service.SendSundayFinishTaskReminders(context.Background(), time.Now())
	require.NoError(t, err)
	require.Len(t, notifications.calls, 2)
}

func TestDutyReminderServiceReturnsActiveDutyError(t *testing.T) {
	service := NewDutyReminderService(
		&activeDutyQueryStub{err: errors.New("db failure")},
		&dutyTeamMembersQueryStub{},
		&dutyFinishReminderQueryStub{},
		&userNotificationUseCaseStub{},
	)

	err := service.SendSaturdayTaskReminders(context.Background(), time.Now())
	require.Error(t, err)
	assert.ErrorContains(t, err, "find active duties")
}
