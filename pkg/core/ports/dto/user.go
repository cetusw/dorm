package dto

import (
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/user"

	"github.com/google/uuid"
)

type UserListItem struct {
	ID            uuid.UUID
	TelegramID    int64
	FullName      string
	DormitoryName string
	RoomNumber    string
	GroupName     string
	TeamName      string
}

type CreateUserRequest struct {
	FirstName   string `form:"first_name"`
	LastName    string `form:"last_name"`
	MiddleName  string `form:"middle_name"`
	DormitoryID int64  `form:"dormitory_id"`
	Floor       int    `form:"floor"`
	RoomNumber  string `form:"room_number"`
	TeamID      string `form:"team_id"`
}

type ProfileViewModel struct {
	FirstName     string
	LastName      string
	RoomNumber    string
	DormitoryName string
	GroupName     string
	TeamID        uuid.UUID
	TeamName      string
}

type UserStats struct {
	ID              uuid.UUID
	FirstName       string
	LastName        string
	TotalPoints     int
	ConfirmedPoints int
	IsTeamLeader    bool
}

func NewUserStats(
	user *user.User,
	stats *duty.UserStats,
	isTeamLeader bool,
) *UserStats {
	if user == nil || stats == nil {
		return nil
	}
	return &UserStats{
		ID:              user.ID(),
		FirstName:       user.FirstName(),
		LastName:        user.LastName(),
		TotalPoints:     stats.TotalPoints,
		ConfirmedPoints: stats.ConfirmedPoints,
		IsTeamLeader:    isTeamLeader,
	}
}
