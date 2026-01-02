package cleaning

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/domain/structure"
)

type testFixtures struct {
	teamID1, teamID2        uuid.UUID
	taskPrivate, taskPublic *catalog.TaskDefinition
	groups                  []*structure.Group
	teams1, teams2          []*structure.Team
	areas                   []*catalog.Area
	tasks                   []*catalog.TaskDefinition
}

type testEnv struct {
	groupRepo *MockGroupRepo
	teamRepo  *MockTeamRepo
	dutyRepo  *MockDutyRepo
	catRepo   *MockCatalogRepo
	areaRepo  *MockAreaRepo
	eventBus  *MockEventBus
	service   *Service
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

func runWeekTest(t *testing.T, f *testFixtures, weekNum int, assertions func(dBoys, dGirls *duty.Duty)) {
	ctx := context.Background()
	env := newTestEnv()

	setupCommonExpectations(ctx, env, f)

	env.dutyRepo.On("CountDistinctStartDates", ctx).Return(weekNum, nil).Once()

	var savedDuties []*duty.Duty
	env.dutyRepo.On("Save", ctx, mock.MatchedBy(func(d *duty.Duty) bool {
		savedDuties = append(savedDuties, d)
		return true
	})).Return(nil)

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

	areaBoys := catalog.RestoreArea(1, "Boys Room", 1, &groupID1)
	areaKit := catalog.RestoreArea(2, "Kitchen", 1, nil)

	taskPrivate := catalog.RestoreTaskDefinition(uuid.New(), areaBoys.ID(), "Clean Boys", 5, 1)
	taskPublic := catalog.RestoreTaskDefinition(uuid.New(), areaKit.ID(), "Clean Kitchen", 10, 1)

	return &testFixtures{
		teamID1:     teamID1,
		teamID2:     teamID2,
		taskPrivate: taskPrivate,
		taskPublic:  taskPublic,
		groups: []*structure.Group{
			structure.RestoreGroup(groupID1, "Boys", "s1", 1),
			structure.RestoreGroup(groupID2, "Girls", "s2", 1),
		},
		teams1: []*structure.Team{structure.RestoreTeam(teamID1, groupID1, nil, "blue", 1)},
		teams2: []*structure.Team{structure.RestoreTeam(teamID2, groupID2, nil, "pink", 2)},
		areas:  []*catalog.Area{areaBoys, areaKit},
		tasks:  []*catalog.TaskDefinition{taskPrivate, taskPublic},
	}
}

func newTestEnv() *testEnv {
	e := &testEnv{
		groupRepo: new(MockGroupRepo),
		teamRepo:  new(MockTeamRepo),
		dutyRepo:  new(MockDutyRepo),
		catRepo:   new(MockCatalogRepo),
		areaRepo:  new(MockAreaRepo),
		eventBus:  new(MockEventBus),
	}
	e.service = NewCleaningService(nil, e.teamRepo, e.groupRepo, e.dutyRepo, e.catRepo, e.areaRepo, e.eventBus)
	return e
}

func setupCommonExpectations(ctx context.Context, e *testEnv, f *testFixtures) {
	e.groupRepo.On("FindAll", ctx).Return(f.groups, nil)
	e.teamRepo.On("FindByGroupID", ctx, f.groups[0].ID()).Return(f.teams1, nil)
	e.teamRepo.On("FindByGroupID", ctx, f.groups[1].ID()).Return(f.teams2, nil)
	e.areaRepo.On("GetAllAreas", ctx).Return(f.areas, nil)
	e.catRepo.On("GetAllTaskDefinitions", ctx).Return(f.tasks, nil)

	e.dutyRepo.On("FindCurrentByTeamID", ctx, mock.Anything).Return(nil, nil)
	e.dutyRepo.On("FindLastByTaskDefID", ctx, mock.Anything).Return(nil, nil)

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
