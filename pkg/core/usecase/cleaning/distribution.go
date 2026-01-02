package cleaning

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/domain/structure"
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

func (s *Service) StartNewWeek(ctx context.Context) error {
	distributingCtx, err := s.loadSchedulingData(ctx)
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

func (s *Service) loadSchedulingData(ctx context.Context) (*distributingContext, error) {
	groups, err := s.groupRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("load groups: %w", err)
	}
	taskDefs, err := s.catalogRepo.GetAllTaskDefinitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("load tasks: %w", err)
	}
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, fmt.Errorf("load areas: %w", err)
	}

	weekNum, _ := s.dutyRepo.CountDistinctStartDates(ctx)

	distributingCtx := &distributingContext{
		groups:        groups,
		taskDefs:      taskDefs,
		areas:         areas,
		weekNumber:    weekNum,
		start:         time.Now(),
		end:           time.Now().Add(7 * 24 * time.Hour),
		areaGroupMap:  make(map[int]*uuid.UUID),
		tasksByArea:   make(map[int][]*catalog.TaskDefinition),
		dutiesByGroup: make(map[uuid.UUID]*duty.Duty),
	}

	s.indexData(distributingCtx)
	return distributingCtx, nil
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
	for _, group := range c.groups {
		nextTeam, err := s.determineNextTeam(ctx, group.ID())
		if err != nil {
			fmt.Printf("Skipping group %s: %v\n", group.Name(), err)
			continue
		}

		newDuty := duty.NewDuty(nextTeam.ID(), c.start, c.end)
		c.dutiesByGroup[group.ID()] = newDuty
		c.dutiesList = append(c.dutiesList, newDuty)
	}

	sort.Slice(c.dutiesList, func(i, j int) bool {
		return c.dutiesList[i].TeamID().String() < c.dutiesList[j].TeamID().String()
	})

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

		if err := s.assignBatchToDuty(ctx, targetDuty, tasks); err != nil {
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

func (s *Service) assignBatchToDuty(ctx context.Context, d *duty.Duty, tasks []*catalog.TaskDefinition) error {
	for _, def := range tasks {
		isDue, err := s.isTaskDue(ctx, def)
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

func (s *Service) determineNextTeam(ctx context.Context, groupID uuid.UUID) (*structure.Team, error) {
	teams, err := s.teamRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, fmt.Errorf("no teams in group")
	}

	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Order() < teams[j].Order()
	})

	var lastDuty *duty.Duty

	for _, team := range teams {
		d, err := s.dutyRepo.FindCurrentByTeamID(ctx, team.ID())
		if err != nil {
			continue
		}
		if d != nil {
			if lastDuty == nil || d.Start().After(lastDuty.Start()) {
				lastDuty = d
			}
		}
	}

	if lastDuty == nil {
		return teams[0], nil
	}

	var lastTeamOrder int
	for _, t := range teams {
		if t.ID() == lastDuty.TeamID() {
			lastTeamOrder = t.Order()
			break
		}
	}

	totalTeams := len(teams)
	nextOrder := (lastTeamOrder % totalTeams) + 1

	for _, t := range teams {
		if t.Order() == nextOrder {
			return t, nil
		}
	}

	return nil, fmt.Errorf("logical error: could not find team with order %d", nextOrder)
}

func (s *Service) isTaskDue(ctx context.Context, def *catalog.TaskDefinition) (bool, error) {
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

	daysPassed := int(time.Since(lastDuty.Start()).Hours() / 24)

	return daysPassed >= def.Frequency(), nil
}

func (s *Service) finalizeNewWeek(ctx context.Context, c *distributingContext) error {
	for _, d := range c.dutiesList {
		if err := s.dutyRepo.Save(ctx, d); err != nil {
			return fmt.Errorf("save duty %s: %w", d.ID(), err)
		}
	}

	go func() {
		bgCtx := context.WithoutCancel(ctx)
		_ = s.eventBus.Publish(bgCtx, events.TopicWeekStarted, events.WeekStartedEvent{
			StartDate: c.start,
		})
	}()

	return nil
}
