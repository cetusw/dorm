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
	telegramID  int64
	firstName   string
	lastName    string
	teamID      *uuid.UUID
	roomNumber  *string
	dormitoryID *int64
	createdAt   time.Time
}

func NewUser(telegramID int64, firstName, lastName string) (*User, error) {
	if telegramID == 0 {
		return nil, ErrInvalidTelegramID
	}
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
	telegramID int64,
	firstName, lastName string,
	teamID *uuid.UUID,
	roomNumber *string,
	dormitoryID *int64,
	createdAt time.Time,
) *User {
	return &User{
		id:          id,
		telegramID:  telegramID,
		firstName:   firstName,
		lastName:    lastName,
		teamID:      teamID,
		roomNumber:  roomNumber,
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

func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) TelegramID() int64    { return u.telegramID }
func (u *User) FirstName() string    { return u.firstName }
func (u *User) LastName() string     { return u.lastName }
func (u *User) TeamID() *uuid.UUID   { return u.teamID }
func (u *User) RoomNumber() *string  { return u.roomNumber }
func (u *User) DormitoryID() *int64  { return u.dormitoryID }
func (u *User) CreatedAt() time.Time { return u.createdAt }

type Repository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]*User, error)
}
