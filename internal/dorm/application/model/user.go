package model

import (
	"database/sql"

	"github.com/google/uuid"
)

type User struct {
	UserId      uuid.UUID
	TelegramId  int64
	FirstName   string
	LastName    string
	MiddleName  string
	TeamId      uuid.UUID
	RoomNumber  int
	DormitoryId uuid.UUID
	RoleLevel   int
	CreatedAt   []uint8
	DeletedAt   sql.Null[[]uint8]
}
