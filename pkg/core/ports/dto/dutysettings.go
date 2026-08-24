package dto

import "time"

type DutySettingsGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DormitoryID int64  `json:"dormitory_id"`
}

type DutySettingsTask struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Cost               int        `json:"cost"`
	RecurrenceInterval int        `json:"recurrenceInterval"`
	StartSequence      int        `json:"startSequence"`
	LastCompletedAt    *time.Time `json:"last_completed_at"`
	IsIncluded         bool       `json:"is_included"`
	AssigneeName       *string    `json:"assignee_name"`
	Status             string     `json:"status"`
}

type DutySettingsTaskSummary struct {
	TaskCount       int `json:"task_count"`
	TotalCost       int `json:"total_cost"`
	CostPerMember   int `json:"cost_per_member"`
	TeamMemberCount int `json:"team_member_count"`
}

type DutySettingsActiveDuty struct {
	ID        string                  `json:"id"`
	TeamID    string                  `json:"team_id"`
	TeamName  string                  `json:"team_name"`
	StartDate string                  `json:"start_date"`
	EndDate   string                  `json:"end_date"`
	Summary   DutySettingsTaskSummary `json:"summary"`
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

type DutySettingsRemoveTeamMemberRequest struct {
	ReplacementLeaderID *string `json:"replacement_leader_id"`
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
	Group            DutySettingsGroup       `json:"group"`
	Areas            []DutySettingsArea      `json:"areas"`
	Teams            []DutySettingsTeam      `json:"teams"`
	ActiveDutyTeamID *string                 `json:"active_duty_team_id"`
	TaskEditorState  string                  `json:"task_editor_state"`
	TaskEditorAlert  string                  `json:"task_editor_alert"`
	ActiveDuty       *DutySettingsActiveDuty `json:"active_duty"`
}

type ReorderDutySettingsTeamsRequest struct {
	TeamIDs []string `json:"team_ids"`
}
