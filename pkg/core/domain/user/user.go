package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidTelegramID = errors.New("invalid telegram id")
	ErrEmptyName         = errors.New("name cannot be empty")
)

type User struct {
	id          uuid.UUID
	telegramID  *int64
	firstName   string
	middleName  *string
	lastName    string
	teamID      *uuid.UUID
	roomNumber  *string
	floorNumber *int
	dormitoryID *int64
	createdAt   time.Time
}

func NewUser(telegramID int64, firstName, lastName string) (*User, error) {
	if telegramID <= 0 {
		return nil, ErrInvalidTelegramID
	}
	return newUserWithTelegramID(&telegramID, firstName, lastName)
}

func NewManualUser(firstName, lastName string) (*User, error) {
	return newUserWithTelegramID(nil, firstName, lastName)
}

func newUserWithTelegramID(telegramID *int64, firstName, lastName string) (*User, error) {
	if firstName == "" {
		return nil, ErrEmptyName
	}

	return &User{
		id:         uuid.New(),
		telegramID: telegramID,
		firstName:  firstName,
		lastName:   lastName,
		createdAt:  time.Now(),
	}, nil
}

func RestoreUser(
	id uuid.UUID,
	telegramID *int64,
	firstName string,
	middleName *string,
	lastName string,
	teamID *uuid.UUID,
	roomNumber *string,
	floorNumber *int,
	dormitoryID *int64,
	createdAt time.Time,
) *User {
	return &User{
		id:          id,
		telegramID:  telegramID,
		firstName:   firstName,
		middleName:  middleName,
		lastName:    lastName,
		teamID:      teamID,
		roomNumber:  roomNumber,
		floorNumber: floorNumber,
		dormitoryID: dormitoryID,
		createdAt:   createdAt,
	}
}

func (u *User) JoinTeam(teamID uuid.UUID) {
	u.teamID = &teamID
}

func (u *User) LeaveTeam() {
	u.teamID = nil
}

func (u *User) MoveInto(dormID int64, roomNumber string) {
	u.dormitoryID = &dormID
	u.roomNumber = &roomNumber
}

func (u *User) SetMiddleName(middleName string) {
	if middleName == "" {
		u.middleName = nil
		return
	}
	u.middleName = &middleName
}

func (u *User) SetFloorNumber(floorNumber int) {
	if floorNumber == 0 {
		u.floorNumber = nil
		return
	}
	u.floorNumber = &floorNumber
}

func (u *User) Rename(firstName, lastName string) error {
	if firstName == "" {
		return ErrEmptyName
	}
	u.firstName = firstName
	u.lastName = lastName
	return nil
}

func (u *User) ID() uuid.UUID { return u.id }
func (u *User) TelegramID() int64 {
	if u.telegramID == nil {
		return 0
	}
	return *u.telegramID
}
func (u *User) TelegramIDValue() *int64 { return u.telegramID }
func (u *User) FirstName() string       { return u.firstName }
func (u *User) MiddleName() *string     { return u.middleName }
func (u *User) LastName() string        { return u.lastName }
func (u *User) TeamID() *uuid.UUID      { return u.teamID }
func (u *User) RoomNumber() *string     { return u.roomNumber }
func (u *User) FloorNumber() *int       { return u.floorNumber }
func (u *User) DormitoryID() *int64     { return u.dormitoryID }
func (u *User) CreatedAt() time.Time    { return u.createdAt }

type Repository interface {
	Save(ctx context.Context, user *User) error
	FindAll(ctx context.Context) ([]*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]*User, error)
	FindByDormitoryID(ctx context.Context, dormitoryID int64) ([]*User, error)
	MoveUserToTeam(ctx context.Context, userID uuid.UUID, teamID *uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
