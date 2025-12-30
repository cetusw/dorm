package notification

import (
	"fmt"
	"strings"

	"dorm/pkg/dorm/application/service"
	"dorm/pkg/dorm/domain/notification"
	"dorm/pkg/dorm/infrastructure"

	"github.com/google/uuid"
)

type NotificationService struct {
	userService *service.UserService
	telegram    *infrastructure.Telegram
}

func NewNotificationService(userService *service.UserService, telegram *infrastructure.Telegram) *NotificationService {
	return &NotificationService{
		userService: userService,
		telegram:    telegram,
	}
}

func (s *NotificationService) Notify(
	userID uuid.UUID,
	event notification.EventType,
	data notification.Payload,
) error {
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		return err
	}
	if user == nil || user.TelegramID == 0 {
		return nil
	}

	msg := s.composeMessage(event, data)

	if _, err := s.telegram.SendMessage(user.TelegramID, msg); err != nil {
		fmt.Printf("failed to send notification: %v\n", err)
	}

	return nil
}

func (s *NotificationService) composeMessage(event notification.EventType, data notification.Payload) string {
	URL := ""
	if data.URL != "" {
		URL = fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", data.URL)
	}

	switch event {
	case notification.EventWeeklyIssue:
		return fmt.Sprintf("⚠️ По текущей уборке есть задачи, требующие внимания.\n%s", URL)
	case notification.EventDormSummary:
		return s.composeDormitoryLeadersNotification(data)
	default:
		return "Новое уведомление"
	}
}

func (s *NotificationService) composeDormitoryLeadersNotification(data notification.Payload) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚠️ В общежитии \"%s\" есть задачи:\n\n", data.EntityName))
	if groups, ok := data.ExtraData.([]notification.Payload); ok {
		for _, group := range groups {
			spreadsheetURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", group.URL)
			sb.WriteString(fmt.Sprintf("\"%s\" — %s\n\n", group.EntityName, spreadsheetURL))
		}
	}

	return sb.String()
}
