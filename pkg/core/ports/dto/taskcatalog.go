package dto

import "github.com/google/uuid"

type TaskCatalogItem struct {
	ID                 uuid.UUID
	AreaID             int
	AreaName           string
	AreaFloor          int
	Title              string
	Cost               int
	RecurrenceInterval int
	StartSequence      int
}

type TaskCatalogGroup struct {
	AreaID   int
	AreaName string
	Floor    int
	Tasks    []TaskCatalogItem
}

type UpsertTaskCatalogRequest struct {
	AreaID             int    `form:"area_id"`
	Title              string `form:"title"`
	Cost               int    `form:"cost"`
	RecurrenceInterval int    `form:"recurrenceInterval"`
}

type AreaResponseItem struct {
	ID    int          `json:"id"`
	Name  string       `json:"name"`
	Floor *int         `json:"floor"`
	Group *GroupOption `json:"group"`
}

type AreaListResponse struct {
	Areas []AreaResponseItem `json:"areas"`
}

type AreaDetails struct {
	ID    int          `json:"id"`
	Name  string       `json:"name"`
	Floor *int         `json:"floor"`
	Group *GroupOption `json:"group"`
}

type GroupOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateAreaRequest struct {
	Name    string  `json:"name"`
	GroupID *string `json:"group_id"`
	Floor   *int    `json:"floor"`
}

type UpdateAreaRequest struct {
	Name    string  `json:"name"`
	GroupID *string `json:"group_id"`
	Floor   *int    `json:"floor"`
}

type AreaSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TaskResponseItem struct {
	ID                 string      `json:"id"`
	Title              string      `json:"title"`
	Cost               int         `json:"cost"`
	RecurrenceInterval int         `json:"recurrenceInterval"`
	StartSequence      int         `json:"startSequence"`
	Area               AreaSummary `json:"area"`
}

type TaskListResponse struct {
	Tasks []TaskResponseItem `json:"tasks"`
}

type TaskDetails struct {
	ID                 string      `json:"id"`
	Title              string      `json:"title"`
	Cost               int         `json:"cost"`
	RecurrenceInterval int         `json:"recurrenceInterval"`
	StartSequence      int         `json:"startSequence"`
	Area               AreaSummary `json:"area"`
}

type CreateTaskRequest struct {
	Title                string `json:"title"`
	Cost                 int    `json:"cost"`
	RecurrenceInterval   int    `json:"recurrenceInterval"`
	AreaID               int    `json:"area_id"`
	IncludeInCurrentDuty bool   `json:"include_in_current_duty"`
}

type UpdateTaskRequest struct {
	Title              string `json:"title"`
	Cost               int    `json:"cost"`
	RecurrenceInterval int    `json:"recurrenceInterval"`
	AreaID             int    `json:"area_id"`
}
