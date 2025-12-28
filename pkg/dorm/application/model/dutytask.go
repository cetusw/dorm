package model

import (
	"time"

	"github.com/google/uuid"
)

type DutyTask struct {
	DutyTaskID       uuid.UUID
	DutyID           uuid.UUID
	TaskID           uuid.UUID
	AssigneeID       *uuid.UUID
	ReviewerID       uuid.UUID
	AssignmentDate   *time.Time
	CompletionDate   *time.Time
	VerificationDate *time.Time
}

type DutyTaskView struct {
	AreaFloor         int
	AreaName          string
	TaskID            uuid.UUID
	TaskTitle         string
	TaskCost          int
	AssigneeID        *uuid.UUID
	AssigneeFirstName *string
	AssigneeLastName  *string
	CompletionDate    *time.Time
	VerificationDate  *time.Time
}
