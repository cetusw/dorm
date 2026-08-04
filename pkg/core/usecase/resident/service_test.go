package resident

import (
	"context"
	"testing"
	"time"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userRepositoryStub struct {
	user          *user.User
	teamResidents []*user.User
}

func (s *userRepositoryStub) Save(context.Context, *user.User) error { return nil }
func (s *userRepositoryStub) FindAll(context.Context) ([]*user.User, error) {
	return nil, nil
}
func (s *userRepositoryStub) FindByID(context.Context, uuid.UUID) (*user.User, error) {
	return s.user, nil
}
func (s *userRepositoryStub) FindByLogin(context.Context, string) (*user.User, error) {
	return nil, nil
}
func (s *userRepositoryStub) FindByTeamID(context.Context, uuid.UUID) ([]*user.User, error) {
	return s.teamResidents, nil
}
func (s *userRepositoryStub) FindByDormitoryID(context.Context, int64) ([]*user.User, error) {
	return nil, nil
}
func (s *userRepositoryStub) MoveUserToTeam(context.Context, uuid.UUID, *uuid.UUID) error { return nil }
func (s *userRepositoryStub) SoftDelete(context.Context, uuid.UUID) error                 { return nil }

type teamRepositoryStub struct {
	teamsByGroup map[uuid.UUID][]*structure.Team
}

func (s *teamRepositoryStub) FindByID(context.Context, uuid.UUID) (*structure.Team, error) {
	return nil, nil
}
func (s *teamRepositoryStub) FindByGroupID(_ context.Context, groupID uuid.UUID) ([]*structure.Team, error) {
	return s.teamsByGroup[groupID], nil
}
func (s *teamRepositoryStub) UpdateRotationPositions(context.Context, uuid.UUID, []uuid.UUID) error {
	return nil
}
func (s *teamRepositoryStub) Save(context.Context, *structure.Team) error { return nil }
func (s *teamRepositoryStub) Delete(context.Context, uuid.UUID) error     { return nil }

type groupRepositoryStub struct {
	groups []*structure.Group
}

func (s *groupRepositoryStub) FindAll(context.Context) ([]*structure.Group, error) { return nil, nil }
func (s *groupRepositoryStub) FindByID(_ context.Context, id uuid.UUID) (*structure.Group, error) {
	for _, group := range s.groups {
		if group.ID() == id {
			return group, nil
		}
	}
	return nil, nil
}
func (s *groupRepositoryStub) FindByDormitoryID(context.Context, int64) ([]*structure.Group, error) {
	return s.groups, nil
}
func (s *groupRepositoryStub) Save(context.Context, *structure.Group) error { return nil }
func (s *groupRepositoryStub) Delete(context.Context, uuid.UUID) error      { return nil }

type dormitoryRepositoryStub struct {
	dormitory *structure.Dormitory
}

func (s *dormitoryRepositoryStub) FindAll(context.Context) ([]*structure.Dormitory, error) {
	return nil, nil
}
func (s *dormitoryRepositoryStub) FindByID(context.Context, int64) (*structure.Dormitory, error) {
	return s.dormitory, nil
}
func (s *dormitoryRepositoryStub) ExistsByLeaderID(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (s *dormitoryRepositoryStub) Save(context.Context, *structure.Dormitory) error { return nil }
func (s *dormitoryRepositoryStub) Delete(context.Context, int64) error              { return nil }

type dutyRepositoryStub struct {
	activeByTeam map[uuid.UUID]*duty.Duty
	byID         map[uuid.UUID]*duty.Duty
}

func (s *dutyRepositoryStub) CreateWithTasks(context.Context, *duty.Duty, []*duty.DutyTask) error {
	return nil
}
func (s *dutyRepositoryStub) FindCurrentByTeamID(context.Context, uuid.UUID) (*duty.Duty, error) {
	return nil, nil
}
func (s *dutyRepositoryStub) FindActiveByTeamID(_ context.Context, teamID uuid.UUID, _ time.Time) (*duty.Duty, error) {
	return s.activeByTeam[teamID], nil
}
func (s *dutyRepositoryStub) FindLatestByTeamID(_ context.Context, teamID uuid.UUID) (*duty.Duty, error) {
	for _, currentDuty := range s.byID {
		if currentDuty.TeamID() == teamID {
			return currentDuty, nil
		}
	}
	return s.activeByTeam[teamID], nil
}
func (s *dutyRepositoryStub) FindByID(_ context.Context, id uuid.UUID) (*duty.Duty, error) {
	return s.byID[id], nil
}
func (s *dutyRepositoryStub) FindByGroupID(context.Context, uuid.UUID) ([]*duty.Duty, error) {
	return nil, nil
}
func (s *dutyRepositoryStub) FindLatestByGroupID(context.Context, uuid.UUID) (*duty.Duty, error) {
	return nil, nil
}
func (s *dutyRepositoryStub) FindHistoryByGroupID(context.Context, uuid.UUID) ([]duty.DutyHistoryEntry, error) {
	return nil, nil
}
func (s *dutyRepositoryStub) ReassignTeamAndResetTasks(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (s *dutyRepositoryStub) CountDistinctStartDates(context.Context) (int, error) { return 0, nil }
func (s *dutyRepositoryStub) FindLastByTaskDefID(context.Context, uuid.UUID) (*duty.Duty, error) {
	return nil, nil
}
func (s *dutyRepositoryStub) FindAllLatest(context.Context) ([]*duty.Duty, error) {
	return nil, nil
}

type taskDefinitionRepositoryStub struct{}

func (s *taskDefinitionRepositoryStub) GetAllTaskDefinitions(context.Context) ([]*catalog.TaskDefinition, error) {
	return []*catalog.TaskDefinition{}, nil
}
func (s *taskDefinitionRepositoryStub) FindLastCompletionDates(context.Context, []uuid.UUID) (map[uuid.UUID]*time.Time, error) {
	return map[uuid.UUID]*time.Time{}, nil
}
func (s *taskDefinitionRepositoryStub) FindByGroupID(context.Context, uuid.UUID) ([]*catalog.TaskDefinition, error) {
	return nil, nil
}
func (s *taskDefinitionRepositoryStub) FindCommon(context.Context) ([]*catalog.TaskDefinition, error) {
	return nil, nil
}
func (s *taskDefinitionRepositoryStub) FindByID(context.Context, uuid.UUID) (*catalog.TaskDefinition, error) {
	return nil, nil
}
func (s *taskDefinitionRepositoryStub) Save(context.Context, *catalog.TaskDefinition) error {
	return nil
}
func (s *taskDefinitionRepositoryStub) SoftDelete(context.Context, uuid.UUID) error { return nil }

type areaRepositoryStub struct{}

func (s *areaRepositoryStub) GetAllAreas(context.Context) ([]*catalog.Area, error) {
	return []*catalog.Area{}, nil
}
func (s *areaRepositoryStub) FindByID(context.Context, int) (*catalog.Area, error) {
	return nil, nil
}
func (s *areaRepositoryStub) Save(context.Context, *catalog.Area) error { return nil }
func (s *areaRepositoryStub) Delete(context.Context, int) error         { return nil }

func TestGetCurrentDutyFallsBackToObserverViewForResidentWithoutTeam(t *testing.T) {
	t.Parallel()

	residentID := uuid.New()
	teamMemberID := uuid.New()
	groupID := uuid.New()
	teamID := uuid.New()
	dormitoryID := int64(7)
	now := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)

	resident := user.RestoreUser(
		residentID,
		"resident",
		"hash",
		"Ivan",
		nil,
		"Ivanov",
		nil,
		nil,
		nil,
		&dormitoryID,
		now,
	)
	teamMember := user.RestoreUser(
		teamMemberID,
		"member",
		"hash",
		"Petr",
		nil,
		"Petrov",
		&teamID,
		nil,
		nil,
		&dormitoryID,
		now,
	)
	group := structure.RestoreGroup(groupID, nil, "Group A", dormitoryID)
	team := structure.RestoreTeam(teamID, "Team A", groupID, nil, "#00AAFF", 1)
	activeDuty := duty.RestoreDuty(
		uuid.New(),
		teamID,
		now.Add(-24*time.Hour),
		now.Add(24*time.Hour),
		1,
		nil,
	)
	dormitory := structure.RestoreDormitory(dormitoryID, "Dorm", nil, "Moscow", "st", "Lenina", "1")

	service := NewResidentDutyService(
		&userRepositoryStub{
			user:          resident,
			teamResidents: []*user.User{teamMember},
		},
		&teamRepositoryStub{
			teamsByGroup: map[uuid.UUID][]*structure.Team{
				groupID: {team},
			},
		},
		&groupRepositoryStub{
			groups: []*structure.Group{group},
		},
		&dormitoryRepositoryStub{
			dormitory: dormitory,
		},
		&dutyRepositoryStub{
			activeByTeam: map[uuid.UUID]*duty.Duty{
				teamID: activeDuty,
			},
		},
		&taskDefinitionRepositoryStub{},
		&areaRepositoryStub{},
		nil,
	)
	service.now = func() time.Time { return now }

	response, err := service.GetCurrentDuty(context.Background(), residentID, nil)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.HasActiveDuty)
	assert.True(t, response.ReadOnly)
	assert.False(t, response.ShowGroupSelect)
	assert.Equal(t, groupID.String(), response.SelectedGroupID)
	assert.Equal(t, "Group A", response.Group)
	assert.Equal(t, "Глава команды не назначен", response.Team)
	assert.Equal(t, []string{"all", "team"}, response.VisibleTabs)
	assert.Equal(t, "На этой неделе дежурит команда Глава команды не назначен", response.NoticeMessage)
	assert.Empty(t, response.Tasks)
	assert.Len(t, response.TeamMembers, 1)
	assert.Equal(t, teamMemberID.String(), response.TeamMembers[0].ID)
	assert.Equal(t, "Petr Petrov", response.TeamMembers[0].Name)
}
