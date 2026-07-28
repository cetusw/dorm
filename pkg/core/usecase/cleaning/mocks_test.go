package cleaning

import (
	"context"
	"dorm/pkg/core/ports"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
)

type MockDutyRepo struct{ mock.Mock }

func (m *MockDutyRepo) CreateWithTasks(ctx context.Context, d *duty.Duty, tasks []*duty.DutyTask) error {
	return m.Called(ctx, d, tasks).Error(0)
}
func (m *MockDutyRepo) FindCurrentByTeamID(ctx context.Context, id uuid.UUID) (*duty.Duty, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*duty.Duty), args.Error(1)
}
func (m *MockDutyRepo) FindActiveByTeamID(ctx context.Context, id uuid.UUID, at time.Time) (*duty.Duty, error) {
	args := m.Called(ctx, id, at)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*duty.Duty), args.Error(1)
}
func (m *MockDutyRepo) FindByID(ctx context.Context, id uuid.UUID) (*duty.Duty, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*duty.Duty), args.Error(1)
}
func (m *MockDutyRepo) FindByGroupID(ctx context.Context, id uuid.UUID) ([]*duty.Duty, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*duty.Duty), args.Error(1)
}
func (m *MockDutyRepo) FindLatestByGroupID(ctx context.Context, id uuid.UUID) (*duty.Duty, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*duty.Duty), args.Error(1)
}
func (m *MockDutyRepo) ReassignTeamAndResetTasks(ctx context.Context, dutyID uuid.UUID, teamID uuid.UUID) error {
	return m.Called(ctx, dutyID, teamID).Error(0)
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

type MockDutyTaskRepo struct{ mock.Mock }

func (m *MockDutyTaskRepo) FindByID(ctx context.Context, id uuid.UUID) (*duty.DutyTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*duty.DutyTask), args.Error(1)
}
func (m *MockDutyTaskRepo) FindByDutyID(ctx context.Context, id uuid.UUID) ([]*duty.DutyTask, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]*duty.DutyTask), args.Error(1)
}
func (m *MockDutyTaskRepo) Create(ctx context.Context, task *duty.DutyTask) error {
	return m.Called(ctx, task).Error(0)
}
func (m *MockDutyTaskRepo) DeletePending(ctx context.Context, taskID uuid.UUID) error {
	return m.Called(ctx, taskID).Error(0)
}
func (m *MockDutyTaskRepo) Assign(ctx context.Context, taskID, assigneeID uuid.UUID, assignedAt time.Time) error {
	return m.Called(ctx, taskID, assigneeID, assignedAt).Error(0)
}
func (m *MockDutyTaskRepo) Unassign(ctx context.Context, taskID, assigneeID uuid.UUID) error {
	return m.Called(ctx, taskID, assigneeID).Error(0)
}
func (m *MockDutyTaskRepo) Complete(ctx context.Context, taskID, assigneeID uuid.UUID, completedAt time.Time) error {
	return m.Called(ctx, taskID, assigneeID, completedAt).Error(0)
}
func (m *MockDutyTaskRepo) CancelCompletion(ctx context.Context, taskID, assigneeID uuid.UUID) error {
	return m.Called(ctx, taskID, assigneeID).Error(0)
}
func (m *MockDutyTaskRepo) Verify(ctx context.Context, taskID, reviewerID uuid.UUID, verifiedAt time.Time) error {
	return m.Called(ctx, taskID, reviewerID, verifiedAt).Error(0)
}
func (m *MockDutyTaskRepo) Reopen(ctx context.Context, taskID, reviewerID uuid.UUID) error {
	return m.Called(ctx, taskID, reviewerID).Error(0)
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
func (m *MockTeamRepo) UpdateRotationPositions(ctx context.Context, groupID uuid.UUID, orderedTeamIDs []uuid.UUID) error {
	return nil
}
func (m *MockTeamRepo) Save(ctx context.Context, team *structure.Team) error { return nil }
func (m *MockTeamRepo) Delete(ctx context.Context, id uuid.UUID) error       { return nil }

type MockCatalogRepo struct{ mock.Mock }

func (m *MockCatalogRepo) GetAllTaskDefinitions(ctx context.Context) ([]*catalog.TaskDefinition, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*catalog.TaskDefinition), args.Error(1)
}
func (m *MockCatalogRepo) FindLastCompletionDates(ctx context.Context, taskIDs []uuid.UUID) (map[uuid.UUID]*time.Time, error) {
	return nil, nil
}
func (m *MockCatalogRepo) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*catalog.TaskDefinition, error) {
	return nil, nil
}
func (m *MockCatalogRepo) FindCommon(ctx context.Context) ([]*catalog.TaskDefinition, error) {
	return nil, nil
}
func (m *MockCatalogRepo) FindByID(ctx context.Context, id uuid.UUID) (*catalog.TaskDefinition, error) {
	return nil, nil
}
func (m *MockCatalogRepo) Save(ctx context.Context, task *catalog.TaskDefinition) error {
	return nil
}
func (m *MockCatalogRepo) SoftDelete(ctx context.Context, id uuid.UUID) error { return nil }

type MockTaskOverrideRepo struct{ mock.Mock }

func (m *MockTaskOverrideRepo) FindAll(ctx context.Context) ([]*catalog.DutyTaskOverride, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*catalog.DutyTaskOverride), args.Error(1)
}

func (m *MockTaskOverrideRepo) ReplaceForTasks(ctx context.Context, taskIDs []uuid.UUID, overrides []*catalog.DutyTaskOverride) error {
	return m.Called(ctx, taskIDs, overrides).Error(0)
}

func (m *MockTaskOverrideRepo) DeleteByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) error {
	return m.Called(ctx, taskIDs).Error(0)
}

type MockAreaRepo struct{ mock.Mock }

func (m *MockAreaRepo) GetAllAreas(ctx context.Context) ([]*catalog.Area, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*catalog.Area), args.Error(1)
}
func (m *MockAreaRepo) FindByID(ctx context.Context, id int) (*catalog.Area, error) {
	return nil, nil
}
func (m *MockAreaRepo) Save(ctx context.Context, area *catalog.Area) error {
	return nil
}
func (m *MockAreaRepo) Delete(ctx context.Context, id int) error {
	return nil
}

type MockEventBus struct{ mock.Mock }

func (m *MockEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	return m.Called(ctx, topic, event).Error(0)
}
func (m *MockEventBus) Subscribe(_ string, _ ports.EventHandler) {}

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) Save(ctx context.Context, u *user.User) error { return nil }
func (m *MockUserRepo) FindAll(ctx context.Context) ([]*user.User, error) {
	return nil, nil
}
func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}
func (m *MockUserRepo) FindByLogin(ctx context.Context, login string) (*user.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
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
