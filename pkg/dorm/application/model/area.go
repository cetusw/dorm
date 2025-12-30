package model

import "github.com/google/uuid"

type Area struct {
	ID      int
	Floor   int
	Name    string
	GroupID uuid.UUID
}
