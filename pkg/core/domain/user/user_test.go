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
	login := "john.doe"
	passwordHash := "hash"

	u, err := NewUser(telegramID, firstName, lastName, login, passwordHash)

	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.NotEqual(t, uuid.Nil, u.ID())
	assert.Equal(t, telegramID, u.TelegramID())
	assert.Equal(t, login, u.Login())
	assert.Equal(t, passwordHash, u.PasswordHash())
	assert.Equal(t, firstName, u.FirstName())
	assert.Equal(t, lastName, u.LastName())
	assert.WithinDuration(t, time.Now(), u.CreatedAt(), time.Second)
	assert.Nil(t, u.TeamID())
	assert.Nil(t, u.DormitoryID())
}

func TestNewUser_Validation(t *testing.T) {
	t.Run("Invalid TelegramID", func(t *testing.T) {
		_, err := NewUser(0, "John", "Doe", "john.doe", "hash")
		assert.ErrorIs(t, err, ErrInvalidTelegramID)
	})

	t.Run("Empty Name", func(t *testing.T) {
		_, err := NewUser(123, "", "Doe", "john.doe", "hash")
		assert.ErrorIs(t, err, ErrEmptyName)
	})

	t.Run("Empty Login", func(t *testing.T) {
		_, err := NewUser(123, "John", "Doe", "", "hash")
		assert.ErrorIs(t, err, ErrEmptyLogin)
	})

	t.Run("Empty PasswordHash", func(t *testing.T) {
		_, err := NewUser(123, "John", "Doe", "john.doe", "")
		assert.ErrorIs(t, err, ErrEmptyPasswordHash)
	})
}

func TestNewManualUser(t *testing.T) {
	u, err := NewManualUser("John", "Doe", "john.doe", "hash")

	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, int64(0), u.TelegramID())
	assert.Nil(t, u.TelegramIDValue())
	assert.Equal(t, "john.doe", u.Login())
	assert.Equal(t, "hash", u.PasswordHash())
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
		"jane.smith",
		"hashed-password",
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
	assert.Equal(t, "jane.smith", u.Login())
	assert.Equal(t, "hashed-password", u.PasswordHash())
	assert.Equal(t, "Jane", u.FirstName())
	assert.Equal(t, &middleName, u.MiddleName())
	assert.Equal(t, &teamID, u.TeamID())
	assert.Equal(t, &dormID, u.DormitoryID())
	assert.Equal(t, now, u.CreatedAt())
}

func TestUser_TeamManagement(t *testing.T) {
	u, _ := NewUser(123, "Test", "User", "test.user", "hash")
	teamID := uuid.New()

	u.JoinTeam(teamID)
	assert.NotNil(t, u.TeamID())
	assert.Equal(t, teamID, *u.TeamID())

	u.LeaveTeam()
	assert.Nil(t, u.TeamID())
}

func TestUser_DormitoryAssignment(t *testing.T) {
	u, _ := NewUser(123, "Test", "User", "test.user", "hash")
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
	u, _ := NewUser(123, "Test", "User", "test.user", "hash")

	u.SetMiddleName("Middle")
	assert.NotNil(t, u.MiddleName())
	assert.Equal(t, "Middle", *u.MiddleName())

	u.SetMiddleName("")
	assert.Nil(t, u.MiddleName())
}

func TestUser_SetCredentials(t *testing.T) {
	u, _ := NewManualUser("John", "Doe", "john.doe", "hash")

	err := u.SetCredentials("johnny", "new-hash")

	assert.NoError(t, err)
	assert.Equal(t, "johnny", u.Login())
	assert.Equal(t, "new-hash", u.PasswordHash())
}
