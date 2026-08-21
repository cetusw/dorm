package dto

type IndividualTaskRequest struct {
	ResidentID       string  `json:"resident_id"`
	Title            string  `json:"title"`
	AreaID           *int    `json:"area_id"`
	RedemptionWeight float64 `json:"redemption_weight"`
	Deadline         *string `json:"deadline"`
	Version          uint64  `json:"version"`
}
type IndividualTaskArea struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Floor *int   `json:"floor"`
}
type IndividualTaskResident struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type IndividualTaskItem struct {
	ID               string                 `json:"id"`
	DormitoryID      int64                  `json:"dormitory_id"`
	Resident         IndividualTaskResident `json:"resident"`
	Title            string                 `json:"title"`
	Area             *IndividualTaskArea    `json:"area"`
	RedemptionWeight float64                `json:"redemption_weight"`
	Deadline         *string                `json:"deadline"`
	Status           string                 `json:"status"`
	IsOverdue        bool                   `json:"is_overdue"`
	CompletedAt      *string                `json:"completed_at"`
	VerifiedAt       *string                `json:"verified_at"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	Version          uint64                 `json:"version"`
	CanEdit          bool                   `json:"can_edit"`
	CanDelete        bool                   `json:"can_delete"`
	CanComplete      bool                   `json:"can_complete"`
	CanVerify        bool                   `json:"can_verify"`
	CanReject        bool                   `json:"can_reject"`
}
type IndividualTaskListResponse struct {
	Tasks []IndividualTaskItem `json:"tasks"`
}
type IndividualTaskResidentsResponse struct {
	Residents []IndividualTaskResidentOption `json:"residents"`
}
type IndividualTaskResidentOption struct {
	ID                        string  `json:"id"`
	Name                      string  `json:"name"`
	DormitoryID               int64   `json:"dormitory_id"`
	PenaltyBalance            float64 `json:"penalty_balance"`
	ReservedRedemptionWeight  float64 `json:"reserved_redemption_weight"`
	AvailableRedemptionWeight float64 `json:"available_redemption_weight"`
}
type IndividualTaskAreasResponse struct {
	Areas []IndividualTaskArea `json:"areas"`
}
