package notification

import (
	"context"
	"fmt"
	"log"
	"time"

	domain "dorm/pkg/core/domain/notification"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	queryports "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
)

type DutyReminderService struct {
	activeDutyQuery         queryports.ActiveDutyQuery
	dutyTeamMembersQuery    queryports.DutyTeamMembersQuery
	dutyFinishReminderQuery queryports.DutyFinishReminderQuery
	notifications           ports.UserNotificationUseCase
}

func NewDutyReminderService(
	activeDutyQuery queryports.ActiveDutyQuery,
	dutyTeamMembersQuery queryports.DutyTeamMembersQuery,
	dutyFinishReminderQuery queryports.DutyFinishReminderQuery,
	notifications ports.UserNotificationUseCase,
) *DutyReminderService {
	return &DutyReminderService{
		activeDutyQuery:         activeDutyQuery,
		dutyTeamMembersQuery:    dutyTeamMembersQuery,
		dutyFinishReminderQuery: dutyFinishReminderQuery,
		notifications:           notifications,
	}
}

func (s *DutyReminderService) SendDutyStartedReminders(ctx context.Context, at time.Time) error {
	return s.notifyForActiveDutyMembers(
		ctx,
		at,
		domain.NotificationDutyStarted,
		"Дежурная неделя",
		"На этой неделе дежурит ваша команда. Ознакомьтесь со списком задач.",
		"/app/current-duty",
	)
}

func (s *DutyReminderService) SendSaturdayTaskReminders(ctx context.Context, at time.Time) error {
	return s.notifyForActiveDutyMembers(
		ctx,
		at,
		domain.NotificationTakeTasksSaturdayReminder,
		"Завтра уборка",
		"Возьмите задачи, которые вы планируете выполнить.",
		"/app/current-duty?tab=free",
	)
}

func (s *DutyReminderService) SendSundayTakeTaskReminders(ctx context.Context, at time.Time) error {
	activeDuties, err := s.activeDutyQuery.FindAllActive(ctx, at)
	if err != nil {
		return fmt.Errorf("find active duties: %w", err)
	}

	for _, activeDuty := range activeDuties {
		state, err := s.dutyFinishReminderQuery.GetByDutyID(ctx, activeDuty.ID, activeDuty.TeamID)
		if err != nil {
			log.Printf("notification reminder: load sunday take state for duty %s: %v", activeDuty.ID, err)
			continue
		}

		if state.FreeTaskCount == 0 || len(state.UsersBelowAssignedGoalIDs) == 0 {
			continue
		}

		s.notifyDutyUsers(
			ctx,
			activeDuty,
			state.UsersBelowAssignedGoalIDs,
			domain.NotificationTakeTasksSundayReminder,
			"Возьмите задачи",
			"Сегодня уборка. У вас пока недостаточно задач по баллам, а в списке еще есть свободные.",
			"/app/current-duty?tab=free",
		)
	}

	return nil
}

func (s *DutyReminderService) SendSundayFinishTaskReminders(ctx context.Context, at time.Time) error {
	activeDuties, err := s.activeDutyQuery.FindAllActive(ctx, at)
	if err != nil {
		return fmt.Errorf("find active duties: %w", err)
	}

	for _, activeDuty := range activeDuties {
		state, err := s.dutyFinishReminderQuery.GetByDutyID(ctx, activeDuty.ID, activeDuty.TeamID)
		if err != nil {
			log.Printf("notification reminder: load finish state for duty %s: %v", activeDuty.ID, err)
			continue
		}

		recipientSet := make(map[uuid.UUID]struct{})
		for _, userID := range state.UsersWithIncompleteTaskIDs {
			recipientSet[userID] = struct{}{}
		}

		if state.FreeTaskCount > 0 {
			for _, userID := range state.UsersBelowAssignedGoalIDs {
				recipientSet[userID] = struct{}{}
			}
		}

		if len(recipientSet) == 0 {
			continue
		}

		userIDs := make([]uuid.UUID, 0, len(recipientSet))
		for userID := range recipientSet {
			userIDs = append(userIDs, userID)
		}

		s.notifyDutyUsers(
			ctx,
			activeDuty,
			userIDs,
			domain.NotificationFinishTasksSundayReminder,
			"Завершите уборку",
			"В дежурстве остались невыполненные задачи. Возьмите и завершите их.",
			"/app/current-duty",
		)
	}

	return nil
}

func (s *DutyReminderService) NotifyDutyStartedForDuties(ctx context.Context, duties []dto.DutyViewModel) error {
	for _, dutyView := range duties {
		for _, userStats := range dutyView.UsersStats {
			if userStats == nil {
				continue
			}

			deduplicationKey, err := domain.BuildDutyNotificationDeduplicationKey(
				domain.NotificationDutyStarted,
				dutyView.DutyID,
				userStats.ID,
			)
			if err != nil {
				log.Printf("notification reminder: build duty started deduplication key for duty %s user %s: %v", dutyView.DutyID, userStats.ID, err)
				continue
			}

			err = s.notifications.NotifyUser(ctx, ports.NotifyUserCommand{
				UserID:           userStats.ID,
				Type:             domain.NotificationDutyStarted,
				Title:            "Дежурная неделя",
				Body:             "На этой неделе дежурит ваша команда. Ознакомьтесь со списком задач.",
				TargetURL:        "/app/current-duty",
				DeduplicationKey: deduplicationKey,
			})
			if err != nil {
				log.Printf("notification reminder: notify duty started for duty %s user %s: %v", dutyView.DutyID, userStats.ID, err)
			}
		}
	}

	return nil
}

func (s *DutyReminderService) notifyForActiveDutyMembers(
	ctx context.Context,
	at time.Time,
	notificationType domain.NotificationType,
	title string,
	body string,
	targetURL string,
) error {
	activeDuties, err := s.activeDutyQuery.FindAllActive(ctx, at)
	if err != nil {
		return fmt.Errorf("find active duties: %w", err)
	}

	for _, activeDuty := range activeDuties {
		userIDs, err := s.dutyTeamMembersQuery.FindUserIDsByTeamID(ctx, activeDuty.TeamID)
		if err != nil {
			log.Printf("notification reminder: load team members for duty %s: %v", activeDuty.ID, err)
			continue
		}

		s.notifyDutyUsers(ctx, activeDuty, userIDs, notificationType, title, body, targetURL)
	}

	return nil
}

func (s *DutyReminderService) notifyDutyUsers(
	ctx context.Context,
	activeDuty queryports.ActiveDuty,
	userIDs []uuid.UUID,
	notificationType domain.NotificationType,
	title string,
	body string,
	targetURL string,
) {
	for _, userID := range userIDs {
		deduplicationKey, err := domain.BuildDutyNotificationDeduplicationKey(notificationType, activeDuty.ID, userID)
		if err != nil {
			log.Printf("notification reminder: build deduplication key for duty %s user %s: %v", activeDuty.ID, userID, err)
			continue
		}

		err = s.notifications.NotifyUser(ctx, ports.NotifyUserCommand{
			UserID:           userID,
			Type:             notificationType,
			Title:            title,
			Body:             body,
			TargetURL:        targetURL,
			DeduplicationKey: deduplicationKey,
		})
		if err != nil {
			log.Printf("notification reminder: notify user %s for duty %s: %v", userID, activeDuty.ID, err)
		}
	}
}
