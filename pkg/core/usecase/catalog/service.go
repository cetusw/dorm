package catalog

import (
	"context"
	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/ports/dto"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

type Service struct {
	taskRepo catalog.TaskDefinitionRepository
	areaRepo catalog.AreaRepository
}

func NewCatalogService(taskRepo catalog.TaskDefinitionRepository, areaRepo catalog.AreaRepository) *Service {
	return &Service{taskRepo: taskRepo, areaRepo: areaRepo}
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
			ID:        task.ID(),
			AreaID:    task.AreaID(),
			AreaName:  area.Name(),
			AreaFloor: area.Floor(),
			Title:     task.Title(),
			Cost:      task.Cost(),
			Frequency: task.Frequency(),
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
			ID:        task.ID(),
			AreaID:    task.AreaID(),
			AreaName:  area.Name(),
			AreaFloor: area.Floor(),
			Title:     task.Title(),
			Cost:      task.Cost(),
			Frequency: task.Frequency(),
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
		ID:        task.ID(),
		AreaID:    task.AreaID(),
		Title:     task.Title(),
		Cost:      task.Cost(),
		Frequency: task.Frequency(),
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
	task, err := catalog.NewTaskDefinition(req.AreaID, req.Title, req.Cost, req.Frequency)
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
	return s.taskRepo.Save(ctx, catalog.RestoreTaskDefinition(id, req.AreaID, req.Title, req.Cost, req.Frequency))
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
	return s.taskRepo.Delete(ctx, id)
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
