package notification

import (
	"context"
	"fmt"

	domainevents "dorm/pkg/core/domain/events"
)

type WeekStartedHandler struct {
	reminders *DutyReminderService
}

func NewWeekStartedHandler(reminders *DutyReminderService) *WeekStartedHandler {
	return &WeekStartedHandler{reminders: reminders}
}

func (h *WeekStartedHandler) Handle(ctx context.Context, event interface{}) error {
	startedEvent, ok := event.(domainevents.WeekStartedEvent)
	if !ok {
		return fmt.Errorf("unexpected event type %T", event)
	}

	return h.reminders.NotifyDutyStartedForDuties(ctx, startedEvent.Duties)
}
