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
func (s *teamRepositoryStub) Save(context.Context, *structure.Team) error             { return nil }
func (s *teamRepositoryStub) CreateWithLeader(context.Context, *structure.Team) error { return nil }
func (s *teamRepositoryStub) SoftDelete(context.Context, uuid.UUID, time.Time) error  { return nil }
func (s *teamRepositoryStub) ReplaceLeaderAndRemoveMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return nil
}

type groupRepositoryStub struct {
	groups            []*structure.Group
	groupsByDormitory map[int64][]*structure.Group
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
func (s *groupRepositoryStub) FindByDormitoryID(_ context.Context, dormitoryID int64) ([]*structure.Group, error) {
	if len(s.groupsByDormitory) > 0 {
		return s.groupsByDormitory[dormitoryID], nil
	}
	return s.groups, nil
}
func (s *groupRepositoryStub) Save(context.Context, *structure.Group) error { return nil }
func (s *groupRepositoryStub) Delete(context.Context, uuid.UUID) error      { return nil }

type dormitoryRepositoryStub struct {
	dormitory       *structure.Dormitory
	dormitoriesByID map[int64]*structure.Dormitory
}

func (s *dormitoryRepositoryStub) FindAll(context.Context) ([]*structure.Dormitory, error) {
	return nil, nil
}
func (s *dormitoryRepositoryStub) FindByID(_ context.Context, dormitoryID int64) (*structure.Dormitory, error) {
	if len(s.dormitoriesByID) > 0 {
		return s.dormitoriesByID[dormitoryID], nil
	}
	return s.dormitory, nil
}
func (s *dormitoryRepositoryStub) ExistsByLeaderID(_ context.Context, leaderID uuid.UUID) (bool, error) {
	if len(s.dormitoriesByID) > 0 {
		for _, dormitory := range s.dormitoriesByID {
			if dormitory.LeaderID() != nil && *dormitory.LeaderID() == leaderID {
				return true, nil
			}
		}
		return false, nil
	}
	if s.dormitory == nil || s.dormitory.LeaderID() == nil {
		return false, nil
	}
	return *s.dormitory.LeaderID() == leaderID, nil
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
func (s *dutyRepositoryStub) FindActiveByGroupID(context.Context, uuid.UUID, time.Time) (*duty.Duty, error) {
	return nil, nil
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
func (s *dutyRepositoryStub) FindGroupIDByDutyID(context.Context, uuid.UUID) (*uuid.UUID, error) {
	return nil, nil
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
func (s *dutyRepositoryStub) ReassignTeamAndResetTasks(context.Context, uuid.UUID, uuid.UUID, time.Time) error {
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
	team := structure.RestoreTeam(teamID, groupID, teamMemberID, "#00AAFF", 1)
	activeDuty := duty.RestoreDuty(
		uuid.New(),
		teamID,
		now.Add(-24*time.Hour),
		now.Add(24*time.Hour),
		1,
		nil,
	)
	activeDuty.SetLeaderID(&teamMemberID)
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

	response, err := service.GetCurrentDuty(context.Background(), residentID, nil, nil)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.HasActiveDuty)
	assert.True(t, response.ReadOnly)
	assert.False(t, response.ShowGroupSelect)
	assert.Equal(t, groupID.String(), response.SelectedGroupID)
	assert.Equal(t, "Group A", response.Group)
	assert.Equal(t, "Ivan Ivanov", response.Team)
	assert.Equal(t, []string{"all", "team"}, response.VisibleTabs)
	assert.Equal(t, "На этой неделе ответственный за дежурство — Ivan Ivanov", response.NoticeMessage)
	assert.Empty(t, response.Tasks)
	assert.Len(t, response.TeamMembers, 1)
	assert.Equal(t, teamMemberID.String(), response.TeamMembers[0].ID)
	assert.Equal(t, "Petr Petrov", response.TeamMembers[0].Name)
}

func TestGetCurrentDuty_UsesRequestedDormitoryForDormitoryLeader(t *testing.T) {
	t.Parallel()

	leaderID := uuid.New()
	otherMemberID := uuid.New()
	groupID := uuid.New()
	teamID := uuid.New()
	homeDormitoryID := int64(7)
	selectedDormitoryID := int64(11)
	now := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)

	leader := user.RestoreUser(
		leaderID,
		"leader",
		"hash",
		"Ivan",
		nil,
		"Ivanov",
		nil,
		nil,
		nil,
		&homeDormitoryID,
		now,
	)
	otherMember := user.RestoreUser(
		otherMemberID,
		"member",
		"hash",
		"Petr",
		nil,
		"Petrov",
		&teamID,
		nil,
		nil,
		&selectedDormitoryID,
		now,
	)

	myGroupID := uuid.New()
	myGroup := structure.RestoreGroup(myGroupID, &leaderID, "My Group", homeDormitoryID)
	selectedGroup := structure.RestoreGroup(groupID, nil, "Observed Group", selectedDormitoryID)
	team := structure.RestoreTeam(teamID, groupID, leaderID, "#00AAFF", 1)
	activeDuty := duty.RestoreDuty(
		uuid.New(),
		teamID,
		now.Add(-24*time.Hour),
		now.Add(24*time.Hour),
		1,
		nil,
	)
	activeDuty.SetLeaderID(&leaderID)
	homeDormitory := structure.RestoreDormitory(homeDormitoryID, "Home Dorm", &leaderID, "Moscow", "st", "Lenina", "1")
	selectedDormitory := structure.RestoreDormitory(selectedDormitoryID, "Observed Dorm", nil, "Moscow", "st", "Tverskaya", "2")

	service := NewResidentDutyService(
		&userRepositoryStub{
			user:          leader,
			teamResidents: []*user.User{otherMember},
		},
		&teamRepositoryStub{
			teamsByGroup: map[uuid.UUID][]*structure.Team{
				groupID:   {team},
				myGroupID: {},
			},
		},
		&groupRepositoryStub{
			groupsByDormitory: map[int64][]*structure.Group{
				homeDormitoryID:     {myGroup},
				selectedDormitoryID: {selectedGroup},
			},
		},
		&dormitoryRepositoryStub{
			dormitoriesByID: map[int64]*structure.Dormitory{
				homeDormitoryID:     homeDormitory,
				selectedDormitoryID: selectedDormitory,
			},
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

	response, err := service.GetCurrentDuty(context.Background(), leaderID, nil, &selectedDormitoryID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, selectedDormitoryID, response.DormitoryID)
	require.NotNil(t, response.MyDormitoryID)
	assert.Equal(t, homeDormitoryID, *response.MyDormitoryID)
	assert.Equal(t, selectedGroup.ID().String(), response.SelectedGroupID)
	assert.Len(t, response.Groups, 1)
	assert.Equal(t, selectedGroup.ID().String(), response.Groups[0].ID)
	require.NotNil(t, response.MyGroup)
	assert.Equal(t, myGroup.ID().String(), response.MyGroup.ID)
	assert.Equal(t, myGroup.Name(), response.MyGroup.Name)
}
