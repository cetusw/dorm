package dto

import (
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/user"

	"github.com/google/uuid"
)

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
}

func NewUserStats(
	user *user.User,
	stats *duty.UserStats,
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
	}
}
