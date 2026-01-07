package ports

import "github.com/google/uuid"

type TaskViewModel struct {
	ID               uuid.UUID
	Title            string
	AreaName         string
	AreaID           int
	AreaFloor        int
	Cost             int
	IsDone           bool
	IsAssignedToUser bool
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
