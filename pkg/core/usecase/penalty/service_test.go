package penalty

import (
	"context"
	"testing"
	"time"

	penaltydomain "dorm/pkg/core/domain/penalty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"
	queryports "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type penaltyRepositoryStub struct {
	byID  map[uuid.UUID]*penaltydomain.Penalty
	saved []*penaltydomain.Penalty
}

func (s *penaltyRepositoryStub) FindByID(_ context.Context, id uuid.UUID) (*penaltydomain.Penalty, error) {
	return s.byID[id], nil
}

func (s *penaltyRepositoryStub) Save(_ context.Context, item *penaltydomain.Penalty) error {
	s.byID[item.ID()] = item
	s.saved = append(s.saved, item)
	return nil
}

type penaltyUserRepositoryStub struct {
	byID map[uuid.UUID]*user.User
}

func (s *penaltyUserRepositoryStub) Save(context.Context, *user.User) error { return nil }
func (s *penaltyUserRepositoryStub) FindAll(context.Context) ([]*user.User, error) {
	return nil, nil
}
func (s *penaltyUserRepositoryStub) FindByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	return s.byID[id], nil
}
func (s *penaltyUserRepositoryStub) FindByLogin(context.Context, string) (*user.User, error) {
	return nil, nil
}
func (s *penaltyUserRepositoryStub) FindByTeamID(context.Context, uuid.UUID) ([]*user.User, error) {
	return nil, nil
}
func (s *penaltyUserRepositoryStub) FindByDormitoryID(context.Context, int64) ([]*user.User, error) {
	return nil, nil
}
func (s *penaltyUserRepositoryStub) MoveUserToTeam(context.Context, uuid.UUID, *uuid.UUID) error {
	return nil
}
func (s *penaltyUserRepositoryStub) SoftDelete(context.Context, uuid.UUID) error { return nil }

type penaltyTeamRepositoryStub struct {
	byID map[uuid.UUID]*structure.Team
}

func (s *penaltyTeamRepositoryStub) FindByID(_ context.Context, id uuid.UUID) (*structure.Team, error) {
	return s.byID[id], nil
}
func (s *penaltyTeamRepositoryStub) FindByGroupID(context.Context, uuid.UUID) ([]*structure.Team, error) {
	return nil, nil
}
func (s *penaltyTeamRepositoryStub) UpdateRotationPositions(context.Context, uuid.UUID, []uuid.UUID) error {
	return nil
}
func (s *penaltyTeamRepositoryStub) Save(context.Context, *structure.Team) error { return nil }
func (s *penaltyTeamRepositoryStub) Delete(context.Context, uuid.UUID) error     { return nil }

type penaltyGroupRepositoryStub struct {
	byDormitory map[int64][]*structure.Group
}

func (s *penaltyGroupRepositoryStub) FindAll(context.Context) ([]*structure.Group, error) {
	return nil, nil
}
func (s *penaltyGroupRepositoryStub) FindByID(context.Context, uuid.UUID) (*structure.Group, error) {
	return nil, nil
}
func (s *penaltyGroupRepositoryStub) FindByDormitoryID(_ context.Context, dormitoryID int64) ([]*structure.Group, error) {
	return s.byDormitory[dormitoryID], nil
}
func (s *penaltyGroupRepositoryStub) Save(context.Context, *structure.Group) error { return nil }
func (s *penaltyGroupRepositoryStub) Delete(context.Context, uuid.UUID) error      { return nil }

type penaltyDormitoryRepositoryStub struct {
	all []*structure.Dormitory
}

func (s *penaltyDormitoryRepositoryStub) FindAll(context.Context) ([]*structure.Dormitory, error) {
	return s.all, nil
}
func (s *penaltyDormitoryRepositoryStub) FindByID(context.Context, int64) (*structure.Dormitory, error) {
	return nil, nil
}
func (s *penaltyDormitoryRepositoryStub) ExistsByLeaderID(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (s *penaltyDormitoryRepositoryStub) Save(context.Context, *structure.Dormitory) error {
	return nil
}
func (s *penaltyDormitoryRepositoryStub) Delete(context.Context, int64) error { return nil }

type penaltyQueryServiceStub struct {
	residents  []dto.PenaltyResidentSummary
	residentBy []dto.PenaltyItem
	options    []dto.PenaltyResidentOption
	lastScope  queryports.PenaltyScope
	lastSearch string
}

func (s *penaltyQueryServiceStub) ListResidentsWithActivePenalties(
	_ context.Context,
	scope queryports.PenaltyScope,
) ([]dto.PenaltyResidentSummary, error) {
	s.lastScope = scope
	return s.residents, nil
}

func (s *penaltyQueryServiceStub) SearchEligibleResidents(
	_ context.Context,
	scope queryports.PenaltyScope,
	search string,
) ([]dto.PenaltyResidentOption, error) {
	s.lastScope = scope
	s.lastSearch = search
	return s.options, nil
}

func (s *penaltyQueryServiceStub) ListActivePenaltiesByUser(
	_ context.Context,
	_ uuid.UUID,
) ([]dto.PenaltyItem, error) {
	return s.residentBy, nil
}

func TestGetCurrentUserPenaltiesReturnsActiveItems(t *testing.T) {
	t.Parallel()

	currentUserID := uuid.New()
	dormitoryID := int64(7)
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, time.UTC)
	currentUser := user.RestoreUser(currentUserID, "resident", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)
	queryService := &penaltyQueryServiceStub{
		residentBy: []dto.PenaltyItem{
			{ID: uuid.NewString(), Reason: "Просрочил уборку", Weight: 2, IssuedOn: "2026-08-01"},
		},
	}

	service := NewPenaltyService(
		&penaltyRepositoryStub{byID: map[uuid.UUID]*penaltydomain.Penalty{}},
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
		}},
		&penaltyTeamRepositoryStub{byID: map[uuid.UUID]*structure.Team{}},
		&penaltyGroupRepositoryStub{},
		&penaltyDormitoryRepositoryStub{},
		queryService,
		time.UTC,
	)

	response, err := service.GetCurrentUserPenalties(context.Background(), currentUserID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, queryService.residentBy, response.Penalties)
}

func TestCreatePenaltyUsesDormitoryLeaderScopeAndNormalizesData(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("UTC+3", 3*60*60)
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, location)
	currentUserID := uuid.New()
	residentID := uuid.New()
	dormitoryID := int64(7)

	currentUser := user.RestoreUser(currentUserID, "leader", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)
	resident := user.RestoreUser(residentID, "resident", "hash", "Petr", nil, "Petrov", nil, nil, nil, &dormitoryID, now)
	penaltyRepo := &penaltyRepositoryStub{byID: map[uuid.UUID]*penaltydomain.Penalty{}}
	queryService := &penaltyQueryServiceStub{}

	service := NewPenaltyService(
		penaltyRepo,
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
			residentID:    resident,
		}},
		&penaltyTeamRepositoryStub{byID: map[uuid.UUID]*structure.Team{}},
		&penaltyGroupRepositoryStub{},
		&penaltyDormitoryRepositoryStub{
			all: []*structure.Dormitory{
				structure.RestoreDormitory(dormitoryID, "Dorm", &currentUserID, "Moscow", "st", "Lenina", "1"),
			},
		},
		queryService,
		location,
	)
	service.now = func() time.Time { return now }

	item, err := service.CreatePenalty(context.Background(), currentUserID, dto.CreatePenaltyRequest{
		UserID:   residentID.String(),
		Reason:   " late cleanup ",
		Weight:   1.24,
		IssuedOn: "2026-08-02",
	})

	require.NoError(t, err)
	require.NotNil(t, item)
	require.Len(t, penaltyRepo.saved, 1)
	assert.Equal(t, "late cleanup", penaltyRepo.saved[0].Reason())
	assert.Equal(t, 1.2, penaltyRepo.saved[0].Weight())
	assert.Equal(t, now, penaltyRepo.saved[0].CreatedAt())
	assert.Equal(t, "2026-08-02", item.IssuedOn)
}

func TestCreatePenaltyRejectsFutureDate(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("UTC+3", 3*60*60)
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, location)
	currentUserID := uuid.New()
	residentID := uuid.New()
	dormitoryID := int64(7)

	currentUser := user.RestoreUser(currentUserID, "leader", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)
	resident := user.RestoreUser(residentID, "resident", "hash", "Petr", nil, "Petrov", nil, nil, nil, &dormitoryID, now)

	service := NewPenaltyService(
		&penaltyRepositoryStub{byID: map[uuid.UUID]*penaltydomain.Penalty{}},
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
			residentID:    resident,
		}},
		&penaltyTeamRepositoryStub{byID: map[uuid.UUID]*structure.Team{}},
		&penaltyGroupRepositoryStub{},
		&penaltyDormitoryRepositoryStub{
			all: []*structure.Dormitory{
				structure.RestoreDormitory(dormitoryID, "Dorm", &currentUserID, "Moscow", "st", "Lenina", "1"),
			},
		},
		&penaltyQueryServiceStub{},
		location,
	)
	service.now = func() time.Time { return now }

	_, err := service.CreatePenalty(context.Background(), currentUserID, dto.CreatePenaltyRequest{
		UserID:   residentID.String(),
		Reason:   "reason",
		Weight:   1.0,
		IssuedOn: "2026-08-03",
	})

	assert.ErrorIs(t, err, penaltydomain.ErrIssuedOnInFuture)
}

func TestResolveScopeFallsBackToDormitoryScopeForGroupLeader(t *testing.T) {
	t.Parallel()

	location := time.UTC
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, location)
	currentUserID := uuid.New()
	groupID := uuid.New()
	dormitoryID := int64(7)

	currentUser := user.RestoreUser(currentUserID, "leader", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)
	queryService := &penaltyQueryServiceStub{
		residents: []dto.PenaltyResidentSummary{},
	}

	service := NewPenaltyService(
		&penaltyRepositoryStub{byID: map[uuid.UUID]*penaltydomain.Penalty{}},
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
		}},
		&penaltyTeamRepositoryStub{byID: map[uuid.UUID]*structure.Team{}},
		&penaltyGroupRepositoryStub{
			byDormitory: map[int64][]*structure.Group{
				dormitoryID: {
					structure.RestoreGroup(groupID, &currentUserID, "Group A", dormitoryID),
				},
			},
		},
		&penaltyDormitoryRepositoryStub{},
		queryService,
		location,
	)
	service.now = func() time.Time { return now }

	_, err := service.ListResidents(context.Background(), currentUserID)

	require.NoError(t, err)
	assert.Equal(t, []int64{dormitoryID}, queryService.lastScope.DormitoryIDs)
	assert.Empty(t, queryService.lastScope.GroupIDs)
}

func TestGetResidentPenaltiesAllowsResidentFromAnotherGroupInSameDormitory(t *testing.T) {
	t.Parallel()

	location := time.UTC
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, location)
	currentUserID := uuid.New()
	groupID := uuid.New()
	otherGroupID := uuid.New()
	dormitoryID := int64(7)
	teamID := uuid.New()
	residentID := uuid.New()

	currentUser := user.RestoreUser(currentUserID, "leader", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)
	resident := user.RestoreUser(residentID, "resident", "hash", "Petr", nil, "Petrov", &teamID, nil, nil, &dormitoryID, now)
	queryService := &penaltyQueryServiceStub{
		residentBy: []dto.PenaltyItem{
			{ID: uuid.NewString(), Reason: "Просрочил уборку", Weight: 2, IssuedOn: "2026-08-01"},
		},
	}

	service := NewPenaltyService(
		&penaltyRepositoryStub{byID: map[uuid.UUID]*penaltydomain.Penalty{}},
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
			residentID:    resident,
		}},
		&penaltyTeamRepositoryStub{byID: map[uuid.UUID]*structure.Team{
			teamID: structure.RestoreTeam(teamID, "Team", otherGroupID, nil, "#fff", 1),
		}},
		&penaltyGroupRepositoryStub{
			byDormitory: map[int64][]*structure.Group{
				dormitoryID: {
					structure.RestoreGroup(groupID, &currentUserID, "Group A", dormitoryID),
				},
			},
		},
		&penaltyDormitoryRepositoryStub{},
		queryService,
		location,
	)
	service.now = func() time.Time { return now }

	response, err := service.GetResidentPenalties(context.Background(), currentUserID, residentID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, residentID.String(), response.UserID)
	assert.Equal(t, queryService.residentBy, response.Penalties)
}
