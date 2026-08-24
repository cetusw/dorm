package dto

import "github.com/google/uuid"

type TeamResponseItem struct {
	ID           string       `json:"id"`
	Leader       *UserSummary `json:"leader"`
	MembersCount int          `json:"members_count"`
}

type TeamListResponse struct {
	Teams []TeamResponseItem `json:"teams"`
}

type TeamDetails struct {
	ID        string       `json:"id"`
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
	IsTeamLeader    bool    `json:"is_team_leader"`
}

type TeamMemberOptionsResponse struct {
	Members []TeamMemberOptionItem `json:"members"`
}

type CreateResidentTeamRequest struct {
	GroupID   string   `json:"group_id"`
	LeaderID  string   `json:"leader_id"`
	MemberIDs []string `json:"member_ids"`
}

type UpdateResidentTeamRequest struct {
	GroupID   string   `json:"group_id"`
	LeaderID  string   `json:"leader_id"`
	MemberIDs []string `json:"member_ids"`
}

type TeamMemberItem struct {
	UserID          uuid.UUID
	FullName        string
	CurrentTeamID   *uuid.UUID
	CurrentTeamName string
	IsInCurrentTeam bool
	IsTeamLeader    bool
}
