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

type DormitoryDetails struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	City        string       `json:"city"`
	StreetType  string       `json:"street_type"`
	StreetName  string       `json:"street_name"`
	HouseNumber string       `json:"house_number"`
	Leader      *UserSummary `json:"leader"`
}

type CreateDormitoryRequest struct {
	Name        string  `json:"name"`
	City        string  `json:"city"`
	StreetType  string  `json:"street_type"`
	StreetName  string  `json:"street_name"`
	HouseNumber string  `json:"house_number"`
	LeaderID    *string `json:"leader_id"`
}

type UpdateDormitoryRequest struct {
	Name        string  `json:"name"`
	City        string  `json:"city"`
	StreetType  string  `json:"street_type"`
	StreetName  string  `json:"street_name"`
	HouseNumber string  `json:"house_number"`
	LeaderID    *string `json:"leader_id"`
}

type UserSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserOptionItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserOptionsResponse struct {
	Users []UserOptionItem `json:"users"`
}

type GroupResponseItem struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Leader *UserSummary `json:"leader"`
}

type GroupListResponse struct {
	Groups []GroupResponseItem `json:"groups"`
}

type GroupDetails struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Leader *UserSummary `json:"leader"`
}

type CreateGroupRequest struct {
	Name        string  `json:"name"`
	LeaderID    *string `json:"leader_id"`
	DormitoryID int64   `json:"dormitory_id"`
}

type UpdateGroupRequest struct {
	Name        string  `json:"name"`
	LeaderID    *string `json:"leader_id"`
	DormitoryID int64   `json:"dormitory_id"`
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
