package ports

type DutyTaskProgress struct {
	TotalCount     int
	CompletedCount int
	VerifiedCount  int
}

func (p DutyTaskProgress) IsReadyForReview() bool {
	return p.TotalCount > 0 &&
		p.CompletedCount == p.TotalCount &&
		p.VerifiedCount < p.TotalCount
}
