package dto

import "time"

type DutySettingsGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DormitoryID int64  `json:"dormitory_id"`
}

type DutySettingsTask struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Cost            int        `json:"cost"`
	Frequency       int        `json:"frequency"`
	LastCompletedAt *time.Time `json:"last_completed_at"`
}

type DutySettingsArea struct {
	ID    int                `json:"id"`
	Name  string             `json:"name"`
	Floor *int               `json:"floor"`
	Tasks []DutySettingsTask `json:"tasks"`
}

type DutySettingsResponse struct {
	Group DutySettingsGroup  `json:"group"`
	Areas []DutySettingsArea `json:"areas"`
}
