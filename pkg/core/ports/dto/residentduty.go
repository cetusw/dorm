package dto

const (
	ResidentDutyTaskStatusFree      = "free"
	ResidentDutyTaskStatusAssigned  = "assigned"
	ResidentDutyTaskStatusCompleted = "completed"
	ResidentDutyTaskStatusVerified  = "verified"
)

type ResidentCurrentDutyResponse struct {
	DormitoryID           int64                     `json:"dormitory_id"`
	MyDormitoryID         *int64                    `json:"my_dormitory_id,omitempty"`
	SelectedGroupID       string                    `json:"selected_group_id"`
	HasActiveDuty         bool                      `json:"has_active_duty"`
	CanManageTasks        bool                      `json:"can_manage_tasks"`
	CanManageDutySettings bool                      `json:"can_manage_duty_settings"`
	ReadOnly              bool                      `json:"read_only"`
	ShowGroupSelect       bool                      `json:"show_group_select"`
	VisibleTabs           []string                  `json:"visible_tabs"`
	NoticeMessage         string                    `json:"notice_message"`
	NoticeTone            string                    `json:"notice_tone"`
	PeriodStatus          string                    `json:"period_status"`
	Groups                []ResidentDutyGroupOption `json:"groups"`
	MyGroup               *ResidentDutyGroupOption  `json:"my_group,omitempty"`
	DutyID                string                    `json:"duty_id"`
	Group                 string                    `json:"group"`
	Team                  string                    `json:"team"`
	StartDate             string                    `json:"start_date"`
	EndDate               string                    `json:"end_date"`
	CostPerResidentGoal   int                       `json:"cost_per_resident_goal"`
	MyTakenCostSum        int                       `json:"my_taken_cost_sum"`
	TeamMembers           []ResidentDutyTeamMember  `json:"team_members"`
	Tasks                 []ResidentDutyTask        `json:"tasks"`
}

type ResidentDutyDetailsResponse = ResidentCurrentDutyResponse

type ResidentDutyHistoryResponse struct {
	SelectedGroupID string                    `json:"selected_group_id"`
	ShowGroupSelect bool                      `json:"show_group_select"`
	Groups          []ResidentDutyGroupOption `json:"groups"`
	Duties          []ResidentDutyHistoryItem `json:"duties"`
}

type ResidentDutyHistoryItem struct {
	ID             string                      `json:"id"`
	StartDate      string                      `json:"start_date"`
	EndDate        string                      `json:"end_date"`
	TeamLeaderName string                      `json:"team_leader_name"`
	PeriodStatus   string                      `json:"period_status"`
	Progress       ResidentDutyProgressSummary `json:"progress"`
}

type ResidentDutyProgressSummary struct {
	TotalCostSum        int `json:"total_cost_sum"`
	TakenCostSum        int `json:"taken_cost_sum"`
	TotalTasksCount     int `json:"total_tasks_count"`
	TakenTasksCount     int `json:"taken_tasks_count"`
	CompletedTasksCount int `json:"completed_tasks_count"`
	VerifiedTasksCount  int `json:"verified_tasks_count"`
}

type ResidentDutyGroupOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ResidentDutyTeamMember struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ResidentDutyTask struct {
	ID           string  `json:"id"`
	AreaID       int     `json:"area_id"`
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
