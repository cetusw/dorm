package model

import "github.com/google/uuid"

type Team struct {
	ID           uuid.UUID
	GroupID      uuid.UUID
	TeamLeaderID uuid.UUID
	Color        string
	Order        int
}
