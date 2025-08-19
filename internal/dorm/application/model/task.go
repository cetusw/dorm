package model

import "github.com/google/uuid"

type Task struct {
	TaskID uuid.UUID
	AreaID int
	Title  string
	Cost   int
}
