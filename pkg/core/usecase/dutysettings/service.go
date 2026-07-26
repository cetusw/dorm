package dutysettings

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

var ErrAccessDenied = errors.New("duty settings access denied")

type Service struct {
	groupRepo structure.GroupRepository
	areaRepo  catalog.AreaRepository
	taskRepo  catalog.TaskDefinitionRepository
}

func NewDutySettingsService(
	groupRepo structure.GroupRepository,
	areaRepo catalog.AreaRepository,
	taskRepo catalog.TaskDefinitionRepository,
) *Service {
	return &Service{
		groupRepo: groupRepo,
		areaRepo:  areaRepo,
		taskRepo:  taskRepo,
	}
}

func (s *Service) GetDutySettings(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID) (*dto.DutySettingsResponse, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return nil, err
	}

	areas, err := s.listGroupAreas(ctx, group.ID())
	if err != nil {
		return nil, err
	}

	tasks, err := s.taskRepo.FindByGroupID(ctx, group.ID())
	if err != nil {
		return nil, fmt.Errorf("load group tasks: %w", err)
	}

	taskIDs := make([]uuid.UUID, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID())
	}

	lastCompletionDates, err := s.taskRepo.FindLastCompletionDates(ctx, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("load task completion dates: %w", err)
	}

	tasksByArea := make(map[int][]dto.DutySettingsTask, len(areas))
	for _, task := range tasks {
		tasksByArea[task.AreaID()] = append(tasksByArea[task.AreaID()], dto.DutySettingsTask{
			ID:              task.ID().String(),
			Title:           task.Title(),
			Cost:            task.Cost(),
			Frequency:       task.Frequency(),
			LastCompletedAt: lastCompletionDates[task.ID()],
		})
	}

	responseAreas := make([]dto.DutySettingsArea, 0, len(areas))
	for _, area := range areas {
		areaTasks := tasksByArea[area.ID()]
		sort.Slice(areaTasks, func(i, j int) bool {
			if areaTasks[i].Cost != areaTasks[j].Cost {
				return areaTasks[i].Cost > areaTasks[j].Cost
			}
			return areaTasks[i].Title < areaTasks[j].Title
		})

		responseAreas = append(responseAreas, dto.DutySettingsArea{
			ID:    area.ID(),
			Name:  area.Name(),
			Floor: areaFloor(area),
			Tasks: areaTasks,
		})
	}

	sort.Slice(responseAreas, func(i, j int) bool {
		leftFloor := responseAreas[i].Floor
		rightFloor := responseAreas[j].Floor
		switch {
		case leftFloor == nil && rightFloor != nil:
			return false
		case leftFloor != nil && rightFloor == nil:
			return true
		case leftFloor != nil && rightFloor != nil && *leftFloor != *rightFloor:
			return *leftFloor > *rightFloor
		default:
			return responseAreas[i].Name < responseAreas[j].Name
		}
	})

	return &dto.DutySettingsResponse{
		Group: dto.DutySettingsGroup{
			ID:          group.ID().String(),
			Name:        group.Name(),
			DormitoryID: group.DormitoryID(),
		},
		Areas: responseAreas,
	}, nil
}

func (s *Service) CreateArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, req dto.CreateAreaRequest) (*dto.AreaDetails, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeAreaInput(req.Name, req.Floor)
	if err != nil {
		return nil, err
	}

	area := catalog.NewArea(normalized.name, normalized.floorValue(), &groupID)
	if err := s.areaRepo.Save(ctx, area); err != nil {
		return nil, fmt.Errorf("create area: %w", err)
	}

	return buildAreaDetails(area, group), nil
}

func (s *Service) UpdateArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int, req dto.UpdateAreaRequest) (*dto.AreaDetails, error) {
	group, area, err := s.requireManagedGroupArea(ctx, currentUserID, groupID, areaID)
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeAreaInput(req.Name, req.Floor)
	if err != nil {
		return nil, err
	}

	updated := catalog.RestoreArea(area.ID(), normalized.name, normalized.floorValue(), area.GroupID())
	if err := s.areaRepo.Save(ctx, updated); err != nil {
		return nil, fmt.Errorf("update area: %w", err)
	}

	return buildAreaDetails(updated, group), nil
}

func (s *Service) DeleteArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int) error {
	_, _, err := s.requireManagedGroupArea(ctx, currentUserID, groupID, areaID)
	if err != nil {
		return err
	}

	if err := s.areaRepo.Delete(ctx, areaID); err != nil {
		return fmt.Errorf("delete area: %w", err)
	}

	return nil
}

func (s *Service) CreateTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int, req dto.CreateTaskRequest) (*dto.TaskDetails, error) {
	group, area, err := s.requireManagedGroupArea(ctx, currentUserID, groupID, areaID)
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeTaskInput(req.Title, req.Cost, req.Frequency)
	if err != nil {
		return nil, err
	}

	task, err := catalog.NewTaskDefinition(area.ID(), normalized.title, normalized.cost, normalized.frequency)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return buildTaskDetails(task, group, area), nil
}

func (s *Service) UpdateTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, taskID uuid.UUID, req dto.UpdateTaskRequest) (*dto.TaskDetails, error) {
	group, task, _, err := s.requireManagedGroupTask(ctx, currentUserID, groupID, taskID)
	if err != nil {
		return nil, err
	}

	targetArea, err := s.areaRepo.FindByID(ctx, req.AreaID)
	if err != nil {
		return nil, fmt.Errorf("load target area: %w", err)
	}
	if targetArea == nil {
		return nil, fmt.Errorf("территория не найдена")
	}
	if err := requireAreaInGroup(targetArea, group.ID()); err != nil {
		return nil, err
	}

	normalized, err := normalizeTaskInput(req.Title, req.Cost, req.Frequency)
	if err != nil {
		return nil, err
	}

	updated := catalog.RestoreTaskDefinition(task.ID(), targetArea.ID(), normalized.title, normalized.cost, normalized.frequency)
	if err := s.taskRepo.Save(ctx, updated); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	return buildTaskDetails(updated, group, targetArea), nil
}

func (s *Service) DeleteTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, taskID uuid.UUID) error {
	_, _, _, err := s.requireManagedGroupTask(ctx, currentUserID, groupID, taskID)
	if err != nil {
		return err
	}

	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (s *Service) requireManagedGroup(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID) (*structure.Group, error) {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("load group: %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("группа не найдена")
	}
	if group.LeaderID() == nil || *group.LeaderID() != currentUserID {
		return nil, ErrAccessDenied
	}

	return group, nil
}

func (s *Service) requireManagedGroupArea(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, areaID int) (*structure.Group, *catalog.Area, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return nil, nil, err
	}

	area, err := s.areaRepo.FindByID(ctx, areaID)
	if err != nil {
		return nil, nil, fmt.Errorf("load area: %w", err)
	}
	if area == nil {
		return nil, nil, fmt.Errorf("территория не найдена")
	}
	if err := requireAreaInGroup(area, group.ID()); err != nil {
		return nil, nil, err
	}

	return group, area, nil
}

func (s *Service) requireManagedGroupTask(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, taskID uuid.UUID) (*structure.Group, *catalog.TaskDefinition, *catalog.Area, error) {
	group, err := s.requireManagedGroup(ctx, currentUserID, groupID)
	if err != nil {
		return nil, nil, nil, err
	}

	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load task: %w", err)
	}
	if task == nil {
		return nil, nil, nil, fmt.Errorf("задача не найдена")
	}

	area, err := s.areaRepo.FindByID(ctx, task.AreaID())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load task area: %w", err)
	}
	if area == nil {
		return nil, nil, nil, fmt.Errorf("территория не найдена")
	}
	if err := requireAreaInGroup(area, group.ID()); err != nil {
		return nil, nil, nil, err
	}

	return group, task, area, nil
}

func (s *Service) listGroupAreas(ctx context.Context, groupID uuid.UUID) ([]*catalog.Area, error) {
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, fmt.Errorf("load areas: %w", err)
	}

	result := make([]*catalog.Area, 0, len(areas))
	for _, area := range areas {
		if area.GroupID() != nil && *area.GroupID() == groupID {
			result = append(result, area)
		}
	}

	return result, nil
}

type normalizedAreaInput struct {
	name  string
	floor *int
}

func (i normalizedAreaInput) floorValue() int {
	if i.floor == nil {
		return 0
	}
	return *i.floor
}

func normalizeAreaInput(name string, floor *int) (*normalizedAreaInput, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, fmt.Errorf("Введите название")
	}
	if utf8.RuneCountInString(trimmedName) > 255 {
		return nil, fmt.Errorf("Название не должно превышать 255 символов")
	}

	return &normalizedAreaInput{name: trimmedName, floor: floor}, nil
}

type normalizedTaskInput struct {
	title     string
	cost      int
	frequency int
}

func normalizeTaskInput(title string, cost int, frequency int) (*normalizedTaskInput, error) {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return nil, fmt.Errorf("Введите название")
	}
	if utf8.RuneCountInString(trimmedTitle) > 255 {
		return nil, fmt.Errorf("Название не должно превышать 255 символов")
	}
	if cost <= 0 {
		return nil, fmt.Errorf("Стоимость должна быть больше нуля")
	}
	if frequency <= 0 {
		return nil, fmt.Errorf("Частота должна быть больше нуля")
	}

	return &normalizedTaskInput{
		title:     trimmedTitle,
		cost:      cost,
		frequency: frequency,
	}, nil
}

func requireAreaInGroup(area *catalog.Area, groupID uuid.UUID) error {
	if area.GroupID() == nil || *area.GroupID() != groupID {
		return fmt.Errorf("территория принадлежит другой группе")
	}

	return nil
}

func areaFloor(area *catalog.Area) *int {
	if area == nil || area.Floor() == 0 {
		return nil
	}

	floor := area.Floor()
	return &floor
}

func buildAreaDetails(area *catalog.Area, group *structure.Group) *dto.AreaDetails {
	return &dto.AreaDetails{
		ID:    area.ID(),
		Name:  area.Name(),
		Floor: areaFloor(area),
		Group: &dto.GroupOption{
			ID:   group.ID().String(),
			Name: group.Name(),
		},
	}
}

func buildTaskDetails(task *catalog.TaskDefinition, group *structure.Group, area *catalog.Area) *dto.TaskDetails {
	return &dto.TaskDetails{
		ID:        task.ID().String(),
		Title:     task.Title(),
		Cost:      task.Cost(),
		Frequency: task.Frequency(),
		Area: dto.AreaSummary{
			ID:   area.ID(),
			Name: area.Name(),
		},
	}
}
