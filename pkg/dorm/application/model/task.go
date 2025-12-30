package model

import "github.com/google/uuid"

type Task struct {
	ID        uuid.UUID
	AreaID    int
	Title     string
	Cost      int
	Frequency int
	Scope     string // "public" or "private"
}
