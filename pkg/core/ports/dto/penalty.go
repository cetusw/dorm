package dto

type PenaltyResidentSummary struct {
	UserID           string  `json:"user_id"`
	FullName         string  `json:"full_name"`
	TotalWeight      float64 `json:"total_weight"`
	ThresholdReached bool    `json:"threshold_reached"`
}

type PenaltyResidentsResponse struct {
	Residents []PenaltyResidentSummary `json:"residents"`
}

type PenaltyResidentOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PenaltyResidentOptionsResponse struct {
	Residents []PenaltyResidentOption `json:"residents"`
}

type PenaltyEntryItem struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Reason    string  `json:"reason"`
	Weight    float64 `json:"weight"`
	CreatedAt string  `json:"created_at"`
}

type PenaltyResidentDetailsResponse struct {
	UserID      string             `json:"user_id"`
	FullName    string             `json:"full_name"`
	TotalWeight float64            `json:"total_weight"`
	Entries     []PenaltyEntryItem `json:"entries"`
}

type CurrentUserPenaltiesResponse struct {
	TotalWeight float64            `json:"total_weight"`
	Entries     []PenaltyEntryItem `json:"entries"`
}

type CreatePenaltyRequest struct {
	UserID string  `json:"user_id"`
	Reason string  `json:"reason"`
	Weight float64 `json:"weight"`
}

type ResolvePenaltyRequest struct {
	UserID string  `json:"user_id"`
	Reason string  `json:"reason"`
	Weight float64 `json:"weight"`
}

type UpdatePenaltyEntryRequest struct {
	Reason string  `json:"reason"`
	Weight float64 `json:"weight"`
}
