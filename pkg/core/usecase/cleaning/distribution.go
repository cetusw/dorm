package cleaning

import (
	"context"
	"errors"
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

var (
	ErrGroupHasNoTeams      = errors.New("group has no teams")
	ErrLastDutyTeamNotFound = errors.New("last duty team is not present in group rotation")
)

type distributingContext struct {
	groups        []*structure.Group
	taskDefs      []*catalog.TaskDefinition
	areas         []*catalog.Area
	areaGroupMap  map[int]*uuid.UUID
	tasksByArea   map[int][]*catalog.TaskDefinition
	taskOverrides map[uuid.UUID]bool
	dutiesByGroup map[uuid.UUID]*duty.Duty
	dutiesList    []*duty.Duty
	weekNumber    int
	start         time.Time
	end           time.Time
}

type schedulingOptions struct {
	dormitoryID *int64
	start       time.Time
	end         time.Time
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

func (s *Service) StartNewDutiesForDormitory(ctx context.Context, dormitoryID int64, startDate, endDate time.Time) error {
	if !startDate.Before(endDate) {
		return fmt.Errorf("start date must be before end date")
	}

	distributingCtx, err := s.loadSchedulingData(ctx, schedulingOptions{
		dormitoryID: &dormitoryID,
		start:       startDate,
		end:         endDate,
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
	overrides, err := s.overrideRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("load task overrides: %w", err)
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
		taskOverrides: make(map[uuid.UUID]bool, len(overrides)),
		dutiesByGroup: make(map[uuid.UUID]*duty.Duty),
	}
	for _, override := range overrides {
		distributingCtx.taskOverrides[override.TaskID()] = override.IncludeInNextDuty()
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
	if opts.dormitoryID == nil {
		return s.taskRepo.GetAllTaskDefinitions(ctx)
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
			taskDefs = append(taskDefs, task)
			continue
		}
		if _, ok := groupSet[*area.GroupID()]; ok {
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
		teams, err := s.teamRepo.FindByGroupID(ctx, group.ID())
		if err != nil {
			log.Printf("Skipping group %s: load teams: %v", group.Name(), err)
			skippedGroups++
			continue
		}

		lastDuty, err := s.dutyRepo.FindLatestByGroupID(ctx, group.ID())
		if err != nil {
			log.Printf("Skipping group %s: load last duty: %v", group.Name(), err)
			skippedGroups++
			continue
		}

		nextTeam, err := determineNextTeam(teams, lastDuty)
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

		if err := s.assignBatchToDuty(ctx, targetDuty, tasks, c); err != nil {
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

func (s *Service) assignBatchToDuty(ctx context.Context, d *duty.Duty, tasks []*catalog.TaskDefinition, c *distributingContext) error {
	for _, def := range tasks {
		isDue, err := s.isTaskDue(ctx, def, c.start)
		if err != nil {
			fmt.Printf("Frequency check failed for %s: %v\n", def.Title(), err)
			isDue = true
		}

		if include, ok := c.taskOverrides[def.ID()]; ok {
			isDue = include
		}

		if isDue {
			d.AddTask(uuid.New(), def.ID())
		}
	}
	return nil
}

func determineNextTeam(teams []*structure.Team, lastDuty *duty.Duty) (*structure.Team, error) {
	if len(teams) == 0 {
		return nil, ErrGroupHasNoTeams
	}

	sort.Slice(teams, func(i, j int) bool {
		return teams[i].RotationPosition() < teams[j].RotationPosition()
	})

	if lastDuty == nil {
		return teams[0], nil
	}

	for index, team := range teams {
		if team.ID() == lastDuty.TeamID() {
			nextIndex := (index + 1) % len(teams)
			return teams[nextIndex], nil
		}
	}

	return nil, ErrLastDutyTeamNotFound
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

	daysPassed := int(referenceDate.Sub(lastDuty.End()).Hours() / 24)

	return daysPassed >= def.Frequency(), nil
}

func (s *Service) finalizeNewWeek(ctx context.Context, c *distributingContext) error {
	if len(c.dutiesList) == 0 {
		return fmt.Errorf("cannot finalize week: no duties to persist")
	}

	for _, d := range c.dutiesList {
		if err := s.dutyRepo.CreateWithTasks(ctx, d, d.Tasks()); err != nil {
			return fmt.Errorf("save duty %s: %w", d.ID(), err)
		}
	}
	if err := s.clearAppliedTaskOverrides(ctx, c.taskDefs); err != nil {
		return fmt.Errorf("clear task overrides: %w", err)
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

func (s *Service) clearAppliedTaskOverrides(ctx context.Context, tasks []*catalog.TaskDefinition) error {
	taskIDs := make([]uuid.UUID, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID())
	}
	return s.overrideRepo.DeleteByTaskIDs(ctx, taskIDs)
}
