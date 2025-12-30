package model

import (
	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	TelegramID  int64
	FirstName   string
	LastName    string
	MiddleName  *string
	TeamID      *uuid.UUID
	RoomNumber  *int
	DormitoryID *int64
	RoleID      int
	CreatedAt   []uint8
	DeletedAt   *[]uint8
}

type UsersToNotifyFilter struct {
	Names       []string
	RoleIDs     []int
	DormitoryID *int
}
