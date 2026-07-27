package duty

import (
	"context"
	"testing"
	"time"

	"dorm/pkg/core/domain/catalog"
	dutydomain "dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSaveTaskOverrides_ReplacesOnlyDifferencesFromAutomaticState(t *testing.T) {
	ctx := context.Background()
	taskAutoInclude := catalog.RestoreTaskDefinition(uuid.New(), 1, "Auto include", 1, 1)
	taskAutoExclude := catalog.RestoreTaskDefinition(uuid.New(), 1, "Auto exclude", 1, 7)

	lastDuty := dutydomain.NewDuty(uuid.New(), time.Now().Add(-48*time.Hour), time.Now().Add(-24*time.Hour))
	dutyRepo := new(dutyRepoMock)
	overrideRepo := new(overrideRepoMock)
	service := &Service{dutyRepo: dutyRepo, overrideRepo: overrideRepo}

	dutyRepo.On("FindLastByTaskDefID", ctx, taskAutoInclude.ID()).Return(nil, nil).Once()
	dutyRepo.On("FindLastByTaskDefID", ctx, taskAutoExclude.ID()).Return(lastDuty, nil).Once()
	overrideRepo.On("ReplaceForTasks", ctx, []uuid.UUID{taskAutoInclude.ID(), taskAutoExclude.ID()}, mock.MatchedBy(func(overrides []*catalog.DutyTaskOverride) bool {
		if len(overrides) != 2 {
			return false
		}
		return containsOverride(overrides, taskAutoInclude.ID(), false) &&
			containsOverride(overrides, taskAutoExclude.ID(), true)
	})).Return(nil).Once()

	err := service.saveTaskOverrides(ctx, []*catalog.TaskDefinition{taskAutoInclude, taskAutoExclude}, []uuid.UUID{taskAutoExclude.ID()}, time.Now())
	assert.NoError(t, err)
}

func TestSaveTaskOverrides_RemovesOverridesWhenSelectionMatchesAutomaticState(t *testing.T) {
	ctx := context.Background()
	task := catalog.RestoreTaskDefinition(uuid.New(), 1, "Auto include", 1, 1)

	dutyRepo := new(dutyRepoMock)
	overrideRepo := new(overrideRepoMock)
	service := &Service{dutyRepo: dutyRepo, overrideRepo: overrideRepo}

	dutyRepo.On("FindLastByTaskDefID", ctx, task.ID()).Return(nil, nil).Once()
	overrideRepo.On("ReplaceForTasks", ctx, []uuid.UUID{task.ID()}, mock.MatchedBy(func(overrides []*catalog.DutyTaskOverride) bool {
		return len(overrides) == 0
	})).Return(nil).Once()

	err := service.saveTaskOverrides(ctx, []*catalog.TaskDefinition{task}, []uuid.UUID{task.ID()}, time.Now())
	assert.NoError(t, err)
}

func containsOverride(overrides []*catalog.DutyTaskOverride, taskID uuid.UUID, include bool) bool {
	for _, override := range overrides {
		if override.TaskID() == taskID && override.IncludeInNextDuty() == include {
			return true
		}
	}
	return false
}

type dutyRepoMock struct{ mock.Mock }

func (m *dutyRepoMock) CreateWithTasks(_ context.Context, _ *dutydomain.Duty, _ []*dutydomain.DutyTask) error {
	return nil
}
func (m *dutyRepoMock) FindCurrentByTeamID(_ context.Context, _ uuid.UUID) (*dutydomain.Duty, error) {
	return nil, nil
}
func (m *dutyRepoMock) FindActiveByTeamID(_ context.Context, _ uuid.UUID, _ time.Time) (*dutydomain.Duty, error) {
	return nil, nil
}
func (m *dutyRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*dutydomain.Duty, error) {
	return nil, nil
}
func (m *dutyRepoMock) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*dutydomain.Duty, error) {
	return nil, nil
}
func (m *dutyRepoMock) FindLatestByGroupID(ctx context.Context, groupID uuid.UUID) (*dutydomain.Duty, error) {
	return nil, nil
}
func (m *dutyRepoMock) CountDistinctStartDates(ctx context.Context) (int, error) { return 0, nil }
func (m *dutyRepoMock) FindLastByTaskDefID(ctx context.Context, taskDefID uuid.UUID) (*dutydomain.Duty, error) {
	args := m.Called(ctx, taskDefID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dutydomain.Duty), args.Error(1)
}
func (m *dutyRepoMock) FindAllLatest(ctx context.Context) ([]*dutydomain.Duty, error) {
	return nil, nil
}

type overrideRepoMock struct{ mock.Mock }

func (m *overrideRepoMock) FindAll(ctx context.Context) ([]*catalog.DutyTaskOverride, error) {
	return nil, nil
}
func (m *overrideRepoMock) ReplaceForTasks(ctx context.Context, taskIDs []uuid.UUID, overrides []*catalog.DutyTaskOverride) error {
	return m.Called(ctx, taskIDs, overrides).Error(0)
}
func (m *overrideRepoMock) DeleteByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) error {
	return nil
}

type teamRepoStub struct{}

func (teamRepoStub) FindByGroupID(ctx context.Context, id uuid.UUID) ([]*structure.Team, error) {
	return nil, nil
}
func (teamRepoStub) FindByID(ctx context.Context, id uuid.UUID) (*structure.Team, error) {
	return nil, nil
}
func (teamRepoStub) UpdateRotationPositions(ctx context.Context, groupID uuid.UUID, orderedTeamIDs []uuid.UUID) error {
	return nil
}
func (teamRepoStub) Save(ctx context.Context, team *structure.Team) error { return nil }
func (teamRepoStub) Delete(ctx context.Context, id uuid.UUID) error       { return nil }

type groupRepoStub struct{}

func (groupRepoStub) FindAll(ctx context.Context) ([]*structure.Group, error) { return nil, nil }
func (groupRepoStub) FindByID(ctx context.Context, id uuid.UUID) (*structure.Group, error) {
	return nil, nil
}
func (groupRepoStub) FindByDormitoryID(ctx context.Context, dormitoryID int64) ([]*structure.Group, error) {
	return nil, nil
}
func (groupRepoStub) Save(ctx context.Context, group *structure.Group) error { return nil }
func (groupRepoStub) Delete(ctx context.Context, id uuid.UUID) error         { return nil }

type dormRepoStub struct{}

func (dormRepoStub) FindAll(ctx context.Context) ([]*structure.Dormitory, error) { return nil, nil }
func (dormRepoStub) FindByID(ctx context.Context, id int64) (*structure.Dormitory, error) {
	return nil, nil
}
func (dormRepoStub) ExistsByLeaderID(ctx context.Context, leaderID uuid.UUID) (bool, error) {
	return false, nil
}
func (dormRepoStub) Save(ctx context.Context, dormitory *structure.Dormitory) error { return nil }
func (dormRepoStub) Delete(ctx context.Context, id int64) error                     { return nil }

type taskRepoStub struct{}

func (taskRepoStub) GetAllTaskDefinitions(ctx context.Context) ([]*catalog.TaskDefinition, error) {
	return nil, nil
}
func (taskRepoStub) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*catalog.TaskDefinition, error) {
	return nil, nil
}
func (taskRepoStub) FindCommon(ctx context.Context) ([]*catalog.TaskDefinition, error) {
	return nil, nil
}
func (taskRepoStub) FindByID(ctx context.Context, id uuid.UUID) (*catalog.TaskDefinition, error) {
	return nil, nil
}
func (taskRepoStub) Save(ctx context.Context, task *catalog.TaskDefinition) error { return nil }
func (taskRepoStub) Delete(ctx context.Context, id uuid.UUID) error               { return nil }

type areaRepoStub struct{}

func (areaRepoStub) GetAllAreas(ctx context.Context) ([]*catalog.Area, error) { return nil, nil }

type userRepoStub struct{}

func (userRepoStub) Save(ctx context.Context, u *user.User) error                   { return nil }
func (userRepoStub) FindAll(ctx context.Context) ([]*user.User, error)              { return nil, nil }
func (userRepoStub) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) { return nil, nil }
func (userRepoStub) FindByLogin(ctx context.Context, login string) (*user.User, error) {
	return nil, nil
}
func (userRepoStub) FindByTelegramID(ctx context.Context, id int64) (*user.User, error) {
	return nil, nil
}
func (userRepoStub) FindByTeamID(ctx context.Context, id uuid.UUID) ([]*user.User, error) {
	return nil, nil
}
func (userRepoStub) FindByDormitoryID(ctx context.Context, dormitoryID int64) ([]*user.User, error) {
	return nil, nil
}
func (userRepoStub) MoveUserToTeam(ctx context.Context, userID uuid.UUID, teamID *uuid.UUID) error {
	return nil
}
func (userRepoStub) SoftDelete(ctx context.Context, id uuid.UUID) error { return nil }
