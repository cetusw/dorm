package model

import "github.com/google/uuid"

type Group struct {
	ID            uuid.UUID
	LeaderID      *uuid.UUID
	Name          string
	DormitoryID   int64
	SpreadsheetID string
}
