package duty

type UserStats struct {
	TotalPoints     int
	ConfirmedPoints int
	RequiredPoints  float64
}

func (s UserStats) IsQuotaMet() bool {
	return float64(s.ConfirmedPoints) >= s.RequiredPoints
}

func (s UserStats) IsQuotaCovered() bool {
	return float64(s.TotalPoints) >= s.RequiredPoints
}
