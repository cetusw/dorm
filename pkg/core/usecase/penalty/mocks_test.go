package penalty

import (
	"context"

	penaltydomain "dorm/pkg/core/domain/penalty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"
	queryports "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type penaltyRepositoryMock struct{ mock.Mock }

func (m *penaltyRepositoryMock) Create(c context.Context, p *penaltydomain.Penalty) error {
	return m.Called(c, p).Error(0)
}

func (m *penaltyRepositoryMock) CreateResolve(c context.Context, p *penaltydomain.Penalty) error {
	return m.Called(c, p).Error(0)
}

func (m *penaltyRepositoryMock) FindByID(c context.Context, id uuid.UUID) (*penaltydomain.Penalty, error) {
	a := m.Called(c, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*penaltydomain.Penalty), a.Error(1)
}

func (m *penaltyRepositoryMock) Update(c context.Context, p *penaltydomain.Penalty) error {
	return m.Called(c, p).Error(0)
}

func (m *penaltyRepositoryMock) Delete(c context.Context, id uuid.UUID) error {
	return m.Called(c, id).Error(0)
}

type penaltyUserRepositoryMock struct{ mock.Mock }

func (m *penaltyUserRepositoryMock) Save(c context.Context, u *user.User) error {
	return m.Called(c, u).Error(0)
}

func (m *penaltyUserRepositoryMock) FindAll(c context.Context) ([]*user.User, error) {
	a := m.Called(c)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]*user.User), a.Error(1)
}

func (m *penaltyUserRepositoryMock) FindByID(c context.Context, id uuid.UUID) (*user.User, error) {
	a := m.Called(c, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*user.User), a.Error(1)
}

func (m *penaltyUserRepositoryMock) FindByLogin(c context.Context, s string) (*user.User, error) {
	a := m.Called(c, s)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*user.User), a.Error(1)
}

func (m *penaltyUserRepositoryMock) FindByTeamID(c context.Context, id uuid.UUID) ([]*user.User, error) {
	a := m.Called(c, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]*user.User), a.Error(1)
}

func (m *penaltyUserRepositoryMock) FindByDormitoryID(c context.Context, id int64) ([]*user.User, error) {
	a := m.Called(c, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]*user.User), a.Error(1)
}

func (m *penaltyUserRepositoryMock) MoveUserToTeam(c context.Context, id uuid.UUID, team *uuid.UUID) error {
	return m.Called(c, id, team).Error(0)
}

func (m *penaltyUserRepositoryMock) SoftDelete(c context.Context, id uuid.UUID) error {
	return m.Called(c, id).Error(0)
}

type penaltyGroupRepositoryMock struct{ mock.Mock }

func (m *penaltyGroupRepositoryMock) FindAll(c context.Context) ([]*structure.Group, error) {
	a := m.Called(c)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]*structure.Group), a.Error(1)
}

func (m *penaltyGroupRepositoryMock) FindByID(c context.Context, id uuid.UUID) (*structure.Group, error) {
	a := m.Called(c, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*structure.Group), a.Error(1)
}

func (m *penaltyGroupRepositoryMock) FindByDormitoryID(c context.Context, id int64) ([]*structure.Group, error) {
	a := m.Called(c, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]*structure.Group), a.Error(1)
}

func (m *penaltyGroupRepositoryMock) Save(c context.Context, g *structure.Group) error {
	return m.Called(c, g).Error(0)
}

func (m *penaltyGroupRepositoryMock) Delete(c context.Context, id uuid.UUID) error {
	return m.Called(c, id).Error(0)
}

type penaltyDormitoryRepositoryMock struct{ mock.Mock }

func (m *penaltyDormitoryRepositoryMock) FindAll(c context.Context) ([]*structure.Dormitory, error) {
	a := m.Called(c)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]*structure.Dormitory), a.Error(1)
}

func (m *penaltyDormitoryRepositoryMock) FindByID(c context.Context, id int64) (*structure.Dormitory, error) {
	a := m.Called(c, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*structure.Dormitory), a.Error(1)
}

func (m *penaltyDormitoryRepositoryMock) ExistsByLeaderID(c context.Context, id uuid.UUID) (bool, error) {
	a := m.Called(c, id)
	return a.Bool(0), a.Error(1)
}

func (m *penaltyDormitoryRepositoryMock) Save(c context.Context, d *structure.Dormitory) error {
	return m.Called(c, d).Error(0)
}

func (m *penaltyDormitoryRepositoryMock) Delete(c context.Context, id int64) error {
	return m.Called(c, id).Error(0)
}

type penaltyQueryServiceMock struct{ mock.Mock }

func (m *penaltyQueryServiceMock) ListResidentsWithPenaltyBalance(c context.Context, s queryports.PenaltyScope) ([]dto.PenaltyResidentSummary, error) {
	a := m.Called(c, s)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]dto.PenaltyResidentSummary), a.Error(1)
}

func (m *penaltyQueryServiceMock) SearchEligibleResidents(c context.Context, s queryports.PenaltyScope, q string) ([]dto.PenaltyResidentOption, error) {
	a := m.Called(c, s, q)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]dto.PenaltyResidentOption), a.Error(1)
}

func (m *penaltyQueryServiceMock) ListPenaltyEntriesByUser(c context.Context, id uuid.UUID) ([]dto.PenaltyEntryItem, error) {
	a := m.Called(c, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]dto.PenaltyEntryItem), a.Error(1)
}
