package dutysettings

import (
	"context"
	"testing"
	"time"

	"dorm/pkg/core/domain/duty"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type activeDutyRepositoryStub struct {
	duty.DutyRepository
	activeDuty *duty.Duty
	calls      int
}

func (s *activeDutyRepositoryStub) FindActiveByGroupID(_ context.Context, _ uuid.UUID, _ time.Time) (*duty.Duty, error) {
	s.calls++
	return s.activeDuty, nil
}

func TestRequireActiveDutyUsesCalendarActiveDutyQuery(t *testing.T) {
	t.Parallel()

	groupID := uuid.New()
	activeDuty := duty.RestoreDuty(
		uuid.New(),
		uuid.New(),
		time.Date(2026, time.August, 24, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC),
		5,
		nil,
	)
	repo := &activeDutyRepositoryStub{activeDuty: activeDuty}
	service := &Service{
		dutyRepo: repo,
		now:      func() time.Time { return time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC) },
	}

	actual, err := service.requireActiveDuty(context.Background(), groupID)

	require.NoError(t, err)
	require.Same(t, activeDuty, actual)
	require.Equal(t, 1, repo.calls)
}

func TestRequireActiveDutyRejectsGroupWithoutCalendarActiveDuty(t *testing.T) {
	t.Parallel()

	service := &Service{
		dutyRepo: &activeDutyRepositoryStub{},
		now:      time.Now,
	}

	actual, err := service.requireActiveDuty(context.Background(), uuid.New())

	require.Nil(t, actual)
	require.EqualError(t, err, "В группе нет активного дежурства")
}
