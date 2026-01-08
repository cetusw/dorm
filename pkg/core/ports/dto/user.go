package dto

import "github.com/google/uuid"

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
	FirstName       string
	LastName        string
	TotalPoints     int
	ConfirmedPoints int
}
