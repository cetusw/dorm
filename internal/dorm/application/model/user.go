package model

import (
	"database/sql"

	"github.com/google/uuid"
)

type User struct {
	UserID      uuid.UUID
	TelegramID  int64
	FirstName   string
	LastName    string
	MiddleName  string
	TeamID      uuid.UUID
	RoomNumber  int
	DormitoryID uuid.UUID
	RoleID      int
	CreatedAt   []uint8
	DeletedAt   sql.Null[[]uint8]
}
