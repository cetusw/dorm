package ports

import "github.com/google/uuid"

type TaskViewModel struct {
	ID       uuid.UUID
	Title    string
	AreaName string
	AreaID   int
	Cost     int
	IsDone   bool
}

// TODO: подумать о расположении этой модели
