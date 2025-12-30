package model

import "github.com/google/uuid"

type Team struct {
	ID       uuid.UUID
	GroupID  uuid.UUID
	LeaderID uuid.UUID
	Color    string
	Order    int
}
