package user

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUser_Success(t *testing.T) {
	telegramID := int64(123456789)
	firstName := "John"
	lastName := "Doe"

	u, err := NewUser(telegramID, firstName, lastName)

	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.NotEqual(t, uuid.Nil, u.ID())
	assert.Equal(t, telegramID, u.TelegramID())
	assert.Equal(t, firstName, u.FirstName())
	assert.Equal(t, lastName, u.LastName())
	assert.WithinDuration(t, time.Now(), u.CreatedAt(), time.Second)
	assert.Nil(t, u.TeamID())
	assert.Nil(t, u.DormitoryID())
}

func TestNewUser_Validation(t *testing.T) {
	t.Run("Invalid TelegramID", func(t *testing.T) {
		_, err := NewUser(0, "John", "Doe")
		assert.ErrorIs(t, err, ErrInvalidTelegramID)
	})

	t.Run("Empty Name", func(t *testing.T) {
		_, err := NewUser(123, "", "Doe")
		assert.ErrorIs(t, err, ErrEmptyName)
	})
}

func TestNewManualUser(t *testing.T) {
	u, err := NewManualUser("John", "Doe")

	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, int64(0), u.TelegramID())
	assert.Nil(t, u.TelegramIDValue())
}

func TestRestoreUser(t *testing.T) {
	id := uuid.New()
	telegramID := int64(12345)
	teamID := uuid.New()
	middleName := "Middle"
	roomNumber := "10"
	dormID := int64(1)
	now := time.Now()

	u := RestoreUser(
		id,
		&telegramID,
		"Jane",
		&middleName,
		"Smith",
		&teamID,
		&roomNumber,
		nil,
		&dormID,
		now,
	)

	assert.Equal(t, id, u.ID())
	assert.Equal(t, int64(12345), u.TelegramID())
	assert.Equal(t, "Jane", u.FirstName())
	assert.Equal(t, &middleName, u.MiddleName())
	assert.Equal(t, &teamID, u.TeamID())
	assert.Equal(t, &dormID, u.DormitoryID())
	assert.Equal(t, now, u.CreatedAt())
}

func TestUser_TeamManagement(t *testing.T) {
	u, _ := NewUser(123, "Test", "User")
	teamID := uuid.New()

	u.JoinTeam(teamID)
	assert.NotNil(t, u.TeamID())
	assert.Equal(t, teamID, *u.TeamID())

	u.LeaveTeam()
	assert.Nil(t, u.TeamID())
}

func TestUser_DormitoryAssignment(t *testing.T) {
	u, _ := NewUser(123, "Test", "User")
	roomNumber := "11A"
	dormID := int64(5)

	u.MoveInto(dormID, roomNumber)
	u.SetFloorNumber(3)
	assert.NotNil(t, u.DormitoryID())
	assert.Equal(t, dormID, *u.DormitoryID())
	assert.NotNil(t, u.FloorNumber())
	assert.Equal(t, 3, *u.FloorNumber())

	u.SetFloorNumber(0)
	assert.Nil(t, u.FloorNumber())
}

func TestUser_MiddleName(t *testing.T) {
	u, _ := NewUser(123, "Test", "User")

	u.SetMiddleName("Middle")
	assert.NotNil(t, u.MiddleName())
	assert.Equal(t, "Middle", *u.MiddleName())

	u.SetMiddleName("")
	assert.Nil(t, u.MiddleName())
}
