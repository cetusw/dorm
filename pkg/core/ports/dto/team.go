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
}

type TeamMemberItem struct {
	UserID          uuid.UUID
	FullName        string
	CurrentTeamID   *uuid.UUID
	CurrentTeamName string
	IsInCurrentTeam bool
}

type CreateTeamRequest struct {
	Name    string `form:"name"`
	Color   string `form:"color"`
	Order   int    `form:"order"`
	GroupID string `form:"group_id"`
}

type UpdateTeamRequest struct {
	Name    string `form:"name"`
	Color   string `form:"color"`
	Order   int    `form:"order"`
	GroupID string `form:"group_id"`
}

