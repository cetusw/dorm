package model

import "github.com/google/uuid"

type Group struct {
	GroupID       uuid.UUID
	Name          string
	DormitoryID   int
	SpreadsheetID string
}
