package cleaning

import (
	"context"
	"dorm/pkg/core/ports"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
)

type MockDutyRepo struct{ mock.Mock }

func (m *MockDutyRepo) Save(ctx context.Context, d *duty.Duty) error {
	return m.Called(ctx, d).Error(0)
}
func (m *MockDutyRepo) FindCurrentByTeamID(ctx context.Context, id uuid.UUID) (*duty.Duty, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*duty.Duty), args.Error(1)
}
func (m *MockDutyRepo) FindByID(ctx context.Context, id uuid.UUID) (*duty.Duty, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*duty.Duty), args.Error(1)
}
func (m *MockDutyRepo) CountDistinctStartDates(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}
func (m *MockDutyRepo) FindLastByTaskDefID(ctx context.Context, id uuid.UUID) (*duty.Duty, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*duty.Duty), args.Error(1)
}
func (m *MockDutyRepo) FindAllLatest(ctx context.Context) ([]*duty.Duty, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*duty.Duty), args.Error(1)
}

type MockGroupRepo struct{ mock.Mock }

func (m *MockGroupRepo) FindAll(ctx context.Context) ([]*structure.Group, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*structure.Group), args.Error(1)
}
func (m *MockGroupRepo) FindByID(ctx context.Context, id uuid.UUID) (*structure.Group, error) {
	return nil, nil
}
func (m *MockGroupRepo) FindByDormitoryID(ctx context.Context, dormitoryID int64) ([]*structure.Group, error) {
	return nil, nil
}
func (m *MockGroupRepo) Save(ctx context.Context, group *structure.Group) error { return nil }
func (m *MockGroupRepo) Delete(ctx context.Context, id uuid.UUID) error         { return nil }

type MockTeamRepo struct{ mock.Mock }

func (m *MockTeamRepo) FindByGroupID(ctx context.Context, id uuid.UUID) ([]*structure.Team, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]*structure.Team), args.Error(1)
}
func (m *MockTeamRepo) FindByID(ctx context.Context, id uuid.UUID) (*structure.Team, error) {
	return nil, nil
}
func (m *MockTeamRepo) Save(ctx context.Context, team *structure.Team) error { return nil }
func (m *MockTeamRepo) Delete(ctx context.Context, id uuid.UUID) error       { return nil }

type MockCatalogRepo struct{ mock.Mock }

func (m *MockCatalogRepo) GetAllTaskDefinitions(ctx context.Context) ([]*catalog.TaskDefinition, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*catalog.TaskDefinition), args.Error(1)
}
func (m *MockCatalogRepo) FindByID(ctx context.Context, id uuid.UUID) (*catalog.TaskDefinition, error) {
	return nil, nil
}
func (m *MockCatalogRepo) Save(ctx context.Context, task *catalog.TaskDefinition) error {
	return nil
}
func (m *MockCatalogRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type MockAreaRepo struct{ mock.Mock }

func (m *MockAreaRepo) GetAllAreas(ctx context.Context) ([]*catalog.Area, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*catalog.Area), args.Error(1)
}

type MockEventBus struct{ mock.Mock }

func (m *MockEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	return m.Called(ctx, topic, event).Error(0)
}
func (m *MockEventBus) Subscribe(topic string, handler ports.EventHandler) {}

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) Save(ctx context.Context, u *user.User) error { return nil }
func (m *MockUserRepo) FindAll(ctx context.Context) ([]*user.User, error) {
	return nil, nil
}
func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return nil, nil
}
func (m *MockUserRepo) FindByTelegramID(ctx context.Context, id int64) (*user.User, error) {
	return nil, nil
}
func (m *MockUserRepo) FindByTeamID(ctx context.Context, id uuid.UUID) ([]*user.User, error) {
	return nil, nil
}
func (m *MockUserRepo) FindByDormitoryID(ctx context.Context, dormitoryID int64) ([]*user.User, error) {
	return nil, nil
}
func (m *MockUserRepo) MoveUserToTeam(ctx context.Context, userID uuid.UUID, teamID *uuid.UUID) error {
	return nil
}
func (m *MockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error { return nil }
