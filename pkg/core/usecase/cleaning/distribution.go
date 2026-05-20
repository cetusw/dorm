package cleaning

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/ports/dto"
)

type distributingContext struct {
	groups        []*structure.Group
	taskDefs      []*catalog.TaskDefinition
	areas         []*catalog.Area
	areaGroupMap  map[int]*uuid.UUID
	tasksByArea   map[int][]*catalog.TaskDefinition
	dutiesByGroup map[uuid.UUID]*duty.Duty
	dutiesList    []*duty.Duty
	weekNumber    int
	start         time.Time
	end           time.Time
}

type schedulingOptions struct {
	dormitoryID        *int64
	start              time.Time
	end                time.Time
	commonTaskIDs      map[uuid.UUID]struct{}
	commonTaskOverride bool
}

func (s *Service) StartNewWeek(ctx context.Context) error {
	now := time.Now()
	distributingCtx, err := s.loadSchedulingData(ctx, schedulingOptions{
		start: now,
		end:   now.Add(7 * 24 * time.Hour),
	})
	if err != nil {
		return err
	}

	if err := s.initializeDuties(ctx, distributingCtx); err != nil {
		return err
	}

	if err := s.distributeTasks(ctx, distributingCtx); err != nil {
		return err
	}

	return s.finalizeNewWeek(ctx, distributingCtx)
}

func (s *Service) StartNewDutiesForDormitory(ctx context.Context, dormitoryID int64, startDate, endDate time.Time, commonTaskIDs []uuid.UUID) error {
	if !startDate.Before(endDate) {
		return fmt.Errorf("start date must be before end date")
	}

	selectedCommonTasks := make(map[uuid.UUID]struct{}, len(commonTaskIDs))
	for _, taskID := range commonTaskIDs {
		selectedCommonTasks[taskID] = struct{}{}
	}

	distributingCtx, err := s.loadSchedulingData(ctx, schedulingOptions{
		dormitoryID:        &dormitoryID,
		start:              startDate,
		end:                endDate,
		commonTaskIDs:      selectedCommonTasks,
		commonTaskOverride: true,
	})
	if err != nil {
		return err
	}

	if err := s.initializeDuties(ctx, distributingCtx); err != nil {
		return err
	}

	if err := s.distributeTasks(ctx, distributingCtx); err != nil {
		return err
	}

	return s.finalizeNewWeek(ctx, distributingCtx)
}

func (s *Service) loadSchedulingData(ctx context.Context, opts schedulingOptions) (*distributingContext, error) {
	groups, err := s.loadSchedulingGroups(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("load groups: %w", err)
	}
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, fmt.Errorf("load areas: %w", err)
	}
	taskDefs, err := s.loadSchedulingTasks(ctx, groups, areas, opts)
	if err != nil {
		return nil, fmt.Errorf("load tasks: %w", err)
	}

	weekNum, err := s.dutyRepo.CountDistinctStartDates(ctx)
	if err != nil {
		return nil, fmt.Errorf("load week number: %w", err)
	}

	distributingCtx := &distributingContext{
		groups:        groups,
		taskDefs:      taskDefs,
		areas:         areas,
		weekNumber:    weekNum,
		start:         opts.start,
		end:           opts.end,
		areaGroupMap:  make(map[int]*uuid.UUID),
		tasksByArea:   make(map[int][]*catalog.TaskDefinition),
		dutiesByGroup: make(map[uuid.UUID]*duty.Duty),
	}

	s.indexData(distributingCtx)
	return distributingCtx, nil
}

func (s *Service) loadSchedulingGroups(ctx context.Context, opts schedulingOptions) ([]*structure.Group, error) {
	if opts.dormitoryID != nil {
		return s.groupRepo.FindByDormitoryID(ctx, *opts.dormitoryID)
	}
	return s.groupRepo.FindAll(ctx)
}

func (s *Service) loadSchedulingTasks(ctx context.Context, groups []*structure.Group, areas []*catalog.Area, opts schedulingOptions) ([]*catalog.TaskDefinition, error) {
	if !opts.commonTaskOverride {
		return s.taskRepo.GetActiveTaskDefinitions(ctx)
	}

	allTasks, err := s.taskRepo.GetAllTaskDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	groupSet := make(map[uuid.UUID]struct{}, len(groups))
	for _, group := range groups {
		groupSet[group.ID()] = struct{}{}
	}
	areaMap := make(map[int]*catalog.Area, len(areas))
	for _, area := range areas {
		areaMap[area.ID()] = area
	}

	taskDefs := make([]*catalog.TaskDefinition, 0, len(allTasks))
	for _, task := range allTasks {
		area := areaMap[task.AreaID()]
		if area == nil {
			continue
		}
		if area.GroupID() == nil {
			if _, ok := opts.commonTaskIDs[task.ID()]; ok {
				taskDefs = append(taskDefs, task)
			}
			continue
		}
		if _, ok := groupSet[*area.GroupID()]; ok && task.IsActive() {
			taskDefs = append(taskDefs, task)
		}
	}
	return taskDefs, nil
}

func (s *Service) indexData(c *distributingContext) {
	for _, area := range c.areas {
		c.areaGroupMap[area.ID()] = area.GroupID()
	}
	for _, def := range c.taskDefs {
		c.tasksByArea[def.AreaID()] = append(c.tasksByArea[def.AreaID()], def)
	}
	sort.Slice(c.areas, func(i, j int) bool { return c.areas[i].ID() < c.areas[j].ID() })
}

func (s *Service) initializeDuties(ctx context.Context, c *distributingContext) error {
	if len(c.groups) == 0 {
		return fmt.Errorf("cannot start new week: no groups found")
	}

	skippedGroups := 0
	for _, group := range c.groups {
		nextTeam, err := s.determineNextTeam(ctx, group)
		if err != nil {
			log.Printf("Skipping group %s: %v", group.Name(), err)
			skippedGroups++
			continue
		}

		newDuty := duty.NewDuty(nextTeam.ID(), c.start, c.end)
		c.dutiesByGroup[group.ID()] = newDuty
		c.dutiesList = append(c.dutiesList, newDuty)
	}

	sort.Slice(c.dutiesList, func(i, j int) bool {
		return c.dutiesList[i].TeamID().String() < c.dutiesList[j].TeamID().String()
	})

	if len(c.dutiesList) == 0 {
		return fmt.Errorf("cannot start new week: no duties generated (%d groups skipped)", skippedGroups)
	}

	return nil
}

func (s *Service) distributeTasks(ctx context.Context, c *distributingContext) error {
	if len(c.dutiesList) == 0 {
		return nil
	}

	rrIndex := c.weekNumber % len(c.dutiesList)

	for _, area := range c.areas {
		tasks := c.tasksByArea[area.ID()]
		if len(tasks) == 0 {
			continue
		}

		targetDuty := s.resolveTargetDuty(area, c, rrIndex)

		if c.areaGroupMap[area.ID()] == nil {
			rrIndex = (rrIndex + 1) % len(c.dutiesList)
		}

		if targetDuty == nil {
			continue
		}

		if err := s.assignBatchToDuty(ctx, targetDuty, tasks, c.start); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) resolveTargetDuty(area *catalog.Area, c *distributingContext, rrIndex int) *duty.Duty {
	groupID := c.areaGroupMap[area.ID()]

	if groupID != nil {
		if d, ok := c.dutiesByGroup[*groupID]; ok {
			return d
		}
		return nil
	}

	if len(c.dutiesList) > 0 {
		return c.dutiesList[rrIndex]
	}

	return nil
}

func (s *Service) assignBatchToDuty(ctx context.Context, d *duty.Duty, tasks []*catalog.TaskDefinition, referenceDate time.Time) error {
	for _, def := range tasks {
		isDue, err := s.isTaskDue(ctx, def, referenceDate)
		if err != nil {
			fmt.Printf("Frequency check failed for %s: %v\n", def.Title(), err)
			isDue = true
		}

		if isDue {
			d.AddTask(uuid.New(), def.ID())
		}
	}
	return nil
}

func (s *Service) determineNextTeam(ctx context.Context, group *structure.Group) (*structure.Team, error) {
	teams, err := s.teamRepo.FindByGroupID(ctx, group.ID())
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, fmt.Errorf("no teams in group")
	}

	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Order() < teams[j].Order()
	})

	dutyTeamIndex := 0
	if group.NextDutyTeam() != nil {
		for index, team := range teams {
			if team.Order() == *group.NextDutyTeam() {
				dutyTeamIndex = index
				break
			}
		}
	}

	nextIndex := (dutyTeamIndex + 1) % len(teams)
	nextOrder := teams[nextIndex].Order()
	group.SetNextDutyTeam(&nextOrder)
	if err := s.groupRepo.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("update next duty team: %w", err)
	}
	return teams[dutyTeamIndex], nil
}

func (s *Service) isTaskDue(ctx context.Context, def *catalog.TaskDefinition, referenceDate time.Time) (bool, error) {
	if def.Frequency() <= 1 {
		return true, nil
	}

	lastDuty, err := s.dutyRepo.FindLastByTaskDefID(ctx, def.ID())
	if err != nil {
		return false, err
	}
	if lastDuty == nil {
		return true, nil
	}

	daysPassed := int(referenceDate.Sub(lastDuty.Start()).Hours() / 24)

	return daysPassed >= def.Frequency(), nil
}

func (s *Service) finalizeNewWeek(ctx context.Context, c *distributingContext) error {
	if len(c.dutiesList) == 0 {
		return fmt.Errorf("cannot finalize week: no duties to persist")
	}

	for _, d := range c.dutiesList {
		if err := s.dutyRepo.Save(ctx, d); err != nil {
			return fmt.Errorf("save duty %s: %w", d.ID(), err)
		}
	}

	duties, err := s.GetLatestDuties(ctx)
	if err != nil {
		log.Printf("Failed to load duties for week started event: %v", err)
		return nil
	}

	go func(dutiesSnapshot []dto.DutyViewModel, start, end time.Time) {
		pubCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := s.eventBus.Publish(pubCtx, events.TopicWeekStarted, events.WeekStartedEvent{
			StartDate: start,
			EndDate:   end,
			Duties:    dutiesSnapshot,
		}); err != nil {
			log.Printf("Failed to publish week started event: %v", err)
		}
	}(duties, c.start, c.end)

	return nil
}
