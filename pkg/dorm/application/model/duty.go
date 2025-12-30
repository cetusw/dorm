package model

import (
	"time"

	"github.com/google/uuid"
)

type Duty struct {
	ID     uuid.UUID
	TeamID uuid.UUID
	Start  time.Time
	End    time.Time
}
