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

type DutySettingsTeam struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	RotationPosition int          `json:"rotation_position"`
	Leader           *UserSummary `json:"leader"`
	MembersCount     int          `json:"members_count"`
}

type DutySettingsTeamMember struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsLeader bool   `json:"is_leader"`
}

type DutySettingsTeamMembersResponse struct {
	TeamName string                   `json:"team_name"`
	Leader   *UserSummary             `json:"leader"`
	Members  []DutySettingsTeamMember `json:"members"`
}

type DutySettingsTeamSearchItem struct {
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	CurrentTeamID     *string      `json:"current_team_id"`
	CurrentTeamLeader *UserSummary `json:"current_team_leader"`
}

type DutySettingsTeamSearchResponse struct {
	Users []DutySettingsTeamSearchItem `json:"users"`
}

type DutySettingsTeamMemberRequest struct {
	UserID string `json:"user_id"`
}

type DutySettingsTeamLeaderRequest struct {
	UserID string `json:"user_id"`
}

type DutySettingsArea struct {
	ID    int                `json:"id"`
	Name  string             `json:"name"`
	Floor *int               `json:"floor"`
	Tasks []DutySettingsTask `json:"tasks"`
}

type DutySettingsResponse struct {
	Group            DutySettingsGroup  `json:"group"`
	Areas            []DutySettingsArea `json:"areas"`
	Teams            []DutySettingsTeam `json:"teams"`
	ActiveDutyTeamID *string            `json:"active_duty_team_id"`
}

type ReorderDutySettingsTeamsRequest struct {
	TeamIDs []string `json:"team_ids"`
}
