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

type PenaltyItem struct {
	ID       string  `json:"id"`
	Reason   string  `json:"reason"`
	Weight   float64 `json:"weight"`
	IssuedOn string  `json:"issued_on"`
}

type PenaltyResidentDetailsResponse struct {
	UserID    string        `json:"user_id"`
	FullName  string        `json:"full_name"`
	Penalties []PenaltyItem `json:"penalties"`
}

type CreatePenaltyRequest struct {
	UserID   string  `json:"user_id"`
	Reason   string  `json:"reason"`
	Weight   float64 `json:"weight"`
	IssuedOn string  `json:"issued_on"`
}
