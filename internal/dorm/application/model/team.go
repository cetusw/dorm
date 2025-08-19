package model

import "github.com/google/uuid"

type Team struct {
	TeamID       int
	TeamLeaderID uuid.UUID
	Color        int
}
