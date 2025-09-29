package model

import (
	"github.com/google/uuid"
)

type User struct {
	UserID      uuid.UUID
	TelegramID  int64
	FirstName   string
	LastName    string
	MiddleName  *string
	TeamID      *uuid.UUID
	RoomNumber  *int
	DormitoryID *int
	RoleID      int
	CreatedAt   []uint8
	DeletedAt   *[]uint8
}
