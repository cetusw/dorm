package model

import (
	"time"

	"github.com/google/uuid"
)

type Duty struct {
	DutyID uuid.UUID
	TeamID int
	Start  time.Time
	End    time.Time
}
