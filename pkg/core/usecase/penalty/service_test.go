package penalty

import (
	"context"
	"errors"
	"testing"
	"time"

	penaltydomain "dorm/pkg/core/domain/penalty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"
	queryports "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var fixedPenaltyTime = time.Date(2026, time.August, 2, 12, 0, 0, 0, time.UTC)

type penaltyMocks struct {
	repo    *penaltyRepositoryMock
	users   *penaltyUserRepositoryMock
	groups  *penaltyGroupRepositoryMock
	dorms   *penaltyDormitoryRepositoryMock
	query   *penaltyQueryServiceMock
	service *Service
}

func newPenaltyMocks() penaltyMocks {
	m := penaltyMocks{
		repo:   new(penaltyRepositoryMock),
		users:  new(penaltyUserRepositoryMock),
		groups: new(penaltyGroupRepositoryMock),
		dorms:  new(penaltyDormitoryRepositoryMock),
		query:  new(penaltyQueryServiceMock),
	}
	m.service = NewPenaltyService(m.repo, m.users, m.groups, m.dorms, m.query, time.UTC)
	m.service.now = func() time.Time { return fixedPenaltyTime }
	return m
}

func (m penaltyMocks) verify(t *testing.T) {
	t.Helper()
	m.repo.AssertExpectations(t)
	m.users.AssertExpectations(t)
	m.groups.AssertExpectations(t)
	m.dorms.AssertExpectations(t)
	m.query.AssertExpectations(t)
}

func resident(
	id uuid.UUID,
	dorm *int64,
	middle *string,
) *user.User {
	return user.RestoreUser(
		id,
		"login",
		"hash",
		"Ivan",
		middle,
		"Ivanov",
		nil,
		nil,
		nil,
		dorm,
		fixedPenaltyTime,
	)
}

func dormitory(id int64, leader *uuid.UUID) *structure.Dormitory {
	return structure.RestoreDormitory(
		id,
		"Dorm",
		leader,
		"Moscow",
		"st",
		"Lenina",
		"1",
	)
}

func int64Ptr(id int64) *int64 { return &id }

func (m penaltyMocks) scope(
	ctx context.Context,
	leader *user.User,
	dormitories []*structure.Dormitory,
	groups []*structure.Group,
) {
	m.users.On("FindByID", ctx, leader.ID()).Return(leader, nil).Once()
	m.dorms.On("FindAll", ctx).Return(dormitories, nil).Once()
	if leader.DormitoryID() != nil {
		m.groups.On("FindByDormitoryID", ctx, *leader.DormitoryID()).Return(groups, nil).Once()
	}
}

func TestListResidents_DormitoryLeaderPassesScopeAndResponse(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	dormID := int64(7)
	leader := resident(id, &dormID, nil)

	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &id)}, nil)

	want := []dto.PenaltyResidentSummary{{UserID: uuid.NewString(), FullName: "Petrov Petr", TotalWeight: 1.5}}
	m.query.On(
		"ListResidentsWithPenaltyBalance",
		ctx,
		queryports.PenaltyScope{DormitoryIDs: []int64{dormID}},
	).Return(want, nil).Once()

	got, err := m.service.ListResidents(ctx, id)

	require.NoError(t, err)
	assert.Equal(t, want, got.Residents)
	m.verify(t)
}

func TestSearchResidents_GroupLeaderScopeAndSearch(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	dormID := int64(7)
	leader := resident(id, &dormID, nil)

	m := newPenaltyMocks()
	m.scope(ctx, leader, nil, []*structure.Group{structure.RestoreGroup(uuid.New(), &id, "A", dormID)})

	want := []dto.PenaltyResidentOption{{ID: uuid.NewString(), Name: "Petrov Petr"}}
	m.query.On(
		"SearchEligibleResidents",
		ctx,
		queryports.PenaltyScope{DormitoryIDs: []int64{dormID}}, "petr",
	).Return(want, nil).Once()

	got, err := m.service.SearchResidents(ctx, id, "petr")

	require.NoError(t, err)
	assert.Equal(t, want, got.Residents)
	m.verify(t)
}

func TestListResidents_CombinesScopesWithoutDuplicates(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	groupDormitoryID := int64(7)
	ownedDormitoryID := int64(8)
	leader := resident(id, &groupDormitoryID, nil)

	m := newPenaltyMocks()
	m.scope(
		ctx,
		leader,
		[]*structure.Dormitory{
			dormitory(groupDormitoryID, &id),
			dormitory(ownedDormitoryID, &id),
		},
		[]*structure.Group{
			structure.RestoreGroup(uuid.New(), &id, "A", groupDormitoryID),
		},
	)

	m.query.On(
		"ListResidentsWithPenaltyBalance",
		ctx,
		queryports.PenaltyScope{DormitoryIDs: []int64{groupDormitoryID, ownedDormitoryID}},
	).Return([]dto.PenaltyResidentSummary{}, nil).Once()
	_, err := m.service.ListResidents(ctx, id)

	require.NoError(t, err)
	m.verify(t)
}

func TestPenaltyScope_DeniesOrdinaryAndMissingUser(t *testing.T) {
	for _, tc := range []struct {
		name     string
		existing bool
	}{
		{name: "ordinary", existing: true},
		{name: "missing", existing: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			id := uuid.New()
			m := newPenaltyMocks()
			if tc.existing {
				u := resident(id, nil, nil)
				m.scope(ctx, u, nil, nil)
			} else {
				m.users.On("FindByID", ctx, id).Return(nil, nil).Once()
			}
			_, err := m.service.ListResidents(ctx, id)
			assert.ErrorIs(t, err, penaltydomain.ErrAccessDenied)
			m.query.AssertNotCalled(t, "ListResidentsWithPenaltyBalance", mock.Anything, mock.Anything)
			m.verify(t)
		})
	}
}

func TestListAndSearch_WrapQueryErrors(t *testing.T) {
	for _, tc := range []struct {
		name, method string
		call         func(*Service, context.Context, uuid.UUID) error
	}{
		{
			name:   "list",
			method: "ListResidentsWithPenaltyBalance",
			call: func(service *Service, ctx context.Context, userID uuid.UUID) error {
				_, err := service.ListResidents(ctx, userID)
				return err
			},
		},
		{
			name:   "search",
			method: "SearchEligibleResidents",
			call: func(service *Service, ctx context.Context, userID uuid.UUID) error {
				_, err := service.SearchResidents(ctx, userID, "ivan")
				return err
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			id := uuid.New()
			dormID := int64(7)
			u := resident(id, &dormID, nil)
			failure := errors.New("query failed")
			m := newPenaltyMocks()
			m.scope(ctx, u, []*structure.Dormitory{dormitory(dormID, &id)}, nil)
			scope := queryports.PenaltyScope{DormitoryIDs: []int64{dormID}}
			if tc.method == "ListResidentsWithPenaltyBalance" {
				m.query.On(tc.method, ctx, scope).Return(nil, failure).Once()
			} else {
				m.query.On(tc.method, ctx, scope, "ivan").Return(nil, failure).Once()
			}
			assert.ErrorIs(t, tc.call(m.service, ctx, id), failure)
			m.verify(t)
		})
	}
}

func TestGetResidentPenalties_ReturnsHistoryBalanceAndName(t *testing.T) {
	ctx := context.Background()
	leaderID, residentID := uuid.New(), uuid.New()
	dormID := int64(7)
	middle := " Sergeevich "
	leader, target := resident(leaderID, &dormID, nil), resident(residentID, &dormID, &middle)
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	m.users.On("FindByID", ctx, residentID).Return(target, nil).Once()
	entries := []dto.PenaltyEntryItem{
		{Type: "issue", Weight: 1},
		{Type: "issue", Weight: 2},
		{Type: "resolve", Weight: .5},
	}
	m.query.On("ListPenaltyEntriesByUser", ctx, residentID).Return(entries, nil).Once()
	got, err := m.service.GetResidentPenalties(ctx, leaderID, residentID)
	require.NoError(t, err)
	assert.Equal(t, "Ivanov Ivan Sergeevich", got.FullName)
	assert.Equal(t, 2.5, got.TotalWeight)
	assert.Equal(t, entries, got.Entries)
	m.verify(t)
}

func TestGetResidentPenalties_RejectsMissingAndOutOfScopeResident(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target *user.User
		want   error
	}{
		{name: "missing", target: nil, want: penaltydomain.ErrResidentNotFound},
		{name: "outside scope", target: resident(uuid.New(), int64Ptr(8), nil), want: penaltydomain.ErrAccessDenied},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			leaderID, targetID := uuid.New(), uuid.New()
			dormID := int64(7)
			leader := resident(leaderID, &dormID, nil)
			m := newPenaltyMocks()
			m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
			if tc.target != nil {
				targetID = tc.target.ID()
			}
			m.users.On("FindByID", ctx, targetID).Return(tc.target, nil).Once()
			_, err := m.service.GetResidentPenalties(ctx, leaderID, targetID)
			assert.ErrorIs(t, err, tc.want)
			m.query.AssertNotCalled(t, "ListPenaltyEntriesByUser", mock.Anything, mock.Anything)
			m.verify(t)
		})
	}
}

func TestGetCurrentUserPenalties_ReturnsOwnHistoryAndErrors(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	u := resident(id, nil, nil)
	m := newPenaltyMocks()
	m.users.On("FindByID", ctx, id).Return(u, nil).Once()
	entries := []dto.PenaltyEntryItem{
		{Type: "issue", Weight: 1},
		{Type: "resolve", Weight: .4},
	}
	m.query.On("ListPenaltyEntriesByUser", ctx, id).Return(entries, nil).Once()
	got, err := m.service.GetCurrentUserPenalties(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, .6, got.TotalWeight)
	m.verify(t)
	missing := newPenaltyMocks()
	missing.users.On("FindByID", ctx, id).Return(nil, nil).Once()
	_, err = missing.service.GetCurrentUserPenalties(ctx, id)
	assert.ErrorIs(t, err, penaltydomain.ErrAccessDenied)
	missing.verify(t)
	failure := errors.New("storage")
	broken := newPenaltyMocks()
	broken.users.On("FindByID", ctx, id).Return(u, nil).Once()
	broken.query.On("ListPenaltyEntriesByUser", ctx, id).Return(nil, failure).Once()
	_, err = broken.service.GetCurrentUserPenalties(ctx, id)
	assert.ErrorIs(t, err, failure)
	broken.verify(t)
}

func TestCreatePenalty_NormalizesInputCreatesIssueAtControlledTime(t *testing.T) {
	ctx := context.Background()
	leaderID, targetID := uuid.New(), uuid.New()
	dormID := int64(7)
	leader, target := resident(leaderID, &dormID, nil), resident(targetID, &dormID, nil)
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	m.users.On("FindByID", ctx, targetID).Return(target, nil).Once()
	m.repo.On("Create", ctx, mock.MatchedBy(func(p *penaltydomain.Penalty) bool {
		return p.UserID() == targetID &&
			p.Type() == penaltydomain.EntryTypeIssue &&
			p.Reason() == "late cleanup" &&
			p.Weight() == 1.2 &&
			p.CreatedAt().Equal(fixedPenaltyTime)
	})).Return(nil).Once()

	got, err := m.service.CreatePenalty(ctx, leaderID, dto.CreatePenaltyRequest{
		UserID: " " + targetID.String() + " ",
		Reason: " late cleanup ",
		Weight: 1.24,
	})
	require.NoError(t, err)
	assert.Equal(t, "issue", got.Type)
	assert.Equal(t, 1.2, got.Weight)
	assert.Equal(t, fixedPenaltyTime.Format(time.RFC3339), got.CreatedAt)
	m.verify(t)
}

func TestCreatePenalty_RejectsInvalidInputsAndRepositoryError(t *testing.T) {
	ctx := context.Background()
	leaderID, targetID := uuid.New(), uuid.New()
	dormID := int64(7)
	leader, target := resident(leaderID, &dormID, nil), resident(targetID, &dormID, nil)
	for _, tc := range []struct {
		name string
		r    dto.CreatePenaltyRequest
		want error
		load bool
	}{
		{
			name: "bad UUID",
			r: dto.CreatePenaltyRequest{
				UserID: "bad",
				Weight: 1,
				Reason: "reason",
			},
			want: penaltydomain.ErrInvalidResidentID,
		},
		{
			name: "bad weight",
			r: dto.CreatePenaltyRequest{
				UserID: targetID.String(),
				Weight: 0,
				Reason: "reason",
			},
			want: penaltydomain.ErrInvalidWeight,
			load: true,
		},
		{
			name: "empty reason",
			r: dto.CreatePenaltyRequest{
				UserID: targetID.String(),
				Weight: 1,
				Reason: " ",
			},
			want: penaltydomain.ErrInvalidReason,
			load: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newPenaltyMocks()
			m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
			if tc.load {
				m.users.On("FindByID", ctx, targetID).Return(target, nil).Once()
			}
			_, err := m.service.CreatePenalty(ctx, leaderID, tc.r)
			assert.ErrorIs(t, err, tc.want)
			m.repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			m.verify(t)
		})
	}
	failure := errors.New("write failed")
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	m.users.On("FindByID", ctx, targetID).Return(target, nil).Once()
	m.repo.On("Create", ctx, mock.Anything).Return(failure).Once()
	_, err := m.service.CreatePenalty(ctx, leaderID, dto.CreatePenaltyRequest{
		UserID: targetID.String(),
		Weight: 1,
		Reason: "reason",
	})
	assert.ErrorIs(t, err, failure)
	m.verify(t)
}

func TestResolvePenalty_NormalizesAndUsesAtomicRepository(t *testing.T) {
	ctx := context.Background()
	leaderID, targetID := uuid.New(), uuid.New()
	dormID := int64(7)
	leader, target := resident(leaderID, &dormID, nil), resident(targetID, &dormID, nil)
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	m.users.On("FindByID", ctx, targetID).Return(target, nil).Once()
	m.repo.On("CreateResolve", ctx, mock.MatchedBy(func(p *penaltydomain.Penalty) bool {
		return p.Type() == penaltydomain.EntryTypeResolve &&
			p.Reason() == "extra task" &&
			p.Weight() == .5 &&
			p.CreatedAt().Equal(fixedPenaltyTime)
	})).Return(nil).Once()

	err := m.service.ResolvePenalty(ctx, leaderID, dto.ResolvePenaltyRequest{
		UserID: targetID.String(),
		Reason: " extra task ",
		Weight: .54,
	})
	require.NoError(t, err)
	m.repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	m.verify(t)
}

func TestResolvePenalty_PropagatesRepositoryBalanceAndReservationConflicts(t *testing.T) {
	for _, want := range []error{
		penaltydomain.ErrInsufficientPenaltyBalance,
		penaltydomain.ErrPenaltyBalanceReserved,
	} {
		t.Run(want.Error(), func(t *testing.T) {
			ctx := context.Background()
			leaderID, targetID := uuid.New(), uuid.New()
			dormID := int64(7)
			leader, target := resident(leaderID, &dormID, nil), resident(targetID, &dormID, nil)
			m := newPenaltyMocks()
			m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
			m.users.On("FindByID", ctx, targetID).Return(target, nil).Once()
			m.repo.On("CreateResolve", ctx, mock.Anything).Return(want).Once()
			err := m.service.ResolvePenalty(ctx, leaderID, dto.ResolvePenaltyRequest{
				UserID: targetID.String(),
				Weight: 1,
				Reason: "reason",
			})
			assert.ErrorIs(t, err, want)
			m.verify(t)
		})
	}
}

func TestDeletePenaltyEntry_DeletesAccessibleEntry(t *testing.T) {
	ctx := context.Background()
	leaderID, targetID, entryID := uuid.New(), uuid.New(), uuid.New()
	dormID := int64(7)
	leader, target := resident(leaderID, &dormID, nil), resident(targetID, &dormID, nil)
	entry, err := penaltydomain.RestorePenalty(
		entryID,
		targetID,
		penaltydomain.EntryTypeIssue,
		1,
		"reason",
		fixedPenaltyTime,
	)
	require.NoError(t, err)
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	m.repo.On("FindByID", ctx, entryID).Return(entry, nil).Once()
	m.users.On("FindByID", ctx, targetID).Return(target, nil).Once()
	m.repo.On("Delete", ctx, entryID).Return(nil).Once()
	require.NoError(t, m.service.DeletePenaltyEntry(ctx, leaderID, entryID))
	m.verify(t)
}

func TestDeletePenaltyEntry_HandlesMissingForbiddenAndRepositoryConflict(t *testing.T) {
	ctx := context.Background()
	leaderID, targetID, entryID := uuid.New(), uuid.New(), uuid.New()
	dormID := int64(7)
	leader := resident(leaderID, &dormID, nil)
	entry, err := penaltydomain.RestorePenalty(
		entryID,
		targetID,
		penaltydomain.EntryTypeIssue,
		1,
		"reason",
		fixedPenaltyTime,
	)
	require.NoError(t, err)
	for _, tc := range []struct {
		name string
		e    *penaltydomain.Penalty
		u    *user.User
		del  error
		want error
	}{
		{
			name: "missing",
			want: penaltydomain.ErrPenaltyEntryNotFound,
		},
		{
			name: "forbidden",
			e:    entry,
			u:    resident(targetID, int64Ptr(8), nil),
			want: penaltydomain.ErrAccessDenied,
		},
		{
			name: "reserved",
			e:    entry,
			u:    resident(targetID, &dormID, nil),
			del:  penaltydomain.ErrPenaltyBalanceReserved,
			want: penaltydomain.ErrPenaltyBalanceReserved,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newPenaltyMocks()
			m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
			m.repo.On("FindByID", ctx, entryID).Return(tc.e, nil).Once()
			if tc.e != nil {
				m.users.On("FindByID", ctx, targetID).Return(tc.u, nil).Once()
			}
			if tc.del != nil {
				m.repo.On("Delete", ctx, entryID).Return(tc.del).Once()
			}
			assert.ErrorIs(t, m.service.DeletePenaltyEntry(ctx, leaderID, entryID), tc.want)
			m.verify(t)
		})
	}
}

func TestUpdatePenaltyEntry_PreservesTypeAndTimeAndPropagatesInvariantError(t *testing.T) {
	ctx := context.Background()
	leaderID, targetID, entryID := uuid.New(), uuid.New(), uuid.New()
	dormID := int64(7)
	leader, target := resident(leaderID, &dormID, nil), resident(targetID, &dormID, nil)
	created := fixedPenaltyTime.Add(-time.Hour)
	entry, err := penaltydomain.RestorePenalty(
		entryID,
		targetID,
		penaltydomain.EntryTypeResolve,
		1,
		"old",
		created,
	)
	require.NoError(t, err)
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	m.repo.On("FindByID", ctx, entryID).Return(entry, nil).Once()
	m.users.On("FindByID", ctx, targetID).Return(target, nil).Once()
	m.repo.On("Update", ctx, mock.MatchedBy(func(p *penaltydomain.Penalty) bool {
		return p.ID() == entryID &&
			p.UserID() == targetID &&
			p.Type() == penaltydomain.EntryTypeResolve &&
			p.CreatedAt().Equal(created) &&
			p.Reason() == "new" &&
			p.Weight() == 1.3
	})).Return(nil).Once()

	got, err := m.service.UpdatePenaltyEntry(ctx, leaderID, entryID, dto.UpdatePenaltyEntryRequest{
		Reason: " new ",
		Weight: 1.26,
	})
	require.NoError(t, err)
	assert.Equal(t, "resolve", got.Type)
	assert.Equal(t, created.Format(time.RFC3339), got.CreatedAt)
	m.verify(t)
	conflict := newPenaltyMocks()
	conflict.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	conflict.repo.On("FindByID", ctx, entryID).Return(entry, nil).Once()
	conflict.users.On("FindByID", ctx, targetID).Return(target, nil).Once()
	conflict.repo.On("Update", ctx, mock.Anything).Return(penaltydomain.ErrNegativePenaltyHistory).Once()
	_, err = conflict.service.UpdatePenaltyEntry(ctx, leaderID, entryID, dto.UpdatePenaltyEntryRequest{
		Reason: "new",
		Weight: 2,
	})
	assert.ErrorIs(t, err, penaltydomain.ErrNegativePenaltyHistory)
	conflict.verify(t)
}

func TestUpdatePenaltyEntry_ReportsMissingEntry(t *testing.T) {
	ctx := context.Background()
	leaderID, entryID := uuid.New(), uuid.New()
	dormID := int64(7)
	leader := resident(leaderID, &dormID, nil)
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	m.repo.On("FindByID", ctx, entryID).Return(nil, nil).Once()
	_, err := m.service.UpdatePenaltyEntry(ctx, leaderID, entryID, dto.UpdatePenaltyEntryRequest{Reason: "new", Weight: 1})
	assert.ErrorIs(t, err, penaltydomain.ErrPenaltyEntryNotFound)
	m.verify(t)
}

func TestScope_WrapsCollaboratorFailures(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	dormID := int64(7)
	u := resident(id, &dormID, nil)
	failure := errors.New("storage unavailable")

	currentUserFailure := newPenaltyMocks()
	currentUserFailure.users.On("FindByID", ctx, id).Return(nil, failure).Once()
	_, err := currentUserFailure.service.ListResidents(ctx, id)
	assert.ErrorIs(t, err, failure)
	currentUserFailure.verify(t)

	dormitoryFailure := newPenaltyMocks()
	dormitoryFailure.users.On("FindByID", ctx, id).Return(u, nil).Once()
	dormitoryFailure.dorms.On("FindAll", ctx).Return(nil, failure).Once()
	_, err = dormitoryFailure.service.ListResidents(ctx, id)
	assert.ErrorIs(t, err, failure)
	dormitoryFailure.verify(t)

	groupFailure := newPenaltyMocks()
	groupFailure.users.On("FindByID", ctx, id).Return(u, nil).Once()
	groupFailure.dorms.On("FindAll", ctx).Return(nil, nil).Once()
	groupFailure.groups.On("FindByDormitoryID", ctx, dormID).Return(nil, failure).Once()
	_, err = groupFailure.service.ListResidents(ctx, id)
	assert.ErrorIs(t, err, failure)
	groupFailure.verify(t)
}

func TestResidentAndEntryLookupErrorsAreWrapped(t *testing.T) {
	ctx := context.Background()
	leaderID, residentID, entryID := uuid.New(), uuid.New(), uuid.New()
	dormID := int64(7)
	leader := resident(leaderID, &dormID, nil)
	failure := errors.New("storage unavailable")

	residentFailure := newPenaltyMocks()
	residentFailure.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	residentFailure.users.On("FindByID", ctx, residentID).Return(nil, failure).Once()
	_, err := residentFailure.service.GetResidentPenalties(ctx, leaderID, residentID)
	assert.ErrorIs(t, err, failure)
	residentFailure.verify(t)

	deleteFailure := newPenaltyMocks()
	deleteFailure.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	deleteFailure.repo.On("FindByID", ctx, entryID).Return(nil, failure).Once()
	err = deleteFailure.service.DeletePenaltyEntry(ctx, leaderID, entryID)
	assert.ErrorIs(t, err, failure)
	deleteFailure.verify(t)

	updateFailure := newPenaltyMocks()
	updateFailure.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	updateFailure.repo.On("FindByID", ctx, entryID).Return(nil, failure).Once()
	_, err = updateFailure.service.UpdatePenaltyEntry(ctx, leaderID, entryID, dto.UpdatePenaltyEntryRequest{Reason: "reason", Weight: 1})
	assert.ErrorIs(t, err, failure)
	updateFailure.verify(t)
}

func TestGetResidentPenalties_WrapsHistoryQueryError(t *testing.T) {
	ctx := context.Background()
	leaderID, residentID := uuid.New(), uuid.New()
	dormID := int64(7)
	leader, target := resident(leaderID, &dormID, nil), resident(residentID, &dormID, nil)
	failure := errors.New("query unavailable")
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	m.users.On("FindByID", ctx, residentID).Return(target, nil).Once()
	m.query.On("ListPenaltyEntriesByUser", ctx, residentID).Return(nil, failure).Once()
	_, err := m.service.GetResidentPenalties(ctx, leaderID, residentID)
	assert.ErrorIs(t, err, failure)
	m.verify(t)
}

func TestResolvePenalty_RejectsInvalidInputBeforeWrite(t *testing.T) {
	ctx := context.Background()
	leaderID := uuid.New()
	dormID := int64(7)
	leader := resident(leaderID, &dormID, nil)
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	err := m.service.ResolvePenalty(ctx, leaderID, dto.ResolvePenaltyRequest{UserID: "bad", Reason: "reason", Weight: 1})
	assert.ErrorIs(t, err, penaltydomain.ErrInvalidResidentID)
	m.repo.AssertNotCalled(t, "CreateResolve", mock.Anything, mock.Anything)
	m.verify(t)
}

func TestCurrentUserAndResidentValidationErrors(t *testing.T) {
	ctx := context.Background()
	currentID := uuid.New()
	failure := errors.New("user storage unavailable")
	current := newPenaltyMocks()
	current.users.On("FindByID", ctx, currentID).Return(nil, failure).Once()
	_, err := current.service.GetCurrentUserPenalties(ctx, currentID)
	assert.ErrorIs(t, err, failure)
	current.verify(t)

	leaderID, residentID := uuid.New(), uuid.New()
	dormID := int64(7)
	leader := resident(leaderID, &dormID, nil)
	withoutDormitory := resident(residentID, nil, nil)
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	m.users.On("FindByID", ctx, residentID).Return(withoutDormitory, nil).Once()
	_, err = m.service.GetResidentPenalties(ctx, leaderID, residentID)
	assert.ErrorIs(t, err, penaltydomain.ErrAccessDenied)
	m.verify(t)
}

func TestCreateAndResolve_RejectMissingResidentAndInvalidWeight(t *testing.T) {
	ctx := context.Background()
	leaderID, residentID := uuid.New(), uuid.New()
	dormID := int64(7)
	leader := resident(leaderID, &dormID, nil)

	create := newPenaltyMocks()
	create.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	create.users.On("FindByID", ctx, residentID).Return(nil, nil).Once()
	_, err := create.service.CreatePenalty(ctx, leaderID, dto.CreatePenaltyRequest{UserID: residentID.String(), Reason: "reason", Weight: 1})
	assert.ErrorIs(t, err, penaltydomain.ErrResidentNotFound)
	create.verify(t)

	resolve := newPenaltyMocks()
	resolve.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	resolve.users.On("FindByID", ctx, residentID).Return(resident(residentID, &dormID, nil), nil).Once()
	err = resolve.service.ResolvePenalty(ctx, leaderID, dto.ResolvePenaltyRequest{UserID: residentID.String(), Reason: "reason", Weight: 0})
	assert.ErrorIs(t, err, penaltydomain.ErrInvalidWeight)
	resolve.repo.AssertNotCalled(t, "CreateResolve", mock.Anything, mock.Anything)
	resolve.verify(t)
}

func TestUpdatePenaltyEntry_RejectsInvalidMutableValues(t *testing.T) {
	ctx := context.Background()
	leaderID, residentID, entryID := uuid.New(), uuid.New(), uuid.New()
	dormID := int64(7)
	leader, target := resident(leaderID, &dormID, nil), resident(residentID, &dormID, nil)
	entry, err := penaltydomain.RestorePenalty(entryID, residentID, penaltydomain.EntryTypeIssue, 1, "old", fixedPenaltyTime)
	require.NoError(t, err)
	m := newPenaltyMocks()
	m.scope(ctx, leader, []*structure.Dormitory{dormitory(dormID, &leaderID)}, nil)
	m.repo.On("FindByID", ctx, entryID).Return(entry, nil).Once()
	m.users.On("FindByID", ctx, residentID).Return(target, nil).Once()
	_, err = m.service.UpdatePenaltyEntry(ctx, leaderID, entryID, dto.UpdatePenaltyEntryRequest{Reason: " ", Weight: 1})
	assert.ErrorIs(t, err, penaltydomain.ErrInvalidReason)
	m.repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	m.verify(t)
}
