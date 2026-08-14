package user

import (
	"context"
	"testing"
	"time"

	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userRepositoryStub struct {
	byID map[uuid.UUID]*user.User
}

func (s *userRepositoryStub) Save(context.Context, *user.User) error { return nil }
func (s *userRepositoryStub) FindAll(context.Context) ([]*user.User, error) {
	return nil, nil
}
func (s *userRepositoryStub) FindByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	return s.byID[id], nil
}
func (s *userRepositoryStub) FindByLogin(context.Context, string) (*user.User, error) {
	return nil, nil
}
func (s *userRepositoryStub) FindByTeamID(context.Context, uuid.UUID) ([]*user.User, error) {
	return nil, nil
}
func (s *userRepositoryStub) FindByDormitoryID(context.Context, int64) ([]*user.User, error) {
	return nil, nil
}
func (s *userRepositoryStub) MoveUserToTeam(context.Context, uuid.UUID, *uuid.UUID) error {
	return nil
}
func (s *userRepositoryStub) SoftDelete(context.Context, uuid.UUID) error { return nil }

type dormitoryRepositoryStub struct {
	dormitory        *structure.Dormitory
	existsByLeaderID bool
}

func (s *dormitoryRepositoryStub) FindAll(context.Context) ([]*structure.Dormitory, error) {
	return nil, nil
}
func (s *dormitoryRepositoryStub) FindByID(context.Context, int64) (*structure.Dormitory, error) {
	return s.dormitory, nil
}
func (s *dormitoryRepositoryStub) ExistsByLeaderID(context.Context, uuid.UUID) (bool, error) {
	return s.existsByLeaderID, nil
}
func (s *dormitoryRepositoryStub) Save(context.Context, *structure.Dormitory) error { return nil }
func (s *dormitoryRepositoryStub) Delete(context.Context, int64) error              { return nil }

type groupRepositoryStub struct {
	groups []*structure.Group
}

func (s *groupRepositoryStub) FindAll(context.Context) ([]*structure.Group, error) {
	return nil, nil
}
func (s *groupRepositoryStub) FindByID(context.Context, uuid.UUID) (*structure.Group, error) {
	return nil, nil
}
func (s *groupRepositoryStub) FindByDormitoryID(context.Context, int64) ([]*structure.Group, error) {
	return s.groups, nil
}
func (s *groupRepositoryStub) Save(context.Context, *structure.Group) error { return nil }
func (s *groupRepositoryStub) Delete(context.Context, uuid.UUID) error      { return nil }

func TestGetCurrentUserIncludesWarehousePermissionForGroupLeader(t *testing.T) {
	t.Parallel()

	currentUserID := uuid.New()
	dormitoryID := int64(7)
	now := time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)
	currentUser := user.RestoreUser(currentUserID, "group-head", "hash", "Ivan", nil, "Ivanov", nil, nil, nil, &dormitoryID, now)

	service := NewUserService(
		&userRepositoryStub{byID: map[uuid.UUID]*user.User{currentUserID: currentUser}},
		&dormitoryRepositoryStub{dormitory: structure.RestoreDormitory(dormitoryID, "Dorm", nil, "Moscow", "st", "Lenina", "1")},
		&groupRepositoryStub{groups: []*structure.Group{
			structure.RestoreGroup(uuid.New(), &currentUserID, "Group A", dormitoryID),
		}},
		userQueryServiceAdapter{},
	)

	response, err := service.GetCurrentUser(context.Background(), currentUserID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.CanManageWarehouse)
}

type userQueryServiceAdapter struct{}

func (userQueryServiceAdapter) GetResidentsByDormitoryID(context.Context, int64) ([]dto.ResidentListItem, error) {
	return nil, nil
}
