package cleaning

import (
	"context"
	"dorm/pkg/core/domain/catalog"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/ports"
)

func (s *Service) IsUserOnDuty(ctx context.Context, userID uuid.UUID) (bool, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u.TeamID() == nil {
		return false, err
	}

	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil {
		return false, err
	}

	if d == nil {
		return false, nil
	}

	now := time.Now()
	return (now.After(d.Start()) || now.Equal(d.Start())) &&
		(now.Before(d.End()) || now.Equal(d.End())), nil
}

func (s *Service) GetUserStats(ctx context.Context, userID uuid.UUID) (*duty.UserStats, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u.TeamID() == nil {
		return &duty.UserStats{}, nil
	}

	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil || d == nil {
		return &duty.UserStats{}, nil
	}

	members, err := s.userRepo.FindByTeamID(ctx, *u.TeamID())
	if err != nil {
		return nil, fmt.Errorf("stats: failed to get team members: %w", err)
	}

	defs, err := s.taskRepo.GetAllTaskDefinitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("stats: failed to get task defs: %w", err)
	}

	costMap := make(map[uuid.UUID]int)
	for _, def := range defs {
		costMap[def.ID()] = def.Cost()
	}

	var totalDutyCost, userAssigned, userConfirmed int

	for _, task := range d.Tasks() {
		cost, ok := costMap[task.TaskDefID()]
		if !ok {
			continue
		}

		totalDutyCost += cost

		if task.AssigneeID() != nil && *task.AssigneeID() == userID {
			userAssigned += cost
			if task.IsCompleted() {
				userConfirmed += cost
			}
		}
	}

	var required float64
	if len(members) > 0 {
		required = float64(totalDutyCost) / float64(len(members))
	}

	return &duty.UserStats{
		TotalPoints:     userAssigned,
		ConfirmedPoints: userConfirmed,
		RequiredPoints:  required,
	}, nil
}

func (s *Service) GetTaskCandidates(ctx context.Context, userID uuid.UUID) ([]ports.TaskViewModel, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u.TeamID() == nil {
		return nil, fmt.Errorf("user or team not found")
	}
	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil || d == nil {
		return []ports.TaskViewModel{}, nil
	}

	taskDefs, err := s.taskRepo.GetAllTaskDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, err
	}

	defMap := make(map[uuid.UUID]struct {
		Title  string
		AreaID int
		Cost   int
	})
	for _, t := range taskDefs {
		defMap[t.ID()] = struct {
			Title  string
			AreaID int
			Cost   int
		}{t.Title(), t.AreaID(), t.Cost()}
	}
	areaMap := make(map[int]string)
	for _, a := range areas {
		areaMap[a.ID()] = a.Name()
	}

	var result []ports.TaskViewModel
	for _, task := range d.Tasks() {
		if task.AssigneeID() != nil {
			continue
		}

		def, ok := defMap[task.TaskDefID()]
		if !ok {
			continue
		}

		result = append(result, ports.TaskViewModel{
			ID:       task.ID(),
			Title:    def.Title,
			AreaID:   def.AreaID,
			AreaName: areaMap[def.AreaID],
			Cost:     def.Cost,
			IsDone:   false,
		})
	}

	return result, nil
}

func (s *Service) GetAssignedTasks(ctx context.Context, userID uuid.UUID) ([]ports.TaskViewModel, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u.TeamID() == nil {
		return nil, fmt.Errorf("user or team not found")
	}
	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil || d == nil {
		return []ports.TaskViewModel{}, nil
	}

	taskDefs, err := s.taskRepo.GetAllTaskDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, err
	}

	defMap := make(map[uuid.UUID]*catalog.TaskDefinition)
	for _, t := range taskDefs {
		defMap[t.ID()] = t
	}

	areaMap := make(map[int]string)
	for _, a := range areas {
		areaMap[a.ID()] = a.Name()
	}

	var result []ports.TaskViewModel
	for _, task := range d.Tasks() {
		if task.AssigneeID() == nil || *task.AssigneeID() != userID || task.IsCompleted() {
			continue
		}

		def, ok := defMap[task.TaskDefID()]
		if !ok {
			continue
		}

		result = append(result, ports.TaskViewModel{
			ID:       task.ID(),
			Title:    def.Title(),
			Cost:     def.Cost(),
			AreaID:   def.AreaID(),
			AreaName: areaMap[def.AreaID()],
			IsDone:   task.IsCompleted(),
		})
	}
	return result, nil
}

// TODO: отрефакторить и подумать о расположении query
