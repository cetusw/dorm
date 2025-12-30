package service

import (
	"fmt"
	"log"
	"strings"

	"dorm/pkg/dorm/application/model"
	"dorm/pkg/dorm/infrastructure"
	"dorm/pkg/dorm/infrastructure/mysql/repository"
)

type NotificationService struct {
	telegram        *infrastructure.Telegram
	dormRepo        *repository.DormitoryRepository
	groupService    *GroupService
	dutyService     *DutyService
	dutyTaskService *DutyTaskService
	teamService     *TeamService
	userService     *UserService
}

func NewNotificationService(
	telegram *infrastructure.Telegram,
	dormRepo *repository.DormitoryRepository,
	groupService *GroupService,
	dutyService *DutyService,
	dutyTaskService *DutyTaskService,
	teamService *TeamService,
	userService *UserService,
) *NotificationService {
	return &NotificationService{
		telegram:        telegram,
		dormRepo:        dormRepo,
		groupService:    groupService,
		dutyService:     dutyService,
		dutyTaskService: dutyTaskService,
		teamService:     teamService,
		userService:     userService,
	}
}

func (s *NotificationService) SendWeeklyCleaningReport() error {
	groups, err := s.groupService.GetAllGroups()
	if err != nil {
		return fmt.Errorf("failed to fetch groups: %w", err)
	}

	dormIssues := make(map[int64][]model.Group)

	for _, group := range groups {
		hasIssues, err := s.groupHasIssues(group)
		if err != nil {
			log.Printf("Error checking issues for group %s: %v", group.Name, err)
			continue
		}

		if hasIssues {
			if err := s.notifyGroupLeader(group); err != nil {
				log.Printf("Failed to notify leader of group %s: %v", group.Name, err)
			}

			dormIssues[group.DormitoryID] = append(dormIssues[group.DormitoryID], group)
		}
	}

	if err := s.notifyDormitoryLeaders(dormIssues); err != nil {
		return err
	}

	return nil
}

func (s *NotificationService) notifyGroupLeader(group model.Group) error {
	if group.LeaderID == nil {
		return nil
	}

	leader, err := s.userService.GetUserByID(*group.LeaderID)
	if err != nil {
		return fmt.Errorf("failed to get group leader: %w", err)
	}
	if leader == nil || leader.TelegramID == 0 {
		return nil
	}

	link := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", group.SpreadsheetID)
	msg := fmt.Sprintf("⚠️ В группе \"%s\" есть задачи, требующие внимания.\n%s", group.Name, link)

	s.sendTelegramMessage(leader.TelegramID, msg)

	return nil
}

func (s *NotificationService) notifyDormitoryLeaders(dormIssues map[int64][]model.Group) error {
	dorms, err := s.dormRepo.FindAll()
	if err != nil {
		return fmt.Errorf("failed to fetch dormitories: %w", err)
	}

	for _, dorm := range dorms {
		problematicGroups, exists := dormIssues[dorm.ID]
		if !exists || len(problematicGroups) == 0 {
			continue
		}

		if dorm.LeaderID == nil {
			continue
		}

		leader, err := s.userService.GetUserByID(*dorm.LeaderID)
		if err != nil {
			log.Printf("Failed to get dorm leader for dorm %s: %v", dorm.Name, err)
			continue
		}

		if leader == nil || leader.TelegramID == 0 {
			continue
		}

		var msgBuilder strings.Builder
		msgBuilder.WriteString(fmt.Sprintf("⚠️ В общежитии \"%s\" есть задачи, требующие внимания:</b>\n\n", dorm.Name))

		for _, grp := range problematicGroups {
			link := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", grp.SpreadsheetID)
			msgBuilder.WriteString(fmt.Sprintf("\"%s\" Следует обратить внимание на задачи: %s\n\n", grp.Name, link))
		}

		s.sendTelegramMessage(leader.TelegramID, msgBuilder.String())
	}
	return nil
}

func (s *NotificationService) groupHasIssues(group model.Group) (bool, error) {
	duty, err := s.dutyService.GetGroupLastDuty(group.ID)
	if err != nil {
		return false, nil
	}
	if duty == nil {
		return false, nil
	}

	tasks, err := s.dutyTaskService.GetDutyTasksView(duty.ID)
	if err != nil {
		return false, err
	}

	for _, t := range tasks {
		if !s.isTaskCompletedCorrectly(t) {
			return true, nil
		}
	}

	return false, nil
}

func (s *NotificationService) isTaskCompletedCorrectly(t model.DutyTaskView) bool {
	isDone := t.CompletionDate != nil
	hasAssignee := t.AssigneeID != nil

	return isDone && hasAssignee
}

func (s *NotificationService) sendTelegramMessage(chatID int64, text string) {
	_, err := s.telegram.SendMessage(chatID, text)
	if err != nil {
		log.Printf("Failed to send notification to user %d: %v", chatID, err)
	}
}
