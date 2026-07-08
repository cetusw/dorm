package dto

const (
	ResidentDutyTaskStatusFree      = "free"
	ResidentDutyTaskStatusAssigned  = "assigned"
	ResidentDutyTaskStatusCompleted = "completed"
	ResidentDutyTaskStatusVerified  = "verified"
)

type ResidentCurrentDutyResponse struct {
	DutyID              string             `json:"duty_id"`
	Group               string             `json:"group"`
	Team                string             `json:"team"`
	StartDate           string             `json:"start_date"`
	EndDate             string             `json:"end_date"`
	CostPerResidentGoal int                `json:"cost_per_resident_goal"`
	MyTakenCostSum      int                `json:"my_taken_cost_sum"`
	Tasks               []ResidentDutyTask `json:"tasks"`
}

type ResidentDutyTask struct {
	ID           string  `json:"id"`
	AreaName     string  `json:"area_name"`
	AreaFloor    int     `json:"area_floor"`
	Title        string  `json:"title"`
	Cost         int     `json:"cost"`
	Status       string  `json:"status"`
	AssigneeID   *string `json:"assignee_id,omitempty"`
	AssigneeName *string `json:"assignee_name,omitempty"`
	IsMine       bool    `json:"is_mine"`

	CanTake     bool `json:"can_take"`
	CanReturn   bool `json:"can_return"`
	CanComplete bool `json:"can_complete"`
	CanOpen     bool `json:"can_open"`
}
