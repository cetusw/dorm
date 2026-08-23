package dto

type DutyParticipantResponse struct {
	ParticipantID string  `json:"participant_id"`
	FullName      string  `json:"full_name"`
	Type          string  `json:"type"`
	ExcludedAt    *string `json:"excluded_at"`
	IsLeader      bool    `json:"is_leader"`
	TeamID        *string `json:"team_id,omitempty"`
	TeamName      *string `json:"team_name,omitempty"`
}

type DutyParticipantCandidateResponse struct {
	ParticipantID string  `json:"participant_id"`
	FullName      string  `json:"full_name"`
	TeamID        *string `json:"team_id,omitempty"`
	TeamName      *string `json:"team_name,omitempty"`
}
