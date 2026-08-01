package catalog

import (
	"context"
	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/ports/dto"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Service struct {
	taskRepo  catalog.TaskDefinitionRepository
	areaRepo  catalog.AreaRepository
	groupRepo structure.GroupRepository
}

func NewCatalogService(
	taskRepo catalog.TaskDefinitionRepository,
	areaRepo catalog.AreaRepository,
	groupRepo structure.GroupRepository,
) *Service {
	return &Service{taskRepo: taskRepo, areaRepo: areaRepo, groupRepo: groupRepo}
}

func (s *Service) ListTaskGroupsByGroup(ctx context.Context, groupID uuid.UUID) ([]dto.TaskCatalogGroup, error) {
	tasks, err := s.taskRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	areas, err := s.ListAreasByGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	areaMap := make(map[int]*catalog.Area, len(areas))
	for _, area := range areas {
		areaMap[area.ID()] = area
	}
	items := make([]dto.TaskCatalogItem, 0, len(tasks))
	for _, task := range tasks {
		area := areaMap[task.AreaID()]
		if area == nil {
			continue
		}
		items = append(items, dto.TaskCatalogItem{
			ID:                 task.ID(),
			AreaID:             task.AreaID(),
			AreaName:           area.Name(),
			AreaFloor:          area.Floor(),
			Title:              task.Title(),
			Cost:               task.Cost(),
			RecurrenceInterval: task.RecurrenceInterval(),
			StartSequence:      task.StartSequence(),
		})
	}
	return groupTaskCatalogItems(items), nil
}

func (s *Service) ListCommonTaskGroups(ctx context.Context) ([]dto.TaskCatalogGroup, error) {
	tasks, err := s.taskRepo.FindCommon(ctx)
	if err != nil {
		return nil, err
	}
	areas, err := s.ListCommonAreas(ctx)
	if err != nil {
		return nil, err
	}
	areaMap := make(map[int]*catalog.Area, len(areas))
	for _, area := range areas {
		areaMap[area.ID()] = area
	}
	items := make([]dto.TaskCatalogItem, 0, len(tasks))
	for _, task := range tasks {
		area := areaMap[task.AreaID()]
		if area == nil {
			continue
		}
		items = append(items, dto.TaskCatalogItem{
			ID:                 task.ID(),
			AreaID:             task.AreaID(),
			AreaName:           area.Name(),
			AreaFloor:          area.Floor(),
			Title:              task.Title(),
			Cost:               task.Cost(),
			RecurrenceInterval: task.RecurrenceInterval(),
			StartSequence:      task.StartSequence(),
		})
	}
	return groupTaskCatalogItems(items), nil
}

func groupTaskCatalogItems(tasks []dto.TaskCatalogItem) []dto.TaskCatalogGroup {
	groupMap := make(map[int]*dto.TaskCatalogGroup)
	for _, task := range tasks {
		group := groupMap[task.AreaID]
		if group == nil {
			group = &dto.TaskCatalogGroup{
				AreaID:   task.AreaID,
				AreaName: task.AreaName,
				Floor:    task.AreaFloor,
			}
			groupMap[task.AreaID] = group
		}
		group.Tasks = append(group.Tasks, task)
	}
	groups := make([]dto.TaskCatalogGroup, 0, len(groupMap))
	for _, group := range groupMap {
		sort.Slice(group.Tasks, func(i, j int) bool {
			return group.Tasks[i].Title < group.Tasks[j].Title
		})
		groups = append(groups, *group)
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Floor != groups[j].Floor {
			return groups[i].Floor > groups[j].Floor
		}
		return groups[i].AreaName < groups[j].AreaName
	})
	return groups
}

func (s *Service) ListAreas(ctx context.Context) ([]*catalog.Area, error) {
	return s.areaRepo.GetAllAreas(ctx)
}

func (s *Service) ListAreasByGroup(ctx context.Context, groupID uuid.UUID) ([]*catalog.Area, error) {
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, err
	}
	groupAreas := make([]*catalog.Area, 0, len(areas))
	for _, area := range areas {
		if area.GroupID() != nil && *area.GroupID() == groupID {
			groupAreas = append(groupAreas, area)
		}
	}
	return groupAreas, nil
}

func (s *Service) ListCommonAreas(ctx context.Context) ([]*catalog.Area, error) {
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, err
	}
	commonAreas := make([]*catalog.Area, 0, len(areas))
	for _, area := range areas {
		if area.GroupID() == nil {
			commonAreas = append(commonAreas, area)
		}
	}
	return commonAreas, nil
}

func (s *Service) ListAreasByDormitory(ctx context.Context, dormitoryID int64) ([]*catalog.Area, error) {
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, err
	}

	groups, err := s.groupRepo.FindByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return nil, err
	}
	groupIDs := make(map[uuid.UUID]struct{}, len(groups))
	for _, group := range groups {
		groupIDs[group.ID()] = struct{}{}
	}

	result := make([]*catalog.Area, 0, len(areas))
	for _, area := range areas {
		if area.GroupID() == nil {
			result = append(result, area)
			continue
		}

		if _, ok := groupIDs[*area.GroupID()]; ok {
			result = append(result, area)
		}
	}

	return result, nil
}

func (s *Service) GetAreasResponse(ctx context.Context, dormitoryID int64) (dto.AreaListResponse, error) {
	areas, err := s.ListAreasByDormitory(ctx, dormitoryID)
	if err != nil {
		return dto.AreaListResponse{}, fmt.Errorf("load areas: %w", err)
	}

	groups, err := s.groupRepo.FindByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return dto.AreaListResponse{}, fmt.Errorf("load area groups: %w", err)
	}
	groupMap := make(map[uuid.UUID]*structure.Group, len(groups))
	for _, group := range groups {
		groupMap[group.ID()] = group
	}

	items := make([]dto.AreaResponseItem, 0, len(areas))
	for _, area := range areas {
		items = append(items, dto.AreaResponseItem{
			ID:    area.ID(),
			Name:  area.Name(),
			Floor: areaFloor(area),
			Group: areaGroupSummary(area, groupMap),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Floor == nil && items[j].Floor != nil {
			return false
		}
		if items[i].Floor != nil && items[j].Floor == nil {
			return true
		}
		if items[i].Floor != nil && items[j].Floor != nil && *items[i].Floor != *items[j].Floor {
			return *items[i].Floor > *items[j].Floor
		}
		return items[i].Name < items[j].Name
	})

	return dto.AreaListResponse{Areas: items}, nil
}

func (s *Service) GetAreaDetails(ctx context.Context, id int) (*dto.AreaDetails, error) {
	area, err := s.areaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load area: %w", err)
	}
	if area == nil {
		return nil, nil
	}

	groupMap, err := s.allGroupsMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("load area groups: %w", err)
	}

	return &dto.AreaDetails{
		ID:    area.ID(),
		Name:  area.Name(),
		Floor: areaFloor(area),
		Group: areaGroupSummary(area, groupMap),
	}, nil
}

func (s *Service) GetTasksResponse(ctx context.Context, dormitoryID int64) (dto.TaskListResponse, error) {
	tasks, err := s.taskRepo.GetAllTaskDefinitions(ctx)
	if err != nil {
		return dto.TaskListResponse{}, fmt.Errorf("load tasks: %w", err)
	}

	areas, err := s.ListAreasByDormitory(ctx, dormitoryID)
	if err != nil {
		return dto.TaskListResponse{}, fmt.Errorf("load task areas: %w", err)
	}

	areaMap := make(map[int]*catalog.Area, len(areas))
	for _, area := range areas {
		areaMap[area.ID()] = area
	}

	items := make([]dto.TaskResponseItem, 0, len(tasks))
	for _, task := range tasks {
		area, ok := areaMap[task.AreaID()]
		if !ok {
			continue
		}

		items = append(items, dto.TaskResponseItem{
			ID:                 task.ID().String(),
			Title:              task.Title(),
			Cost:               task.Cost(),
			RecurrenceInterval: task.RecurrenceInterval(),
			StartSequence:      task.StartSequence(),
			Area: dto.AreaSummary{
				ID:   area.ID(),
				Name: area.Name(),
			},
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Area.Name != items[j].Area.Name {
			return items[i].Area.Name < items[j].Area.Name
		}
		return items[i].Title < items[j].Title
	})

	return dto.TaskListResponse{Tasks: items}, nil
}

func (s *Service) GetTaskDetails(ctx context.Context, dormitoryID int64, id uuid.UUID) (*dto.TaskDetails, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load task: %w", err)
	}
	if task == nil {
		return nil, nil
	}

	area, err := s.areaRepo.FindByID(ctx, task.AreaID())
	if err != nil {
		return nil, fmt.Errorf("load task area: %w", err)
	}
	if area == nil {
		return nil, fmt.Errorf("территория не найдена")
	}
	if err := s.requireAreaInDormitory(ctx, area, dormitoryID); err != nil {
		return nil, err
	}

	return &dto.TaskDetails{
		ID:                 task.ID().String(),
		Title:              task.Title(),
		Cost:               task.Cost(),
		RecurrenceInterval: task.RecurrenceInterval(),
		StartSequence:      task.StartSequence(),
		Area: dto.AreaSummary{
			ID:   area.ID(),
			Name: area.Name(),
		},
	}, nil
}

func (s *Service) CreateArea(ctx context.Context, dormitoryID int64, req dto.CreateAreaRequest) (*dto.AreaDetails, error) {
	input, err := s.normalizeAreaInput(ctx, dormitoryID, req.Name, req.GroupID, req.Floor)
	if err != nil {
		return nil, err
	}

	area := catalog.NewArea(input.name, input.floorValue(), input.groupID)
	if err := s.areaRepo.Save(ctx, area); err != nil {
		return nil, fmt.Errorf("create area: %w", err)
	}

	return s.GetAreaDetails(ctx, area.ID())
}

func (s *Service) CreateTaskDetails(ctx context.Context, dormitoryID int64, req dto.CreateTaskRequest) (*dto.TaskDetails, error) {
	input, err := s.normalizeTaskInput(ctx, dormitoryID, req.Title, req.Cost, req.RecurrenceInterval, req.AreaID)
	if err != nil {
		return nil, err
	}

	task, err := catalog.NewTaskDefinition(input.areaID, input.title, input.cost, input.recurrenceInterval, 1)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return s.GetTaskDetails(ctx, dormitoryID, task.ID())
}

func (s *Service) UpdateArea(ctx context.Context, dormitoryID int64, id int, req dto.UpdateAreaRequest) (*dto.AreaDetails, error) {
	area, err := s.areaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load area: %w", err)
	}
	if area == nil {
		return nil, nil
	}

	if err := s.requireAreaInDormitory(ctx, area, dormitoryID); err != nil {
		return nil, err
	}

	input, err := s.normalizeAreaInput(ctx, dormitoryID, req.Name, req.GroupID, req.Floor)
	if err != nil {
		return nil, err
	}

	updated := catalog.RestoreArea(id, input.name, input.floorValue(), input.groupID)
	if err := s.areaRepo.Save(ctx, updated); err != nil {
		return nil, fmt.Errorf("update area: %w", err)
	}

	return s.GetAreaDetails(ctx, updated.ID())
}

func (s *Service) UpdateTaskDetails(ctx context.Context, dormitoryID int64, id uuid.UUID, req dto.UpdateTaskRequest) (*dto.TaskDetails, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load task: %w", err)
	}
	if task == nil {
		return nil, nil
	}

	currentArea, err := s.areaRepo.FindByID(ctx, task.AreaID())
	if err != nil {
		return nil, fmt.Errorf("load current task area: %w", err)
	}
	if currentArea == nil {
		return nil, fmt.Errorf("текущая территория задачи не найдена")
	}
	if err := s.requireAreaInDormitory(ctx, currentArea, dormitoryID); err != nil {
		return nil, err
	}

	input, err := s.normalizeTaskInput(ctx, dormitoryID, req.Title, req.Cost, req.RecurrenceInterval, req.AreaID)
	if err != nil {
		return nil, err
	}

	updated := catalog.RestoreTaskDefinition(
		id,
		input.areaID,
		input.title,
		input.cost,
		input.recurrenceInterval,
		task.StartSequence(),
	)
	if err := s.taskRepo.Save(ctx, updated); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	return s.GetTaskDetails(ctx, dormitoryID, updated.ID())
}

func (s *Service) DeleteArea(ctx context.Context, dormitoryID int64, id int) error {
	area, err := s.areaRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("load area: %w", err)
	}
	if area == nil {
		return nil
	}

	if err := s.requireAreaInDormitory(ctx, area, dormitoryID); err != nil {
		return err
	}

	if err := s.areaRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete area: %w", err)
	}

	return nil
}

func (s *Service) DeleteTaskDetails(ctx context.Context, dormitoryID int64, id uuid.UUID) error {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("load task: %w", err)
	}
	if task == nil {
		return nil
	}

	area, err := s.areaRepo.FindByID(ctx, task.AreaID())
	if err != nil {
		return fmt.Errorf("load task area: %w", err)
	}
	if area == nil {
		return fmt.Errorf("территория не найдена")
	}
	if err := s.requireAreaInDormitory(ctx, area, dormitoryID); err != nil {
		return err
	}

	if err := s.taskRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (s *Service) GetTask(ctx context.Context, id uuid.UUID) (*dto.TaskCatalogItem, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, err
	}
	item := &dto.TaskCatalogItem{
		ID:                 task.ID(),
		AreaID:             task.AreaID(),
		Title:              task.Title(),
		Cost:               task.Cost(),
		RecurrenceInterval: task.RecurrenceInterval(),
		StartSequence:      task.StartSequence(),
	}
	for _, a := range areas {
		if a.ID() == task.AreaID() {
			item.AreaName = a.Name()
			item.AreaFloor = a.Floor()
			break
		}
	}
	return item, nil
}

func (s *Service) CreateTask(ctx context.Context, req dto.UpsertTaskCatalogRequest) error {
	task, err := catalog.NewTaskDefinition(req.AreaID, req.Title, req.Cost, req.RecurrenceInterval, 1)
	if err != nil {
		return err
	}
	return s.taskRepo.Save(ctx, task)
}

func (s *Service) CreateTaskInGroup(ctx context.Context, groupID uuid.UUID, req dto.UpsertTaskCatalogRequest) error {
	if err := s.requireAreaInGroup(ctx, req.AreaID, groupID); err != nil {
		return err
	}
	return s.CreateTask(ctx, req)
}

func (s *Service) CreateCommonTask(ctx context.Context, req dto.UpsertTaskCatalogRequest) error {
	area, err := s.defaultCommonArea(ctx)
	if err != nil {
		return err
	}
	req.AreaID = area.ID()
	return s.CreateTask(ctx, req)
}

func (s *Service) UpdateTask(ctx context.Context, id uuid.UUID, req dto.UpsertTaskCatalogRequest) error {
	if req.AreaID == 0 || req.Title == "" {
		return fmt.Errorf("area and title are required")
	}
	existing, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}
	return s.taskRepo.Save(ctx, catalog.RestoreTaskDefinition(
		id,
		req.AreaID,
		req.Title,
		req.Cost,
		req.RecurrenceInterval,
		existing.StartSequence(),
	))
}

func (s *Service) UpdateTaskInGroup(ctx context.Context, groupID uuid.UUID, id uuid.UUID, req dto.UpsertTaskCatalogRequest) error {
	if err := s.requireTaskInGroup(ctx, id, groupID); err != nil {
		return err
	}
	if err := s.requireAreaInGroup(ctx, req.AreaID, groupID); err != nil {
		return err
	}
	return s.UpdateTask(ctx, id, req)
}

func (s *Service) UpdateCommonTask(ctx context.Context, id uuid.UUID, req dto.UpsertTaskCatalogRequest) error {
	existing, err := s.requireCommonTask(ctx, id)
	if err != nil {
		return err
	}
	req.AreaID = existing.AreaID()
	return s.UpdateTask(ctx, id, req)
}

func (s *Service) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return s.taskRepo.SoftDelete(ctx, id)
}

func (s *Service) DeleteTaskInGroup(ctx context.Context, groupID uuid.UUID, id uuid.UUID) error {
	if err := s.requireTaskInGroup(ctx, id, groupID); err != nil {
		return err
	}
	return s.DeleteTask(ctx, id)
}

func (s *Service) DeleteCommonTask(ctx context.Context, id uuid.UUID) error {
	if _, err := s.requireCommonTask(ctx, id); err != nil {
		return err
	}
	return s.DeleteTask(ctx, id)
}

func (s *Service) requireTaskInGroup(ctx context.Context, taskID uuid.UUID, groupID uuid.UUID) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}
	return s.requireAreaInGroup(ctx, task.AreaID(), groupID)
}

func (s *Service) requireAreaInGroup(ctx context.Context, areaID int, groupID uuid.UUID) error {
	areas, err := s.ListAreasByGroup(ctx, groupID)
	if err != nil {
		return err
	}
	for _, area := range areas {
		if area.ID() == areaID {
			return nil
		}
	}
	return fmt.Errorf("area does not belong to group")
}

func (s *Service) requireCommonTask(ctx context.Context, taskID uuid.UUID) (*catalog.TaskDefinition, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}
	commonAreas, err := s.ListCommonAreas(ctx)
	if err != nil {
		return nil, err
	}
	for _, area := range commonAreas {
		if area.ID() == task.AreaID() {
			return task, nil
		}
	}
	return nil, fmt.Errorf("task is not common")
}

func (s *Service) defaultCommonArea(ctx context.Context) (*catalog.Area, error) {
	areas, err := s.ListCommonAreas(ctx)
	if err != nil {
		return nil, err
	}
	if len(areas) == 0 {
		return nil, fmt.Errorf("common area not found")
	}
	return areas[0], nil
}

type normalizedAreaInput struct {
	name    string
	groupID *uuid.UUID
	floor   *int
}

type normalizedTaskInput struct {
	title              string
	cost               int
	recurrenceInterval int
	areaID             int
}

func (i *normalizedAreaInput) floorValue() int {
	if i.floor == nil {
		return 0
	}
	return *i.floor
}

func (s *Service) normalizeAreaInput(
	ctx context.Context,
	dormitoryID int64,
	name string,
	groupID *string,
	floor *int,
) (*normalizedAreaInput, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, fmt.Errorf("Введите название")
	}
	if utf8.RuneCountInString(trimmedName) > 255 {
		return nil, fmt.Errorf("Название не должно превышать 255 символов")
	}

	parsedGroupID, err := parseOptionalStringUUIDPointer(groupID)
	if err != nil {
		return nil, err
	}

	if parsedGroupID != nil {
		group, err := s.groupRepo.FindByID(ctx, *parsedGroupID)
		if err != nil {
			return nil, fmt.Errorf("load group: %w", err)
		}
		if group == nil {
			return nil, fmt.Errorf("группа не найдена")
		}
		if group.DormitoryID() != dormitoryID {
			return nil, fmt.Errorf("группа принадлежит другому общежитию")
		}
	}

	return &normalizedAreaInput{
		name:    trimmedName,
		groupID: parsedGroupID,
		floor:   floor,
	}, nil
}

func (s *Service) normalizeTaskInput(
	ctx context.Context,
	dormitoryID int64,
	title string,
	cost int,
	recurrenceInterval int,
	areaID int,
) (*normalizedTaskInput, error) {
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
	if !catalog.IsValidRecurrenceInterval(recurrenceInterval) {
		return nil, fmt.Errorf("Выберите корректную частоту")
	}
	if areaID <= 0 {
		return nil, fmt.Errorf("Выберите территорию")
	}

	area, err := s.areaRepo.FindByID(ctx, areaID)
	if err != nil {
		return nil, fmt.Errorf("load area: %w", err)
	}
	if area == nil {
		return nil, fmt.Errorf("территория не найдена")
	}
	if err := s.requireAreaInDormitory(ctx, area, dormitoryID); err != nil {
		return nil, err
	}

	return &normalizedTaskInput{
		title:              trimmedTitle,
		cost:               cost,
		recurrenceInterval: recurrenceInterval,
		areaID:             areaID,
	}, nil
}

func (s *Service) requireAreaInDormitory(ctx context.Context, area *catalog.Area, dormitoryID int64) error {
	if area.GroupID() == nil {
		return nil
	}

	group, err := s.groupRepo.FindByID(ctx, *area.GroupID())
	if err != nil {
		return fmt.Errorf("load area group: %w", err)
	}
	if group == nil {
		return fmt.Errorf("группа не найдена")
	}
	if group.DormitoryID() != dormitoryID {
		return fmt.Errorf("территория принадлежит другому общежитию")
	}

	return nil
}

func (s *Service) allGroupsMap(ctx context.Context) (map[uuid.UUID]*structure.Group, error) {
	groups, err := s.groupRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	groupMap := make(map[uuid.UUID]*structure.Group, len(groups))
	for _, group := range groups {
		groupMap[group.ID()] = group
	}

	return groupMap, nil
}

func areaFloor(area *catalog.Area) *int {
	if area == nil || area.Floor() == 0 {
		return nil
	}

	floor := area.Floor()
	return &floor
}

func areaGroupSummary(area *catalog.Area, groups map[uuid.UUID]*structure.Group) *dto.GroupOption {
	if area == nil || area.GroupID() == nil {
		return nil
	}

	group, ok := groups[*area.GroupID()]
	if !ok || group == nil {
		return nil
	}

	return &dto.GroupOption{
		ID:   group.ID().String(),
		Name: group.Name(),
	}
}

func parseOptionalStringUUIDPointer(rawID *string) (*uuid.UUID, error) {
	if rawID == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*rawID)
	if trimmed == "" {
		return nil, nil
	}

	id, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	return &id, nil
}
