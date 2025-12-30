package model

import "github.com/google/uuid"

type Dormitory struct {
	ID          int64
	LeaderID    *uuid.UUID
	Name        string
	City        string
	StreetType  string
	StreetName  string
	HouseNumber string
}
