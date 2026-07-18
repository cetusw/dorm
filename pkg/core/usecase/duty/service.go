package duty

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"dorm/pkg/core/domain/catalog"
	dutydomain "dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type Service struct {
	dutyRepo     dutydomain.DutyRepository
	teamRepo     structure.TeamRepository
	groupRepo    structure.GroupRepository
	dormRepo     structure.DormitoryRepository
	taskRepo     catalog.TaskDefinitionRepository
	overrideRepo catalog.DutyTaskOverrideRepository
	areaRepo     catalog.AreaRepository
	userRepo     user.Repository
}

func NewDutyService(
	dutyRepo dutydomain.DutyRepository,
	teamRepo structure.TeamRepository,
	groupRepo structure.GroupRepository,
	dormRepo structure.DormitoryRepository,
	taskRepo catalog.TaskDefinitionRepository,
	overrideRepo catalog.DutyTaskOverrideRepository,
	areaRepo catalog.AreaRepository,
	userRepo user.Repository,
) *Service {
	return &Service{
		dutyRepo:     dutyRepo,
		teamRepo:     teamRepo,
		groupRepo:    groupRepo,
		dormRepo:     dormRepo,
		taskRepo:     taskRepo,
		overrideRepo: overrideRepo,
		areaRepo:     areaRepo,
		userRepo:     userRepo,
	}
}

func (s *Service) GetGroupDuties(ctx context.Context, groupID uuid.UUID) ([]dto.DutyListItem, error) {
	duties, err := s.dutyRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	items := make([]dto.DutyListItem, 0, len(duties))
	for _, d := range duties {
		team, group, dormitory, err := s.structureContext(ctx, d.TeamID())
		if err != nil {
			return nil, err
		}
		if team == nil || group == nil || dormitory == nil {
			continue
		}
		if group.ID() != groupID {
			continue
		}
		items = append(items, dto.DutyListItem{
			ID:            d.ID(),
			StartDate:     d.Start(),
			EndDate:       d.End(),
			DormitoryID:   dormitory.ID(),
			DormitoryName: dormitory.Name(),
			GroupID:       group.ID(),
			GroupName:     group.Name(),
			TeamName:      team.Name(),
		})
	}
	return items, nil
}

func (s *Service) GetDutyDetail(ctx context.Context, groupID uuid.UUID, dutyID uuid.UUID) (*dto.DutyDetailItem, error) {
	d, err := s.dutyRepo.FindByID(ctx, dutyID)
	if err != nil || d == nil {
		return nil, err
	}

	team, group, dormitory, err := s.structureContext(ctx, d.TeamID())
	if err != nil {
		return nil, err
	}
	if team == nil || group == nil || dormitory == nil {
		return nil, nil
	}
	if group.ID() != groupID {
		return nil, nil
	}

	taskMap, areaMap, userMap, err := s.detailLookups(ctx)
	if err != nil {
		return nil, err
	}

	tasks := make([]dto.DutyTaskListItem, 0, len(d.Tasks()))
	for _, dutyTask := range d.Tasks() {
		def := taskMap[dutyTask.TaskDefID()]
		if def == nil {
			continue
		}
		area := areaMap[def.AreaID()]
		areaName := ""
		if area != nil {
			areaName = area.Name()
		}

		assigneeName := "Не назначен"
		if dutyTask.AssigneeID() != nil {
			if assignee := userMap[*dutyTask.AssigneeID()]; assignee != nil {
				assigneeName = fullName(assignee)
			}
		}

		tasks = append(tasks, dto.DutyTaskListItem{
			AreaName:     areaName,
			Title:        def.Title(),
			AssigneeName: assigneeName,
			Status:       taskStatus(dutyTask),
		})
	}

	return &dto.DutyDetailItem{
		ID:            d.ID(),
		StartDate:     d.Start(),
		EndDate:       d.End(),
		DormitoryID:   dormitory.ID(),
		DormitoryName: dormitory.Name(),
		GroupID:       group.ID(),
		GroupName:     group.Name(),
		TeamName:      team.Name(),
		Tasks:         tasks,
	}, nil
}

func (s *Service) GetFutureDutyTasks(ctx context.Context, groupID uuid.UUID) ([]dto.FutureDutyTaskGroup, error) {
	tasks, err := s.taskRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return s.groupFutureDutyTasks(ctx, tasks, time.Now())
}

func (s *Service) GetCommonFutureDutyTasks(ctx context.Context) ([]dto.FutureDutyTaskGroup, error) {
	tasks, err := s.taskRepo.FindCommon(ctx)
	if err != nil {
		return nil, err
	}
	return s.groupFutureDutyTasks(ctx, tasks, time.Now())
}

func (s *Service) groupFutureDutyTasks(ctx context.Context, tasks []*catalog.TaskDefinition, referenceDate time.Time) ([]dto.FutureDutyTaskGroup, error) {
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, err
	}
	overrides, err := s.overrideRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	overrideMap := make(map[uuid.UUID]bool, len(overrides))
	for _, override := range overrides {
		overrideMap[override.TaskID()] = override.IncludeInNextDuty()
	}

	areaMap := make(map[int]*catalog.Area, len(areas))
	for _, area := range areas {
		areaMap[area.ID()] = area
	}

	groupMap := make(map[int]*dto.FutureDutyTaskGroup)
	for _, task := range tasks {
		area := areaMap[task.AreaID()]
		if area == nil {
			continue
		}
		lastDuty, err := s.dutyRepo.FindLastByTaskDefID(ctx, task.ID())
		if err != nil {
			return nil, err
		}
		var lastCompletedAt *time.Time
		if lastDuty != nil {
			completedAt := lastDuty.End()
			lastCompletedAt = &completedAt
		}
		isDue := isTaskDueByLastCompletion(task, lastCompletedAt, referenceDate)
		includeInNextDuty := isDue
		overrideValue, hasOverride := overrideMap[task.ID()]
		if hasOverride {
			includeInNextDuty = overrideValue
		}

		taskGroup := groupMap[area.ID()]
		if taskGroup == nil {
			taskGroup = &dto.FutureDutyTaskGroup{
				AreaID:   area.ID(),
				AreaName: area.Name(),
				Floor:    area.Floor(),
				IsCommon: area.GroupID() == nil,
			}
			groupMap[area.ID()] = taskGroup
		}
		if includeInNextDuty {
			taskGroup.IncludeTasksCount++
		}
		taskGroup.Tasks = append(taskGroup.Tasks, dto.FutureDutyTaskItem{
			ID:                task.ID(),
			AreaID:            area.ID(),
			AreaName:          area.Name(),
			Title:             task.Title(),
			Frequency:         task.Frequency(),
			LastCompletedAt:   lastCompletedAt,
			IsDueByFrequency:  isDue,
			HasOverride:       hasOverride,
			IncludeInNextDuty: includeInNextDuty,
			IsCommon:          area.GroupID() == nil,
		})
	}
	groups := make([]dto.FutureDutyTaskGroup, 0, len(groupMap))
	for _, taskGroup := range groupMap {
		sort.Slice(taskGroup.Tasks, func(i, j int) bool {
			if taskGroup.Tasks[i].IncludeInNextDuty != taskGroup.Tasks[j].IncludeInNextDuty {
				return taskGroup.Tasks[i].IncludeInNextDuty
			}
			return taskGroup.Tasks[i].Title < taskGroup.Tasks[j].Title
		})
		groups = append(groups, *taskGroup)
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].IncludeTasksCount != groups[j].IncludeTasksCount {
			return groups[i].IncludeTasksCount > groups[j].IncludeTasksCount
		}
		return groups[i].AreaName < groups[j].AreaName
	})
	return groups, nil
}

func (s *Service) UpdateDormitoryDutySettings(ctx context.Context, dormitoryID int64, req dto.UpdateDormitoryDutySettingsRequest) error {
	dormitory, err := s.dormRepo.FindByID(ctx, dormitoryID)
	if err != nil {
		return err
	}
	if dormitory == nil {
		return fmt.Errorf("dormitory not found")
	}
	if err := s.UpdateCommonDutySettings(ctx, req.CommonTaskIDs); err != nil {
		return err
	}
	for _, groupSettings := range req.Groups {
		group, err := s.groupRepo.FindByID(ctx, groupSettings.GroupID)
		if err != nil {
			return err
		}
		if group == nil {
			return fmt.Errorf("group not found")
		}
		if group.DormitoryID() != dormitoryID {
			return fmt.Errorf("group %s does not belong to dormitory", group.ID())
		}
		if err := s.UpdateGroupDutySettings(ctx, groupSettings.GroupID, groupSettings.NextDutyTeam, groupSettings.IncludeTaskIDs); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) UpdateGroupDutySettings(ctx context.Context, groupID uuid.UUID, nextDutyTeam *int, includeTaskIDs []uuid.UUID) error {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("group not found")
	}
	if err := s.validateNextDutyTeam(ctx, groupID, nextDutyTeam); err != nil {
		return err
	}

	groupTasks, err := s.taskRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return err
	}
	allowed := make(map[uuid.UUID]struct{}, len(groupTasks))
	for _, task := range groupTasks {
		allowed[task.ID()] = struct{}{}
	}
	for _, taskID := range includeTaskIDs {
		if _, ok := allowed[taskID]; !ok {
			return fmt.Errorf("task %s does not belong to group", taskID)
		}
	}
	if err := s.saveTaskOverrides(ctx, groupTasks, includeTaskIDs, time.Now()); err != nil {
		return err
	}

	group.SetNextDutyTeam(nextDutyTeam)
	if err := s.groupRepo.Save(ctx, group); err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateCommonDutySettings(ctx context.Context, includeTaskIDs []uuid.UUID) error {
	commonTasks, err := s.taskRepo.FindCommon(ctx)
	if err != nil {
		return err
	}
	allowed := make(map[uuid.UUID]struct{}, len(commonTasks))
	for _, task := range commonTasks {
		allowed[task.ID()] = struct{}{}
	}
	for _, taskID := range includeTaskIDs {
		if _, ok := allowed[taskID]; !ok {
			return fmt.Errorf("task %s is not a common task", taskID)
		}
	}
	if err := s.saveTaskOverrides(ctx, commonTasks, includeTaskIDs, time.Now()); err != nil {
		return err
	}

	return nil
}

func (s *Service) saveTaskOverrides(ctx context.Context, tasks []*catalog.TaskDefinition, includeTaskIDs []uuid.UUID, referenceDate time.Time) error {
	includeSet := make(map[uuid.UUID]struct{}, len(includeTaskIDs))
	for _, taskID := range includeTaskIDs {
		includeSet[taskID] = struct{}{}
	}

	taskIDs := make([]uuid.UUID, 0, len(tasks))
	overrides := make([]*catalog.DutyTaskOverride, 0)
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID())
		lastDuty, err := s.dutyRepo.FindLastByTaskDefID(ctx, task.ID())
		if err != nil {
			return err
		}
		var lastCompletedAt *time.Time
		if lastDuty != nil {
			completedAt := lastDuty.End()
			lastCompletedAt = &completedAt
		}
		isDue := isTaskDueByLastCompletion(task, lastCompletedAt, referenceDate)
		_, include := includeSet[task.ID()]
		if include != isDue {
			overrides = append(overrides, catalog.NewDutyTaskOverride(task.ID(), include))
		}
	}
	return s.overrideRepo.ReplaceForTasks(ctx, taskIDs, overrides)
}

func isTaskDueByLastCompletion(task *catalog.TaskDefinition, lastCompletedAt *time.Time, referenceDate time.Time) bool {
	if task.Frequency() <= 1 {
		return true
	}
	if lastCompletedAt == nil {
		return true
	}
	daysPassed := int(referenceDate.Sub(*lastCompletedAt).Hours() / 24)
	return daysPassed >= task.Frequency()
}

func (s *Service) validateNextDutyTeam(ctx context.Context, groupID uuid.UUID, nextDutyTeam *int) error {
	if nextDutyTeam == nil {
		return nil
	}
	teams, err := s.teamRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return err
	}
	for _, team := range teams {
		if team.Order() == *nextDutyTeam {
			return nil
		}
	}
	return fmt.Errorf("next duty team not found")
}

func (s *Service) structureContext(ctx context.Context, teamID uuid.UUID) (*structure.Team, *structure.Group, *structure.Dormitory, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, nil, nil, err
	}
	if team == nil {
		return nil, nil, nil, nil
	}
	group, err := s.groupRepo.FindByID(ctx, team.GroupID())
	if err != nil {
		return nil, nil, nil, err
	}
	if group == nil {
		return team, nil, nil, nil
	}
	dormitory, err := s.dormRepo.FindByID(ctx, group.DormitoryID())
	if err != nil {
		return nil, nil, nil, err
	}
	return team, group, dormitory, nil
}

func (s *Service) detailLookups(ctx context.Context) (map[uuid.UUID]*catalog.TaskDefinition, map[int]*catalog.Area, map[uuid.UUID]*user.User, error) {
	taskDefs, err := s.taskRepo.GetAllTaskDefinitions(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	taskMap := make(map[uuid.UUID]*catalog.TaskDefinition, len(taskDefs))
	for _, task := range taskDefs {
		taskMap[task.ID()] = task
	}
	areaMap := make(map[int]*catalog.Area, len(areas))
	for _, area := range areas {
		areaMap[area.ID()] = area
	}
	userMap := make(map[uuid.UUID]*user.User, len(users))
	for _, resident := range users {
		userMap[resident.ID()] = resident
	}
	return taskMap, areaMap, userMap, nil
}

func fullName(u *user.User) string {
	parts := []string{u.LastName(), u.FirstName()}
	if u.MiddleName() != nil && *u.MiddleName() != "" {
		parts = append(parts, *u.MiddleName())
	}
	return strings.Join(parts, " ")
}

func taskStatus(task *dutydomain.DutyTask) string {
	if task.VerificationDate() != nil {
		return "Проверено"
	}
	if task.CompletionDate() != nil {
		return "Выполнено"
	}
	if task.AssigneeID() != nil {
		return "В работе"
	}
	return "Открыта"
}
