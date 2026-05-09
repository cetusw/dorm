package dto

import "github.com/google/uuid"

type DormitoryListItem struct {
	ID          int64
	Name        string
	Address     string
	LeaderName  string
	GroupsCount int
}

type GroupListItem struct {
	ID             uuid.UUID
	Name           string
	LeaderName     string
	SpreadsheetID  string
	SpreadsheetURL string
	DormitoryID    int64
	TeamsCount     int
}

type UpsertDormitoryRequest struct {
	Name        string `form:"name"`
	LeaderID    string `form:"leader_id"`
	City        string `form:"city"`
	StreetType  string `form:"street_type"`
	StreetName  string `form:"street_name"`
	HouseNumber string `form:"house_number"`
}

type UpsertGroupRequest struct {
	Name          string `form:"name"`
	LeaderID      string `form:"leader_id"`
	SpreadsheetID string `form:"spreadsheet_id"`
}

type UserOption struct {
	ID       uuid.UUID
	FullName string
}
