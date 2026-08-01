package notification

type NotificationType string

const (
	NotificationDutyStarted               NotificationType = "duty_started"
	NotificationTakeTasksSaturdayReminder NotificationType = "take_tasks_saturday_reminder"
	NotificationTakeTasksSundayReminder   NotificationType = "take_tasks_sunday_reminder"
	NotificationFinishTasksSundayReminder NotificationType = "finish_tasks_sunday_reminder"
	NotificationTasksReadyForReview       NotificationType = "tasks_ready_for_review"
	NotificationDutyCompleted             NotificationType = "duty_completed"
)

func (notificationType NotificationType) IsValid() bool {
	switch notificationType {
	case NotificationDutyStarted,
		NotificationTakeTasksSaturdayReminder,
		NotificationTakeTasksSundayReminder,
		NotificationFinishTasksSundayReminder,
		NotificationTasksReadyForReview,
		NotificationDutyCompleted:
		return true
	default:
		return false
	}
}
