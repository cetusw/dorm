package dto

import "github.com/google/uuid"

type TeamListItem struct {
	ID            uuid.UUID
	Name          string
	Color         string
	Order         int
	GroupID       uuid.UUID
	GroupName     string
	DormitoryID   int64
	DormitoryName string
	MembersCount  int
	LeaderID      *uuid.UUID
}

type TeamResponseItem struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Leader       *UserSummary `json:"leader"`
	MembersCount int          `json:"members_count"`
}

type TeamListResponse struct {
	Teams []TeamResponseItem `json:"teams"`
}

type TeamDetails struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	GroupID   string       `json:"group_id"`
	GroupName string       `json:"group_name"`
	Leader    *UserSummary `json:"leader"`
	MemberIDs []string     `json:"member_ids"`
}

type TeamMemberOptionItem struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	CurrentTeamID   *string `json:"current_team_id"`
	CurrentTeamName string  `json:"current_team_name"`
	IsInCurrentTeam bool    `json:"is_in_current_team"`
}

type TeamMemberOptionsResponse struct {
	Members []TeamMemberOptionItem `json:"members"`
}

type CreateResidentTeamRequest struct {
	Name      string   `json:"name"`
	GroupID   string   `json:"group_id"`
	LeaderID  *string  `json:"leader_id"`
	MemberIDs []string `json:"member_ids"`
}

type UpdateResidentTeamRequest struct {
	Name      string   `json:"name"`
	GroupID   string   `json:"group_id"`
	LeaderID  *string  `json:"leader_id"`
	MemberIDs []string `json:"member_ids"`
}

type TeamMemberItem struct {
	UserID          uuid.UUID
	FullName        string
	CurrentTeamID   *uuid.UUID
	CurrentTeamName string
	IsInCurrentTeam bool
}

type CreateTeamRequest struct {
	Name      string   `form:"name"`
	Color     string   `form:"color"`
	Order     int      `form:"order"`
	GroupID   string   `form:"group_id"`
	LeaderID  string   `form:"leader_id"`
	MemberIDs []string `form:"member_ids"`
}

type UpdateTeamRequest struct {
	Name      string   `form:"name"`
	Color     string   `form:"color"`
	Order     int      `form:"order"`
	GroupID   string   `form:"group_id"`
	LeaderID  string   `form:"leader_id"`
	MemberIDs []string `form:"member_ids"`
}
