package cleaning

import (
	"context"
	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/duty"
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
			if task.CompletionDate() != nil {
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

func (s *Service) GetTeamTasks(ctx context.Context, teamID uuid.UUID) ([]dto.TaskViewModel, error) {
	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, teamID)
	if err != nil || d == nil {
		return nil, err
	}

	defs, _ := s.taskRepo.GetAllTaskDefinitions(ctx)
	areas, _ := s.areaRepo.GetAllAreas(ctx)
	users, _ := s.userRepo.FindByTeamID(ctx, teamID)
	team, _ := s.teamRepo.FindByID(ctx, teamID)

	defMap := make(map[uuid.UUID]*catalog.TaskDefinition)
	for _, df := range defs {
		defMap[df.ID()] = df
	}

	areaMap := make(map[int]*catalog.Area)
	for _, ar := range areas {
		areaMap[ar.ID()] = ar
	}

	userMap := make(map[uuid.UUID]*user.User)
	for _, u := range users {
		userMap[u.ID()] = u
	}

	var table []dto.TaskViewModel
	for _, dt := range d.Tasks() {
		def := defMap[dt.TaskDefID()]
		area := areaMap[def.AreaID()]

		var assignee *user.User
		isTeamLeader := false
		var stats *duty.UserStats
		if dt.AssigneeID() != nil {
			assignee = userMap[*dt.AssigneeID()]
			stats, err = s.GetUserStats(ctx, assignee.ID())
			if err != nil {
				return nil, err
			}
			isTeamLeader = team.LeaderID() != nil && *team.LeaderID() == assignee.ID()
		}

		table = append(table, dto.NewTaskViewModel(dt, def, area, dto.NewUserStats(assignee, stats, isTeamLeader)))
	}

	return table, nil
}

func (s *Service) GetLatestDuties(ctx context.Context) ([]dto.DutyViewModel, error) {
	latestDuties, err := s.dutyRepo.FindAllLatest(ctx)
	if err != nil || len(latestDuties) == 0 {
		return nil, err
	}

	var result []dto.DutyViewModel
	for _, d := range latestDuties {
		team, _ := s.teamRepo.FindByID(ctx, d.TeamID())
		group, _ := s.groupRepo.FindByID(ctx, team.GroupID())
		tasks, err := s.GetTeamTasks(ctx, d.TeamID())
		if err != nil {
			return nil, err
		}
		var usersStats []*dto.UserStats
		members, err := s.userRepo.FindByTeamID(ctx, d.TeamID())
		if err != nil {
			return nil, err
		}

		// TODO: этот кринж надо привести к получению одним запросом к бд
		for _, member := range members {
			stats, err := s.GetUserStats(ctx, member.ID())
			if err != nil {
				return nil, err
			}
			isTeamLeader := team.LeaderID() != nil && *team.LeaderID() == member.ID()
			usersStats = append(usersStats, dto.NewUserStats(member, stats, isTeamLeader))
		}

		vm := dto.NewDutyViewModel(d, team, group, usersStats, tasks)
		result = append(result, vm)
	}

	return result, nil
}

func (s *Service) GetTaskCandidates(ctx context.Context, userID uuid.UUID) ([]dto.TaskViewModel, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u.TeamID() == nil {
		return nil, fmt.Errorf("user or team not found")
	}
	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil || d == nil {
		return []dto.TaskViewModel{}, nil
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
	areaMap := make(map[int]*catalog.Area)
	for _, a := range areas {
		areaMap[a.ID()] = a
	}

	var result []dto.TaskViewModel
	for _, task := range d.Tasks() {
		isAssignedToMe := task.AssigneeID() != nil && *task.AssigneeID() == userID
		if task.AssigneeID() != nil && !isAssignedToMe {
			continue
		}
		if task.CompletionDate() != nil {
			continue
		}

		def, ok := defMap[task.TaskDefID()]
		if !ok {
			continue
		}

		area, ok := areaMap[def.AreaID]
		if !ok {
			continue
		}

		result = append(result, dto.TaskViewModel{
			ID:          task.ID(),
			Title:       def.Title,
			AreaID:      def.AreaID,
			AreaName:    area.Name(),
			AreaFloor:   area.Floor(),
			Cost:        def.Cost,
			IsCompleted: false,
		})
	}

	return result, nil
}

func (s *Service) GetUncompletedAssignedTasks(ctx context.Context, userID uuid.UUID) ([]dto.TaskViewModel, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u.TeamID() == nil {
		return nil, fmt.Errorf("user or team not found")
	}
	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil || d == nil {
		return []dto.TaskViewModel{}, nil
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

	areaMap := make(map[int]*catalog.Area)
	for _, a := range areas {
		areaMap[a.ID()] = a
	}

	var result []dto.TaskViewModel
	for _, task := range d.Tasks() {
		if task.AssigneeID() == nil || *task.AssigneeID() != userID || task.CompletionDate() != nil {
			continue
		}

		def, ok := defMap[task.TaskDefID()]
		if !ok {
			continue
		}

		area, ok := areaMap[def.AreaID()]
		if !ok {
			continue
		}

		result = append(result, dto.TaskViewModel{
			ID:          task.ID(),
			Title:       def.Title(),
			Cost:        def.Cost(),
			AreaID:      def.AreaID(),
			AreaName:    area.Name(),
			AreaFloor:   area.Floor(),
			IsCompleted: task.CompletionDate() != nil,
		})
	}
	return result, nil
}

func (s *Service) GetAllAssignedTasks(ctx context.Context, userID uuid.UUID) ([]dto.TaskViewModel, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u.TeamID() == nil {
		return nil, fmt.Errorf("user or team not found")
	}
	d, err := s.dutyRepo.FindCurrentByTeamID(ctx, *u.TeamID())
	if err != nil || d == nil {
		return []dto.TaskViewModel{}, nil
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

	areaMap := make(map[int]*catalog.Area)
	for _, a := range areas {
		areaMap[a.ID()] = a
	}

	var result []dto.TaskViewModel
	for _, task := range d.Tasks() {
		if task.AssigneeID() == nil || *task.AssigneeID() != userID {
			continue
		}

		def, ok := defMap[task.TaskDefID()]
		if !ok {
			continue
		}

		area, ok := areaMap[def.AreaID()]
		if !ok {
			continue
		}

		result = append(result, dto.TaskViewModel{
			ID:          task.ID(),
			Title:       def.Title(),
			Cost:        def.Cost(),
			AreaID:      def.AreaID(),
			AreaName:    area.Name(),
			AreaFloor:   area.Floor(),
			IsCompleted: task.CompletionDate() != nil,
		})
	}
	return result, nil
}

// TODO: Логика повторяется, нужно вынести в отдельный метод получение задач и фильтровать. отрефакторить и подумать о расположении query
