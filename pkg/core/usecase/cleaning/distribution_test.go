package cleaning

import (
	"context"
	"errors"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/ports/dto"
)

type testFixtures struct {
	leaderID1, leaderID2    uuid.UUID
	teamID1, teamID2        uuid.UUID
	taskPrivate, taskPublic *catalog.TaskDefinition
	taskOtherGroup          *catalog.TaskDefinition
	groups                  []*structure.Group
	teams1, teams2          []*structure.Team
	areas                   []*catalog.Area
	tasks                   []*catalog.TaskDefinition
}

type testEnv struct {
	groupRepo    *MockGroupRepo
	teamRepo     *MockTeamRepo
	dutyRepo     *MockDutyRepo
	catRepo      *MockCatalogRepo
	overrideRepo *MockTaskOverrideRepo
	areaRepo     *MockAreaRepo
	eventBus     *MockEventBus
	service      *Service
}

func TestStartNewWeek_DistributionByAreaScope(t *testing.T) {
	data := createFixtures()

	t.Run("Week 0 - Public goes to first team (Boys)", func(t *testing.T) {
		runWeekTest(t, data, 0, func(dBoys, dGirls *duty.Duty) {
			assert.True(t, hasTask(dBoys, data.taskPrivate.ID()), "Boys must have private task")
			assert.True(t, hasTask(dBoys, data.taskPublic.ID()), "Boys must have public task (RR index 0)")

			assert.False(t, hasTask(dGirls, data.taskPrivate.ID()), "Girls must NOT have boys task")
			assert.False(t, hasTask(dGirls, data.taskPublic.ID()), "Girls must NOT have public task yet")
		})
	})

	t.Run("Week 1 - Public goes to second team (Girls)", func(t *testing.T) {
		runWeekTest(t, data, 1, func(dBoys, dGirls *duty.Duty) {
			assert.True(t, hasTask(dBoys, data.taskPrivate.ID()), "Boys STILL have private task")

			assert.False(t, hasTask(dBoys, data.taskPublic.ID()), "Boys should NOT have public task")
			assert.True(t, hasTask(dGirls, data.taskPublic.ID()), "Girls MUST have public task (RR index 1)")
		})
	})
}

func TestDetermineNextTeam_UsesLastDutyRotationPosition(t *testing.T) {
	groupID := uuid.New()
	firstTeamID := uuid.New()
	secondTeamID := uuid.New()
	teams := []*structure.Team{
		structure.RestoreTeam(secondTeamID, "Second Team", groupID, nil, "pink", 20),
		structure.RestoreTeam(firstTeamID, "First Team", groupID, nil, "blue", 10),
	}

	selectedTeam, err := determineNextTeam(teams, nil)

	assert.NoError(t, err)
	assert.Equal(t, firstTeamID, selectedTeam.ID())

	lastDuty := duty.RestoreDuty(uuid.New(), firstTeamID, time.Now().Add(-24*time.Hour), time.Now(), 1, nil)
	selectedTeam, err = determineNextTeam(teams, lastDuty)

	assert.NoError(t, err)
	assert.Equal(t, secondTeamID, selectedTeam.ID())
}

func TestStartNewWeek_AppliesOverrideIncludeForTaskNotDue(t *testing.T) {
	ctx := context.Background()
	data := createFixtures()
	data.taskPublic = catalog.RestoreTaskDefinition(uuid.New(), data.areas[1].ID(), "Clean Kitchen", 10, 0, 1)
	data.tasks = []*catalog.TaskDefinition{data.taskPrivate, data.taskPublic}

	env := newTestEnv()
	env.groupRepo.On("FindAll", ctx).Return(data.groups, nil)
	env.teamRepo.On("FindByGroupID", ctx, data.groups[0].ID()).Return(data.teams1, nil)
	env.teamRepo.On("FindByGroupID", ctx, data.groups[1].ID()).Return(data.teams2, nil)
	env.dutyRepo.On("FindLatestByGroupID", ctx, data.groups[0].ID()).Return(nil, nil)
	env.dutyRepo.On("FindLatestByGroupID", ctx, data.groups[1].ID()).Return(nil, nil)
	env.areaRepo.On("GetAllAreas", ctx).Return(data.areas, nil)
	env.catRepo.On("GetAllTaskDefinitions", ctx).Return(data.tasks, nil)
	env.dutyRepo.On("CountDistinctStartDates", ctx).Return(0, nil).Once()

	env.overrideRepo.On("FindAll", ctx).Return([]*catalog.DutyTaskOverride{
		catalog.NewDutyTaskOverride(data.taskPublic.ID(), true),
	}, nil)
	env.overrideRepo.On("DeleteByTaskIDs", ctx, mock.Anything).Return(nil)
	env.dutyRepo.On("FindAllLatest", ctx).Return(nil, errors.New("snapshot unavailable"))
	env.eventBus.On("Publish", mock.Anything, events.TopicWeekStarted, mock.Anything).Return(nil)

	var savedDuties []*duty.Duty
	env.dutyRepo.On("CreateWithTasks", ctx, mock.MatchedBy(func(d *duty.Duty) bool {
		savedDuties = append(savedDuties, d)
		return true
	}), mock.Anything).Return(nil)

	err := env.service.StartNewWeek(ctx)
	assert.NoError(t, err)

	dBoys := findDutyByTeam(savedDuties, data.teamID1)
	assert.True(t, hasTask(dBoys, data.taskPublic.ID()))
}

func TestStartNewWeek_AppliesOverrideExcludeForTaskDue(t *testing.T) {
	ctx := context.Background()
	data := createFixtures()

	env := newTestEnv()
	env.groupRepo.On("FindAll", ctx).Return(data.groups, nil)
	env.teamRepo.On("FindByGroupID", ctx, data.groups[0].ID()).Return(data.teams1, nil)
	env.teamRepo.On("FindByGroupID", ctx, data.groups[1].ID()).Return(data.teams2, nil)
	env.dutyRepo.On("FindLatestByGroupID", ctx, data.groups[0].ID()).Return(nil, nil)
	env.dutyRepo.On("FindLatestByGroupID", ctx, data.groups[1].ID()).Return(nil, nil)
	env.areaRepo.On("GetAllAreas", ctx).Return(data.areas, nil)
	env.catRepo.On("GetAllTaskDefinitions", ctx).Return(data.tasks, nil)
	env.dutyRepo.On("CountDistinctStartDates", ctx).Return(0, nil).Once()
	env.overrideRepo.On("FindAll", ctx).Return([]*catalog.DutyTaskOverride{
		catalog.NewDutyTaskOverride(data.taskPublic.ID(), false),
	}, nil)
	env.overrideRepo.On("DeleteByTaskIDs", ctx, mock.Anything).Return(nil)
	env.dutyRepo.On("FindAllLatest", ctx).Return(nil, errors.New("snapshot unavailable"))
	env.eventBus.On("Publish", mock.Anything, events.TopicWeekStarted, mock.Anything).Return(nil)

	var savedDuties []*duty.Duty
	env.dutyRepo.On("CreateWithTasks", ctx, mock.MatchedBy(func(d *duty.Duty) bool {
		savedDuties = append(savedDuties, d)
		return true
	}), mock.Anything).Return(nil)

	err := env.service.StartNewWeek(ctx)
	assert.NoError(t, err)

	dBoys := findDutyByTeam(savedDuties, data.teamID1)
	assert.False(t, hasTask(dBoys, data.taskPublic.ID()))
}

func TestStartNewWeek_DoesNotClearOverridesWhenSaveFails(t *testing.T) {
	ctx := context.Background()
	data := createFixtures()

	env := newTestEnv()
	setupCommonExpectations(ctx, env, data)
	env.dutyRepo.On("CountDistinctStartDates", ctx).Return(0, nil).Once()
	env.dutyRepo.On("CreateWithTasks", ctx, mock.Anything, mock.Anything).Return(errors.New("save failed")).Once()

	err := env.service.StartNewWeek(ctx)
	assert.Error(t, err)
	env.overrideRepo.AssertNotCalled(t, "DeleteByTaskIDs", mock.Anything, mock.Anything)
}

func TestStartNewDutyForGroup_CreatesOnlySelectedGroupDuty(t *testing.T) {
	ctx := context.Background()
	data := createFixtures()
	startDate := time.Date(2026, 7, 29, 9, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 7)

	env := newTestEnv()
	env.groupRepo.On("FindByID", ctx, data.groups[0].ID()).Return(data.groups[0], nil).Twice()
	env.teamRepo.On("FindByGroupID", ctx, data.groups[0].ID()).Return(data.teams1, nil)
	env.dutyRepo.On("FindLatestByGroupID", ctx, data.groups[0].ID()).Return(nil, nil)
	env.areaRepo.On("GetAllAreas", ctx).Return(data.areas, nil)
	env.catRepo.On("FindByGroupID", ctx, data.groups[0].ID()).Return([]*catalog.TaskDefinition{data.taskPrivate}, nil)
	env.dutyRepo.On("CountDistinctStartDates", ctx).Return(3, nil).Once()
	env.overrideRepo.On("FindAll", ctx).Return([]*catalog.DutyTaskOverride{
		catalog.NewDutyTaskOverride(data.taskOtherGroup.ID(), false),
	}, nil)
	env.overrideRepo.On("DeleteByTaskIDs", ctx, mock.MatchedBy(func(taskIDs []uuid.UUID) bool {
		return len(taskIDs) == 1 && taskIDs[0] == data.taskPrivate.ID()
	})).Return(nil).Once()
	env.dutyRepo.On("FindAllLatest", ctx).Return(nil, errors.New("snapshot unavailable"))
	env.eventBus.On("Publish", mock.Anything, events.TopicWeekStarted, mock.Anything).Return(nil)

	var savedDuties []*duty.Duty
	env.dutyRepo.On("CreateWithTasks", ctx, mock.MatchedBy(func(d *duty.Duty) bool {
		savedDuties = append(savedDuties, d)
		return true
	}), mock.Anything).Return(nil).Once()

	err := env.service.StartNewDutyForGroup(ctx, data.leaderID1, data.groups[0].ID(), startDate, endDate)
	require.NoError(t, err)
	require.Len(t, savedDuties, 1)
	assert.Equal(t, data.teamID1, savedDuties[0].TeamID())
	assert.True(t, hasTask(savedDuties[0], data.taskPrivate.ID()))
	assert.False(t, hasTask(savedDuties[0], data.taskPublic.ID()))
	assert.False(t, hasTask(savedDuties[0], data.taskOtherGroup.ID()))
	env.teamRepo.AssertNotCalled(t, "FindByGroupID", ctx, data.groups[1].ID())
}

func TestStartNewDutyForGroup_AccessDeniedForNonLeader(t *testing.T) {
	ctx := context.Background()
	data := createFixtures()
	otherUserID := uuid.New()

	env := newTestEnv()
	env.groupRepo.On("FindByID", ctx, data.groups[0].ID()).Return(data.groups[0], nil).Once()

	err := env.service.StartNewDutyForGroup(ctx, otherUserID, data.groups[0].ID(), time.Now(), time.Now().Add(time.Hour))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDutyCreationAccessDenied)
	env.teamRepo.AssertNotCalled(t, "FindByGroupID", mock.Anything, mock.Anything)
}

func TestStartNewDutyForGroup_GroupNotFound(t *testing.T) {
	ctx := context.Background()
	groupID := uuid.New()

	env := newTestEnv()
	env.groupRepo.On("FindByID", ctx, groupID).Return(nil, nil).Once()

	err := env.service.StartNewDutyForGroup(ctx, uuid.New(), groupID, time.Now(), time.Now().Add(time.Hour))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDutyGroupNotFound)
}

func TestStartNewDutyForGroup_MapsUniqueConflict(t *testing.T) {
	ctx := context.Background()
	data := createFixtures()
	startDate := time.Date(2026, 7, 29, 9, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 7)

	env := newTestEnv()
	env.groupRepo.On("FindByID", ctx, data.groups[0].ID()).Return(data.groups[0], nil).Twice()
	env.teamRepo.On("FindByGroupID", ctx, data.groups[0].ID()).Return(data.teams1, nil)
	env.dutyRepo.On("FindLatestByGroupID", ctx, data.groups[0].ID()).Return(nil, nil)
	env.areaRepo.On("GetAllAreas", ctx).Return(data.areas, nil)
	env.catRepo.On("FindByGroupID", ctx, data.groups[0].ID()).Return([]*catalog.TaskDefinition{data.taskPrivate}, nil)
	env.dutyRepo.On("CountDistinctStartDates", ctx).Return(0, nil).Once()
	env.overrideRepo.On("FindAll", ctx).Return([]*catalog.DutyTaskOverride{}, nil)
	env.dutyRepo.On("CreateWithTasks", ctx, mock.Anything, mock.Anything).Return(&mysqlDriver.MySQLError{
		Number:  1062,
		Message: "Duplicate entry 'x' for key 'uq_duty_group_sequence'",
	}).Once()

	err := env.service.StartNewDutyForGroup(ctx, data.leaderID1, data.groups[0].ID(), startDate, endDate)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDutyAlreadyCreated)
	env.overrideRepo.AssertNotCalled(t, "DeleteByTaskIDs", mock.Anything, mock.Anything)
}

func TestFilterDutyViewModels_ReturnsOnlyCreatedDuties(t *testing.T) {
	createdDuty := duty.NewDuty(uuid.New(), time.Now(), time.Now().Add(time.Hour), 1)
	otherDuty := duty.NewDuty(uuid.New(), time.Now(), time.Now().Add(time.Hour), 2)

	filtered := filterDutyViewModels([]dto.DutyViewModel{
		{DutyID: createdDuty.ID()},
		{DutyID: otherDuty.ID()},
	}, []*duty.Duty{createdDuty})

	require.Len(t, filtered, 1)
	assert.Equal(t, createdDuty.ID(), filtered[0].DutyID)
}

func runWeekTest(t *testing.T, f *testFixtures, weekNum int, assertions func(dBoys, dGirls *duty.Duty)) {
	ctx := context.Background()
	env := newTestEnv()

	setupCommonExpectations(ctx, env, f)

	env.dutyRepo.On("CountDistinctStartDates", ctx).Return(weekNum, nil).Once()

	var savedDuties []*duty.Duty
	env.dutyRepo.On("CreateWithTasks", ctx, mock.MatchedBy(func(d *duty.Duty) bool {
		savedDuties = append(savedDuties, d)
		return true
	}), mock.Anything).Return(nil)

	err := env.service.StartNewWeek(ctx)
	assert.NoError(t, err)

	dBoys := findDutyByTeam(savedDuties, f.teamID1)
	dGirls := findDutyByTeam(savedDuties, f.teamID2)

	assertions(dBoys, dGirls)
}

func createFixtures() *testFixtures {
	teamID1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	teamID2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	groupID1, groupID2 := uuid.New(), uuid.New()
	leaderID1, leaderID2 := uuid.New(), uuid.New()

	areaBoys := catalog.RestoreArea(1, "Boys Room", 1, &groupID1)
	areaKit := catalog.RestoreArea(2, "Kitchen", 1, nil)
	areaGirls := catalog.RestoreArea(3, "Girls Room", 1, &groupID2)

	taskPrivate := catalog.RestoreTaskDefinition(uuid.New(), areaBoys.ID(), "Clean Boys", 5, 1, 1)
	taskPublic := catalog.RestoreTaskDefinition(uuid.New(), areaKit.ID(), "Clean Kitchen", 10, 1, 1)
	taskOtherGroup := catalog.RestoreTaskDefinition(uuid.New(), areaGirls.ID(), "Clean Girls", 7, 1, 1)

	return &testFixtures{
		leaderID1:      leaderID1,
		leaderID2:      leaderID2,
		teamID1:        teamID1,
		teamID2:        teamID2,
		taskPrivate:    taskPrivate,
		taskPublic:     taskPublic,
		taskOtherGroup: taskOtherGroup,
		groups: []*structure.Group{
			structure.RestoreGroup(groupID1, &leaderID1, "Boys", "s1", 1),
			structure.RestoreGroup(groupID2, &leaderID2, "Girls", "s2", 1),
		},
		teams1: []*structure.Team{structure.RestoreTeam(teamID1, "Boys Team", groupID1, &leaderID1, "blue", 1)},
		teams2: []*structure.Team{structure.RestoreTeam(teamID2, "Girls Team", groupID2, &leaderID2, "pink", 2)},
		areas:  []*catalog.Area{areaBoys, areaKit, areaGirls},
		tasks:  []*catalog.TaskDefinition{taskPrivate, taskPublic, taskOtherGroup},
	}
}

func newTestEnv() *testEnv {
	e := &testEnv{
		groupRepo:    new(MockGroupRepo),
		teamRepo:     new(MockTeamRepo),
		dutyRepo:     new(MockDutyRepo),
		catRepo:      new(MockCatalogRepo),
		overrideRepo: new(MockTaskOverrideRepo),
		areaRepo:     new(MockAreaRepo),
		eventBus:     new(MockEventBus),
	}
	e.service = NewCleaningService(nil, e.teamRepo, e.groupRepo, e.dutyRepo, nil, e.catRepo, e.overrideRepo, e.areaRepo, e.eventBus)
	return e
}

func setupCommonExpectations(ctx context.Context, e *testEnv, f *testFixtures) {
	e.groupRepo.On("FindAll", ctx).Return(f.groups, nil)
	e.teamRepo.On("FindByGroupID", ctx, f.groups[0].ID()).Return(f.teams1, nil)
	e.teamRepo.On("FindByGroupID", ctx, f.groups[1].ID()).Return(f.teams2, nil)
	e.dutyRepo.On("FindLatestByGroupID", ctx, f.groups[0].ID()).Return(nil, nil)
	e.dutyRepo.On("FindLatestByGroupID", ctx, f.groups[1].ID()).Return(nil, nil)
	e.areaRepo.On("GetAllAreas", ctx).Return(f.areas, nil)
	e.catRepo.On("GetAllTaskDefinitions", ctx).Return(f.tasks, nil)
	e.overrideRepo.On("FindAll", ctx).Return([]*catalog.DutyTaskOverride{}, nil)
	e.overrideRepo.On("DeleteByTaskIDs", ctx, mock.Anything).Return(nil)

	e.dutyRepo.On("FindAllLatest", ctx).Return(nil, errors.New("snapshot unavailable"))

	e.eventBus.On("Publish", mock.Anything, events.TopicWeekStarted, mock.Anything).Return(nil)
}

func findDutyByTeam(duties []*duty.Duty, teamID uuid.UUID) *duty.Duty {
	for _, d := range duties {
		if d.TeamID() == teamID {
			return d
		}
	}
	return nil
}

func hasTask(d *duty.Duty, taskDefID uuid.UUID) bool {
	if d == nil {
		return false
	}
	for _, t := range d.Tasks() {
		if t.TaskDefID() == taskDefID {
			return true
		}
	}
	return false
}
