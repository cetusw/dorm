package notification

import (
	"context"
	"fmt"

	domainevents "dorm/pkg/core/domain/events"
	domain "dorm/pkg/core/domain/notification"
	"dorm/pkg/core/ports"
)

type TasksReadyForReviewHandler struct {
	notifications ports.UserNotificationUseCase
}

func NewTasksReadyForReviewHandler(notifications ports.UserNotificationUseCase) *TasksReadyForReviewHandler {
	return &TasksReadyForReviewHandler{notifications: notifications}
}

func (h *TasksReadyForReviewHandler) Handle(ctx context.Context, event interface{}) error {
	readyEvent, ok := event.(domainevents.TasksReadyForReviewEvent)
	if !ok {
		return fmt.Errorf("unexpected event type %T", event)
	}

	deduplicationKey, err := domain.BuildDutyNotificationDeduplicationKey(
		domain.NotificationTasksReadyForReview,
		readyEvent.DutyID,
		readyEvent.TeamHeadID,
	)
	if err != nil {
		return fmt.Errorf("build notification deduplication key: %w", err)
	}

	return h.notifications.NotifyUser(ctx, ports.NotifyUserCommand{
		UserID:           readyEvent.TeamHeadID,
		Type:             domain.NotificationTasksReadyForReview,
		Title:            "Все задачи выполнены",
		Body:             "Участники команды завершили уборку. Задачи можно проверить.",
		TargetURL:        "/app/current-duty?tab=review",
		DeduplicationKey: deduplicationKey,
	})
}
