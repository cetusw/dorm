package dto

import (
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/user"

	"github.com/google/uuid"
)

type UserListItem struct {
	ID            uuid.UUID
	Login         string
	TelegramID    int64
	FullName      string
	DormitoryName string
	RoomNumber    string
	GroupName     string
	TeamName      string
}

type ResidentListItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	RoomNumber string `json:"room_number"`
}

type ResidentListResponse struct {
	Users []ResidentListItem `json:"users"`
}

type ResidentDetails struct {
	ID         string  `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName *string `json:"middle_name"`
	Login      string  `json:"login"`
	Floor      *int    `json:"floor"`
	RoomNumber *string `json:"room_number"`
}

type CreateResidentRequest struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	MiddleName  *string `json:"middle_name"`
	Login       string  `json:"login"`
	Password    string  `json:"password"`
	DormitoryID int64   `json:"dormitory_id"`
	Floor       *int    `json:"floor"`
	RoomNumber  *string `json:"room_number"`
}

type UpdateResidentRequest struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	MiddleName  *string `json:"middle_name"`
	Login       string  `json:"login"`
	Password    string  `json:"password"`
	DormitoryID int64   `json:"dormitory_id"`
	Floor       *int    `json:"floor"`
	RoomNumber  *string `json:"room_number"`
}

type CreateUserRequest struct {
	FirstName   string `form:"first_name"`
	LastName    string `form:"last_name"`
	MiddleName  string `form:"middle_name"`
	Login       string `form:"login"`
	Password    string `form:"password"`
	DormitoryID int64  `form:"dormitory_id"`
	Floor       int    `form:"floor"`
	RoomNumber  string `form:"room_number"`
	TeamID      string `form:"team_id"`
}

type UpdateUserRequest struct {
	FirstName   string `form:"first_name"`
	MiddleName  string `form:"middle_name"`
	LastName    string `form:"last_name"`
	Login       string `form:"login"`
	Password    string `form:"password"`
	DormitoryID int64  `form:"dormitory_id"`
	Floor       int    `form:"floor"`
	RoomNumber  string `form:"room_number"`
}

type ResidentLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CurrentUserResponse struct {
	ID                   string `json:"id"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	CanManageDormitories bool   `json:"can_manage_dormitories"`
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
