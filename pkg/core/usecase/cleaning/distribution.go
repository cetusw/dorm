package cleaning

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/domain/structure"
)

var (
	ErrGroupHasNoTeams          = errors.New("group has no teams")
	ErrLastDutyTeamNotFound     = errors.New("last duty team is not present in group rotation")
	ErrDutyCreationAccessDenied = errors.New("duty creation access denied")
	ErrDutyGroupNotFound        = errors.New("duty group not found")
	ErrDutyAlreadyCreated       = errors.New("duty already created")
)

type distributingContext struct {
	groups        []*structure.Group
	groupID       *uuid.UUID
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
	groupID     *uuid.UUID
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

func (s *Service) StartNewDutyForGroup(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, startDate, endDate time.Time) error {
	if !startDate.Before(endDate) {
		return fmt.Errorf("start date must be before end date")
	}

	if _, err := s.requireGroupLeaderAccess(ctx, currentUserID, groupID); err != nil {
		return err
	}

	distributingCtx, err := s.loadSchedulingData(ctx, schedulingOptions{
		groupID: &groupID,
		start:   startDate,
		end:     endDate,
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
	areas, err := s.loadSchedulingAreas(ctx, opts)
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
		groupID:       opts.groupID,
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
	if opts.groupID != nil {
		group, err := s.groupRepo.FindByID(ctx, *opts.groupID)
		if err != nil {
			return nil, err
		}
		if group == nil {
			return nil, nil
		}
		return []*structure.Group{group}, nil
	}
	if opts.dormitoryID != nil {
		return s.groupRepo.FindByDormitoryID(ctx, *opts.dormitoryID)
	}
	return s.groupRepo.FindAll(ctx)
}

func (s *Service) loadSchedulingAreas(ctx context.Context, opts schedulingOptions) ([]*catalog.Area, error) {
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, err
	}

	if opts.groupID == nil {
		return areas, nil
	}

	filtered := make([]*catalog.Area, 0, len(areas))
	for _, area := range areas {
		if area.GroupID() != nil && *area.GroupID() == *opts.groupID {
			filtered = append(filtered, area)
		}
	}

	return filtered, nil
}

func (s *Service) loadSchedulingTasks(ctx context.Context, groups []*structure.Group, areas []*catalog.Area, opts schedulingOptions) ([]*catalog.TaskDefinition, error) {
	if opts.groupID != nil {
		return s.taskRepo.FindByGroupID(ctx, *opts.groupID)
	}

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
		if c.groupID != nil {
			return ErrDutyGroupNotFound
		}
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

		nextSequence := 1
		if lastDuty != nil {
			nextSequence = lastDuty.SequenceNumber() + 1
		}

		newDuty := duty.NewDuty(nextTeam.ID(), c.start, c.end, nextSequence)
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
		isDue := def.IsScheduledFor(d.SequenceNumber())

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

func (s *Service) finalizeNewWeek(ctx context.Context, c *distributingContext) error {
	if len(c.dutiesList) == 0 {
		return fmt.Errorf("cannot finalize week: no duties to persist")
	}

	for _, d := range c.dutiesList {
		if err := s.dutyRepo.CreateWithTasks(ctx, d, d.Tasks()); err != nil {
			if errors.Is(err, duty.ErrDutyPeriodOverlap) {
				return duty.ErrDutyPeriodOverlap
			}
			if isDutySequenceConflict(err) {
				return ErrDutyAlreadyCreated
			}
			return fmt.Errorf("save duty %s: %w", d.ID(), err)
		}
	}
	if err := s.clearAppliedTaskOverrides(ctx, c.taskDefs); err != nil {
		return fmt.Errorf("clear task overrides: %w", err)
	}

	startedDuties := make([]events.StartedDuty, 0, len(c.dutiesList))
	for _, createdDuty := range c.dutiesList {
		startedDuties = append(startedDuties, events.StartedDuty{
			DutyID: createdDuty.ID(),
			TeamID: createdDuty.TeamID(),
		})
	}

	go func(dutiesSnapshot []events.StartedDuty) {
		pubCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := s.eventBus.Publish(pubCtx, events.TopicWeekStarted, events.WeekStartedEvent{
			Duties: dutiesSnapshot,
		}); err != nil {
			log.Printf("Failed to publish week started event: %v", err)
		}
	}(startedDuties)

	return nil
}

func (s *Service) clearAppliedTaskOverrides(ctx context.Context, tasks []*catalog.TaskDefinition) error {
	taskIDs := make([]uuid.UUID, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID())
	}
	return s.overrideRepo.DeleteByTaskIDs(ctx, taskIDs)
}

func (s *Service) requireGroupLeaderAccess(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID) (*structure.Group, error) {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("load group: %w", err)
	}
	if group == nil {
		return nil, ErrDutyGroupNotFound
	}
	if group.LeaderID() == nil || *group.LeaderID() != currentUserID {
		return nil, ErrDutyCreationAccessDenied
	}

	return group, nil
}

func isDutySequenceConflict(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) {
		return strings.Contains(strings.ToLower(err.Error()), "uq_duty_group_sequence")
	}

	return mysqlErr.Number == 1062 &&
		strings.Contains(strings.ToLower(mysqlErr.Message), "uq_duty_group_sequence")
}
