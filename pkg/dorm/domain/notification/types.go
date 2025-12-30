package notification

type EventType string

const (
	EventWeeklyIssue EventType = "weekly_issue"
	EventDormSummary EventType = "dorm_summary"
)

type Payload struct {
	EntityName string
	URL        string
	ExtraData  interface{}
}
