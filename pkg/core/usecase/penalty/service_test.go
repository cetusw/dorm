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
	created          []*penaltydomain.Penalty
	resolved         []*penaltydomain.Penalty
	updated          *penaltydomain.Penalty
	foundByID        map[uuid.UUID]*penaltydomain.Penalty
	deletedID        uuid.UUID
	createErr        error
	createResolveErr error
	findErr          error
	updateErr        error
	deleteErr        error
}

func (s *penaltyRepositoryStub) Create(_ context.Context, item *penaltydomain.Penalty) error {
	if s.createErr != nil {
		return s.createErr
	}

	s.created = append(s.created, item)
	return nil
}

func (s *penaltyRepositoryStub) CreateResolve(_ context.Context, item *penaltydomain.Penalty) error {
	if s.createResolveErr != nil {
		return s.createResolveErr
	}

	s.resolved = append(s.resolved, item)
	return nil
}

func (s *penaltyRepositoryStub) FindByID(_ context.Context, id uuid.UUID) (*penaltydomain.Penalty, error) {
	if s.findErr != nil {
		return nil, s.findErr
	}

	if s.foundByID == nil {
		return nil, nil
	}

	return s.foundByID[id], nil
}

func (s *penaltyRepositoryStub) Update(_ context.Context, item *penaltydomain.Penalty) error {
	if s.updateErr != nil {
		return s.updateErr
	}

	s.updated = item
	return nil
}

func (s *penaltyRepositoryStub) Delete(_ context.Context, id uuid.UUID) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}

	s.deletedID = id
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
	residents []dto.PenaltyResidentSummary
	entries   []dto.PenaltyEntryItem
	options   []dto.PenaltyResidentOption
	lastScope queryports.PenaltyScope
}

func (s *penaltyQueryServiceStub) ListResidentsWithPenaltyBalance(
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
	_ = search
	return s.options, nil
}

func (s *penaltyQueryServiceStub) ListPenaltyEntriesByUser(
	_ context.Context,
	_ uuid.UUID,
) ([]dto.PenaltyEntryItem, error) {
	return s.entries, nil
}

func TestGetCurrentUserPenaltiesReturnsBalanceFromEntries(t *testing.T) {
	t.Parallel()

	currentUserID := uuid.New()
	dormitoryID := int64(7)
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, time.UTC)
	currentUser := user.RestoreUser(currentUserID, "resident", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)
	queryService := &penaltyQueryServiceStub{
		entries: []dto.PenaltyEntryItem{
			{ID: uuid.NewString(), Type: "issue", Reason: "Просрочил уборку", Weight: 1.0, CreatedAt: now.Format(time.RFC3339)},
			{ID: uuid.NewString(), Type: "resolve", Reason: "Закрыл долг", Weight: 0.5, CreatedAt: now.Format(time.RFC3339)},
		},
	}

	service := NewPenaltyService(
		&penaltyRepositoryStub{},
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
		}},
		&penaltyGroupRepositoryStub{},
		&penaltyDormitoryRepositoryStub{},
		queryService,
		time.UTC,
	)

	response, err := service.GetCurrentUserPenalties(context.Background(), currentUserID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 0.5, response.TotalWeight)
	assert.Equal(t, queryService.entries, response.Entries)
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
	penaltyRepo := &penaltyRepositoryStub{}

	service := NewPenaltyService(
		penaltyRepo,
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
			residentID:    resident,
		}},
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

	item, err := service.CreatePenalty(context.Background(), currentUserID, dto.CreatePenaltyRequest{
		UserID: residentID.String(),
		Reason: " late cleanup ",
		Weight: 1.24,
	})

	require.NoError(t, err)
	require.NotNil(t, item)
	require.Len(t, penaltyRepo.created, 1)
	assert.Equal(t, penaltydomain.EntryTypeIssue, penaltyRepo.created[0].Type())
	assert.Equal(t, "late cleanup", penaltyRepo.created[0].Reason())
	assert.Equal(t, 1.2, penaltyRepo.created[0].Weight())
	assert.Equal(t, now, penaltyRepo.created[0].CreatedAt())
	assert.Equal(t, "issue", item.Type)
}

func TestResolvePenaltyCreatesResolveEntry(t *testing.T) {
	t.Parallel()

	location := time.UTC
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, location)
	currentUserID := uuid.New()
	residentID := uuid.New()
	dormitoryID := int64(7)

	currentUser := user.RestoreUser(currentUserID, "leader", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)
	resident := user.RestoreUser(residentID, "resident", "hash", "Petr", nil, "Petrov", nil, nil, nil, &dormitoryID, now)
	penaltyRepo := &penaltyRepositoryStub{}

	service := NewPenaltyService(
		penaltyRepo,
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
			residentID:    resident,
		}},
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

	err := service.ResolvePenalty(context.Background(), currentUserID, dto.ResolvePenaltyRequest{
		UserID: residentID.String(),
		Reason: "Выполнено дополнительное задание",
		Weight: 0.5,
	})

	require.NoError(t, err)
	require.Len(t, penaltyRepo.resolved, 1)
	assert.Equal(t, penaltydomain.EntryTypeResolve, penaltyRepo.resolved[0].Type())
	assert.Equal(t, 0.5, penaltyRepo.resolved[0].Weight())
}

func TestResolvePenaltyPropagatesInsufficientBalance(t *testing.T) {
	t.Parallel()

	location := time.UTC
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, location)
	currentUserID := uuid.New()
	residentID := uuid.New()
	dormitoryID := int64(7)

	currentUser := user.RestoreUser(currentUserID, "leader", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)
	resident := user.RestoreUser(residentID, "resident", "hash", "Petr", nil, "Petrov", nil, nil, nil, &dormitoryID, now)

	service := NewPenaltyService(
		&penaltyRepositoryStub{createResolveErr: penaltydomain.ErrInsufficientPenaltyBalance},
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
			residentID:    resident,
		}},
		&penaltyGroupRepositoryStub{},
		&penaltyDormitoryRepositoryStub{
			all: []*structure.Dormitory{
				structure.RestoreDormitory(dormitoryID, "Dorm", &currentUserID, "Moscow", "st", "Lenina", "1"),
			},
		},
		&penaltyQueryServiceStub{},
		location,
	)

	err := service.ResolvePenalty(context.Background(), currentUserID, dto.ResolvePenaltyRequest{
		UserID: residentID.String(),
		Reason: "Попытка снять слишком много",
		Weight: 1.5,
	})

	assert.ErrorIs(t, err, penaltydomain.ErrInsufficientPenaltyBalance)
}

func TestResolvePenaltyRejectsInvalidWeight(t *testing.T) {
	t.Parallel()

	location := time.UTC
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, location)
	currentUserID := uuid.New()
	residentID := uuid.New()
	dormitoryID := int64(7)

	currentUser := user.RestoreUser(currentUserID, "leader", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)
	resident := user.RestoreUser(residentID, "resident", "hash", "Petr", nil, "Petrov", nil, nil, nil, &dormitoryID, now)

	service := NewPenaltyService(
		&penaltyRepositoryStub{},
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
			residentID:    resident,
		}},
		&penaltyGroupRepositoryStub{},
		&penaltyDormitoryRepositoryStub{
			all: []*structure.Dormitory{
				structure.RestoreDormitory(dormitoryID, "Dorm", &currentUserID, "Moscow", "st", "Lenina", "1"),
			},
		},
		&penaltyQueryServiceStub{},
		location,
	)

	err := service.ResolvePenalty(context.Background(), currentUserID, dto.ResolvePenaltyRequest{
		UserID: residentID.String(),
		Reason: "Некорректный вес",
		Weight: 0,
	})

	assert.ErrorIs(t, err, penaltydomain.ErrInvalidWeight)
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
		&penaltyRepositoryStub{},
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
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

	_, err := service.ListResidents(context.Background(), currentUserID)

	require.NoError(t, err)
	assert.Equal(t, []int64{dormitoryID}, queryService.lastScope.DormitoryIDs)
}

func TestGetResidentPenaltiesAllowsResidentInSameDormitory(t *testing.T) {
	t.Parallel()

	location := time.UTC
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, location)
	currentUserID := uuid.New()
	groupID := uuid.New()
	dormitoryID := int64(7)
	residentID := uuid.New()

	currentUser := user.RestoreUser(currentUserID, "leader", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)
	resident := user.RestoreUser(residentID, "resident", "hash", "Petr", nil, "Petrov", nil, nil, nil, &dormitoryID, now)
	queryService := &penaltyQueryServiceStub{
		entries: []dto.PenaltyEntryItem{
			{ID: uuid.NewString(), Type: "issue", Reason: "Просрочил уборку", Weight: 2, CreatedAt: now.Format(time.RFC3339)},
		},
	}

	service := NewPenaltyService(
		&penaltyRepositoryStub{},
		&penaltyUserRepositoryStub{byID: map[uuid.UUID]*user.User{
			currentUserID: currentUser,
			residentID:    resident,
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

	response, err := service.GetResidentPenalties(context.Background(), currentUserID, residentID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, residentID.String(), response.UserID)
	assert.Equal(t, 2.0, response.TotalWeight)
	assert.Equal(t, queryService.entries, response.Entries)
}

func TestPenaltyBalanceFromEntries(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		entries  []dto.PenaltyEntryItem
		expected float64
	}{
		{
			name:     "single issue",
			entries:  []dto.PenaltyEntryItem{{Type: "issue", Weight: 1.0}},
			expected: 1.0,
		},
		{
			name: "two issues",
			entries: []dto.PenaltyEntryItem{
				{Type: "issue", Weight: 1.0},
				{Type: "issue", Weight: 1.0},
			},
			expected: 2.0,
		},
		{
			name: "issue issue resolve",
			entries: []dto.PenaltyEntryItem{
				{Type: "issue", Weight: 1.0},
				{Type: "issue", Weight: 1.0},
				{Type: "resolve", Weight: 0.5},
			},
			expected: 1.5,
		},
		{
			name: "issue resolve to zero",
			entries: []dto.PenaltyEntryItem{
				{Type: "issue", Weight: 1.0},
				{Type: "resolve", Weight: 1.0},
			},
			expected: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, penaltyBalanceFromEntries(testCase.entries))
		})
	}
}
