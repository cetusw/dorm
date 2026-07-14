package dto

import "github.com/google/uuid"

type DormitoryListItem struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Address     string       `json:"address"`
	Leader      *UserSummary `json:"leader"`
	LeaderName  string       `json:"-"`
	GroupsCount int          `json:"-"`
}

type DormitoryListResponse struct {
	Dormitories []DormitoryListItem `json:"dormitories"`
}

type UserSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
