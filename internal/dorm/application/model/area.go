package model

import "github.com/google/uuid"

type Area struct {
	AreaID  int
	Floor   int
	Name    string
	GroupID uuid.UUID
}
