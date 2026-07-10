package dto

const (
	ResidentDutyTaskStatusFree      = "free"
	ResidentDutyTaskStatusAssigned  = "assigned"
	ResidentDutyTaskStatusCompleted = "completed"
	ResidentDutyTaskStatusVerified  = "verified"
)

type ResidentCurrentDutyResponse struct {
	SelectedGroupID     string                    `json:"selected_group_id"`
	HasActiveDuty       bool                      `json:"has_active_duty"`
	CanManageTasks      bool                      `json:"can_manage_tasks"`
	ReadOnly            bool                      `json:"read_only"`
	ShowGroupSelect     bool                      `json:"show_group_select"`
	VisibleTabs         []string                  `json:"visible_tabs"`
	NoticeMessage       string                    `json:"notice_message"`
	Groups              []ResidentDutyGroupOption `json:"groups"`
	DutyID              string                    `json:"duty_id"`
	Group               string                    `json:"group"`
	Team                string                    `json:"team"`
	StartDate           string                    `json:"start_date"`
	EndDate             string                    `json:"end_date"`
	CostPerResidentGoal int                       `json:"cost_per_resident_goal"`
	MyTakenCostSum      int                       `json:"my_taken_cost_sum"`
	Tasks               []ResidentDutyTask        `json:"tasks"`
}

type ResidentDutyGroupOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

	CanTake       bool `json:"can_take"`
	CanReturn     bool `json:"can_return"`
	CanComplete   bool `json:"can_complete"`
	CanOpen       bool `json:"can_open"`
	CanVerify     bool `json:"can_verify"`
	CanReviewOpen bool `json:"can_review_open"`
}
