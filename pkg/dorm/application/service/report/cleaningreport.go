package report

import (
	"fmt"
	"log"

	"dorm/pkg/dorm/application/model"
	"dorm/pkg/dorm/application/service"
	"dorm/pkg/dorm/application/service/notification"
	domainnotification "dorm/pkg/dorm/domain/notification"
	"dorm/pkg/dorm/infrastructure/mysql/repository"

	"github.com/google/uuid"
)

type CleaningReportService struct {
	cleaningService *service.CleaningService
	groupService    *service.GroupService
	teamService     *service.TeamService
	dormRepo        *repository.DormitoryRepository
	dutyService     *service.DutyService
	dutyTaskService *service.DutyTaskService
	notifier        *notification.NotificationService
}

func NewCleaningReportService(
	cleaningService *service.CleaningService,
	groupService *service.GroupService,
	teamService *service.TeamService,
	dormRepo *repository.DormitoryRepository,
	dutyService *service.DutyService,
	dutyTaskService *service.DutyTaskService,
	notifier *notification.NotificationService,
) *CleaningReportService {
	return &CleaningReportService{
		cleaningService: cleaningService,
		groupService:    groupService,
		teamService:     teamService,
		dormRepo:        dormRepo,
		dutyService:     dutyService,
		dutyTaskService: dutyTaskService,
		notifier:        notifier,
	}
}

func (s *CleaningReportService) ProcessWeeklyReports() error {
	groups, err := s.groupService.GetAllGroups()
	if err != nil {
		return fmt.Errorf("fetch groups: %w", err)
	}

	dormIssues := make(map[int64][]domainnotification.Payload)

	for _, group := range groups {
		duty, err := s.dutyService.GetGroupLastDuty(group.ID)
		if err != nil {
			return err
		}
		if duty == nil {
			return nil
		}

		hasIssues, err := s.GetDutyStatus(duty.ID)
		if err != nil {
			log.Printf("Error checking group %s: %v", group.Name, err)
			continue
		}

		if hasIssues {
			payload := domainnotification.Payload{
				EntityName: group.Name,
				URL:        group.SpreadsheetID,
			}

			s.notifyLocalLeaders(payload, group, duty.TeamID)

			dormIssues[group.DormitoryID] = append(dormIssues[group.DormitoryID], payload)
		}
	}

	return s.notifyDormLeaders(dormIssues)
}

func (s *CleaningReportService) GetDutyStatus(dutyID uuid.UUID) (bool, error) {
	hasIssues, err := s.checkDutyIssues(dutyID)
	if err != nil {
		return false, err
	}

	return hasIssues, nil
}

func (s *CleaningReportService) checkDutyIssues(dutyID uuid.UUID) (bool, error) {
	tasks, err := s.dutyTaskService.GetDutyTasksView(dutyID)
	if err != nil {
		return false, err
	}
	for _, t := range tasks {
		if t.CompletionDate == nil || t.AssigneeID == nil {
			return true, nil
		}
	}
	return false, nil
}

func (s *CleaningReportService) notifyLocalLeaders(payload domainnotification.Payload, group model.Group, teamID uuid.UUID) {
	if team, err := s.teamService.GetTeam(teamID); err == nil {
		s.notifier.Notify(team.LeaderID, domainnotification.EventWeeklyIssue, payload)
	}

	if group.LeaderID != nil {
		s.notifier.Notify(*group.LeaderID, domainnotification.EventWeeklyIssue, payload)
	}
}

func (s *CleaningReportService) notifyDormLeaders(issuesMap map[int64][]domainnotification.Payload) error {
	dorms, err := s.dormRepo.FindAll()
	if err != nil {
		return err
	}

	for _, dorm := range dorms {
		payloads, exists := issuesMap[dorm.ID]
		if !exists || len(payloads) == 0 || dorm.LeaderID == nil {
			continue
		}

		summaryPayload := domainnotification.Payload{
			EntityName: dorm.Name,
			ExtraData:  payloads,
		}

		s.notifier.Notify(*dorm.LeaderID, domainnotification.EventDormSummary, summaryPayload)
	}
	return nil
}
