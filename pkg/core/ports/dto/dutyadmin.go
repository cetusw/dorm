package dto

import (
	"time"

	"github.com/google/uuid"
)

type DutyListItem struct {
	ID            uuid.UUID
	StartDate     time.Time
	EndDate       time.Time
	DormitoryID   int64
	DormitoryName string
	GroupID       uuid.UUID
	GroupName     string
	TeamName      string
}

type FutureDutyTaskItem struct {
	ID                uuid.UUID
	AreaID            int
	AreaName          string
	Title             string
	Frequency         int
	LastCompletedAt   *time.Time
	IsDueByFrequency  bool
	HasOverride       bool
	IncludeInNextDuty bool
	IsCommon          bool
}

type FutureDutyTaskGroup struct {
	AreaID            int
	AreaName          string
	Floor             int
	IsCommon          bool
	IncludeTasksCount int
	Tasks             []FutureDutyTaskItem
}

type DutyGroupSettingsUpdate struct {
	GroupID        uuid.UUID
	NextDutyTeam   *int
	IncludeTaskIDs []uuid.UUID
}

type UpdateDormitoryDutySettingsRequest struct {
	CommonTaskIDs []uuid.UUID
	Groups        []DutyGroupSettingsUpdate
}

type DutyTaskListItem struct {
	AreaName     string
	Title        string
	AssigneeName string
	Status       string
}

type DutyDetailItem struct {
	ID            uuid.UUID
	StartDate     time.Time
	EndDate       time.Time
	DormitoryID   int64
	DormitoryName string
	GroupID       uuid.UUID
	GroupName     string
	TeamName      string
	Tasks         []DutyTaskListItem
}
