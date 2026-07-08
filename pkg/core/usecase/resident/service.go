package resident

import (
	"context"
	"fmt"
	"sort"
	"time"

	"dorm/pkg/core/domain/catalog"
	dutydomain "dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type Service struct {
	userRepo   user.Repository
	teamRepo   structure.TeamRepository
	groupRepo  structure.GroupRepository
	dutyRepo   dutydomain.Repository
	taskRepo   catalog.TaskDefinitionRepository
	areaRepo   catalog.AreaRepository
	cleaningUC ports.CleaningUseCase
	now        func() time.Time
}

func NewResidentDutyService(
	userRepo user.Repository,
	teamRepo structure.TeamRepository,
	groupRepo structure.GroupRepository,
	dutyRepo dutydomain.Repository,
	taskRepo catalog.TaskDefinitionRepository,
	areaRepo catalog.AreaRepository,
	cleaningUC ports.CleaningUseCase,
) *Service {
	return &Service{
		userRepo:   userRepo,
		teamRepo:   teamRepo,
		groupRepo:  groupRepo,
		dutyRepo:   dutyRepo,
		taskRepo:   taskRepo,
		areaRepo:   areaRepo,
		cleaningUC: cleaningUC,
		now:        time.Now,
	}
}

func (s *Service) GetCurrentDuty(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.ResidentCurrentDutyResponse, error) {
	resident, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load resident: %w", err)
	}
	if resident == nil {
		return nil, fmt.Errorf("resident not found")
	}
	if resident.TeamID() == nil {
		return nil, fmt.Errorf("resident is not assigned to a team")
	}

	team, err := s.teamRepo.FindByID(ctx, *resident.TeamID())
	if err != nil {
		return nil, fmt.Errorf("load resident team: %w", err)
	}
	if team == nil {
		return nil, fmt.Errorf("team not found")
	}

	group, err := s.groupRepo.FindByID(ctx, team.GroupID())
	if err != nil {
		return nil, fmt.Errorf("load resident group: %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("group not found")
	}

	currentDuty, err := s.dutyRepo.FindActiveByTeamID(ctx, team.ID(), s.now())
	if err != nil {
		return nil, fmt.Errorf("load active duty: %w", err)
	}
	if currentDuty == nil {
		return nil, nil
	}

	taskDefinitions, err := s.taskDefinitionsByID(ctx)
	if err != nil {
		return nil, err
	}

	areas, err := s.areasByID(ctx)
	if err != nil {
		return nil, err
	}

	users, err := s.userNamesByID(ctx, team.ID())
	if err != nil {
		return nil, err
	}

	tasks := make([]dto.ResidentDutyTask, 0, len(currentDuty.Tasks()))

	for _, dutyTask := range currentDuty.Tasks() {
		taskDefinition, ok := taskDefinitions[dutyTask.TaskDefID()]
		if !ok {
			continue
		}

		area, ok := areas[taskDefinition.AreaID()]
		if !ok {
			continue
		}

		status := resolveTaskStatus(dutyTask)
		isMine := dutyTask.AssigneeID() != nil && *dutyTask.AssigneeID() == userID
		canTake, canReturn, canComplete := buildTaskPermissions(status, isMine)

		var assigneeID *string
		var assigneeName *string

		if dutyTask.AssigneeID() != nil {
			rawAssigneeID := dutyTask.AssigneeID().String()
			assigneeID = &rawAssigneeID

			if name, ok := users[*dutyTask.AssigneeID()]; ok {
				assigneeName = &name
			}
		}

		tasks = append(tasks, dto.ResidentDutyTask{
			ID:           dutyTask.ID().String(),
			AreaName:     area.Name(),
			AreaFloor:    area.Floor(),
			Title:        taskDefinition.Title(),
			Cost:         taskDefinition.Cost(),
			Status:       status,
			AssigneeID:   assigneeID,
			AssigneeName: assigneeName,
			IsMine:       isMine,
			CanTake:      canTake,
			CanReturn:    canReturn,
			CanComplete:  canComplete,
		})
	}

	sortResidentDutyTasks(tasks)

	return &dto.ResidentCurrentDutyResponse{
		DutyID:    currentDuty.ID().String(),
		Group:     group.Name(),
		Team:      team.Name(),
		StartDate: currentDuty.Start().Format("2006-01-02"),
		EndDate:   currentDuty.End().Format("2006-01-02"),
		Tasks:     tasks,
	}, nil
}

func (s *Service) TakeTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	return s.cleaningUC.AssignTask(ctx, taskID, userID)
}

func (s *Service) ReturnTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	return s.cleaningUC.UnassignTask(ctx, taskID, userID)
}

func (s *Service) CompleteTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	return s.cleaningUC.CompleteTask(ctx, taskID, userID)
}

func (s *Service) taskDefinitionsByID(
	ctx context.Context,
) (map[uuid.UUID]*catalog.TaskDefinition, error) {
	taskDefinitions, err := s.taskRepo.GetAllTaskDefinitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("load task definitions: %w", err)
	}

	result := make(map[uuid.UUID]*catalog.TaskDefinition, len(taskDefinitions))
	for _, taskDefinition := range taskDefinitions {
		result[taskDefinition.ID()] = taskDefinition
	}

	return result, nil
}

func (s *Service) areasByID(ctx context.Context) (map[int]*catalog.Area, error) {
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, fmt.Errorf("load areas: %w", err)
	}

	result := make(map[int]*catalog.Area, len(areas))
	for _, area := range areas {
		result[area.ID()] = area
	}

	return result, nil
}

func (s *Service) userNamesByID(
	ctx context.Context,
	teamID uuid.UUID,
) (map[uuid.UUID]string, error) {
	users, err := s.userRepo.FindByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("load team users: %w", err)
	}

	result := make(map[uuid.UUID]string, len(users))
	for _, resident := range users {
		result[resident.ID()] = formatUserName(resident)
	}

	return result, nil
}

func resolveTaskStatus(task *dutydomain.DutyTask) string {
	if task.VerificationDate() != nil {
		return dto.ResidentDutyTaskStatusVerified
	}
	if task.CompletionDate() != nil {
		return dto.ResidentDutyTaskStatusCompleted
	}
	if task.AssigneeID() != nil {
		return dto.ResidentDutyTaskStatusAssigned
	}
	return dto.ResidentDutyTaskStatusFree
}

func buildTaskPermissions(status string, isMine bool) (canTake bool, canReturn bool, canComplete bool) {
	switch status {
	case dto.ResidentDutyTaskStatusFree:
		return true, false, false
	case dto.ResidentDutyTaskStatusAssigned:
		if isMine {
			return false, true, true
		}
		return false, false, false
	default:
		return false, false, false
	}
}

func sortResidentDutyTasks(tasks []dto.ResidentDutyTask) {
	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].AreaFloor != tasks[j].AreaFloor {
			return tasks[i].AreaFloor > tasks[j].AreaFloor
		}
		if tasks[i].AreaName != tasks[j].AreaName {
			return tasks[i].AreaName < tasks[j].AreaName
		}
		if tasks[i].Cost != tasks[j].Cost {
			return tasks[i].Cost > tasks[j].Cost
		}
		return tasks[i].Title < tasks[j].Title
	})
}

func formatUserName(resident *user.User) string {
	return fmt.Sprintf("%s %s", resident.FirstName(), resident.LastName())
}
