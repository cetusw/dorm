package duty

type UserStats struct {
	TotalPoints     int
	ConfirmedPoints int
	RequiredPoints  float64
}

func (s UserStats) IsQuotaMet() bool {
	return float64(s.ConfirmedPoints) >= s.RequiredPoints
}
