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
	TeamID      *int
	RoomNumber  *int
	DormitoryID *uuid.UUID
	RoleID      int
	CreatedAt   []uint8
	DeletedAt   *[]uint8
}
