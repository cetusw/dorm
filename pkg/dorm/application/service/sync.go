package service

import (
	"fmt"
	"log"

	"github.com/google/uuid"

	"dorm/pkg/common/consts"
	"dorm/pkg/dorm/application/model"
	"dorm/pkg/dorm/application/service/sheets"
	"dorm/pkg/dorm/infrastructure"
)

type SyncService struct {
	infraSheets     *infrastructure.Sheets
	groupService    *GroupService
	dutyService     *DutyService
	dutyTaskService *DutyTaskService
	userService     *UserService
	parser          *sheets.DutySheetParser
}

func NewSyncService(
	infraSheets *infrastructure.Sheets,
	groupService *GroupService,
	dutyService *DutyService,
	dutyTaskService *DutyTaskService,
	userService *UserService,
) *SyncService {
	return &SyncService{
		infraSheets:     infraSheets,
		groupService:    groupService,
		dutyService:     dutyService,
		dutyTaskService: dutyTaskService,
		userService:     userService,
		parser:          sheets.NewDutySheetParser(),
	}
}

func (s *SyncService) SyncAllActiveDuties() error {
	groups, err := s.groupService.GetAllGroups()
	if err != nil {
		return err
	}

	for _, group := range groups {
		if err = s.syncGroupDuty(group); err != nil {
			log.Printf("Failed to sync group %s: %v", group.Name, err)
		}
	}
	return nil
}

func (s *SyncService) syncGroupDuty(group model.Group) error {
	duty, err := s.dutyService.GetGroupLastDuty(group.ID)
	if err != nil || duty == nil {
		return nil
	}

	dbTasks, err := s.dutyTaskService.GetDutyTasksView(duty.ID)
	if err != nil {
		return err
	}

	teamUsers, err := s.userService.GetUsersByTeamID(duty.TeamID)
	if err != nil {
		return err
	}

	sheetTitle := s.createSheetTitle(*duty)
	readRange := fmt.Sprintf("%s!%s", sheetTitle, consts.TasksTableRange)

	data, err := s.infraSheets.ReadSheet(group.SpreadsheetID, readRange)
	if err != nil {
		return fmt.Errorf("read sheet error: %w", err)
	}

	parsedRows, err := s.parser.Parse(data)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	return s.applyUpdates(duty.ID, dbTasks, parsedRows, teamUsers)
}

func (s *SyncService) applyUpdates(
	dutyID uuid.UUID,
	dbTasks []model.DutyTaskView,
	sheetRows []sheets.ParsedTaskDTO,
	users []model.User,
) error {
	dbTaskMap := make(map[string]model.DutyTaskView)
	for _, t := range dbTasks {
		areaFullName := fmt.Sprintf("%d этаж. %s", t.AreaFloor, t.AreaName)
		key := areaFullName + "|" + t.TaskTitle
		dbTaskMap[key] = t
	}

	for _, row := range sheetRows {
		key := row.AreaName + "|" + row.TaskTitle
		dbTask, exists := dbTaskMap[key]
		if !exists {
			continue
		}

		newAssigneeID := s.resolveUser(row.AssigneeName, users)

		currentAssigneeID := dbTask.AssigneeID

		needsAssignUpdate := false
		if newAssigneeID != nil {
			if currentAssigneeID == nil || *currentAssigneeID != *newAssigneeID {
				needsAssignUpdate = true
			}
		} else if row.AssigneeName == consts.DefaultAssignee && currentAssigneeID != nil {
			needsAssignUpdate = true
		}

		if needsAssignUpdate {
			err := s.dutyTaskService.SetDutyTaskAssigneeID(newAssigneeID, dbTask.TaskID, dutyID)
			if err != nil {
				log.Printf("Error updating assignee for task %s: %v", dbTask.TaskTitle, err)
			}
		}

		isDoneInSheet := row.StatusRaw == consts.StateDone || row.StatusRaw == consts.StateVerified
		isDoneInDB := dbTask.CompletionDate != nil

		if isDoneInSheet && !isDoneInDB {
			err := s.dutyTaskService.CompleteDutyTaskByDutyIDAndTaskID(dutyID, dbTask.TaskID)
			if err != nil {
				log.Printf("Error completing task %s: %v", dbTask.TaskTitle, err)
			}
		}
	}

	return nil
}

func (s *SyncService) resolveUser(shortName string, users []model.User) *uuid.UUID {
	if shortName == consts.DefaultAssignee || shortName == "" {
		return nil
	}

	for _, u := range users {
		generatedName := fmt.Sprintf("%s %s.", u.FirstName, string([]rune(u.LastName)[0]))

		if generatedName == shortName {
			uid := u.ID
			return &uid
		}
	}
	return nil
}

func (s *SyncService) createSheetTitle(duty model.Duty) string {
	// TODO: исправить часовые зоны DORM-17
	start := duty.Start.Add(10800000000000)
	end := duty.End.Add(10800000000000)
	return fmt.Sprintf("%s-%s", start.Format("02.01"), end.Format("02.01"))
}
