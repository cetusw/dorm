package catalog

import (
	"context"
	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/ports/dto"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	taskRepo catalog.TaskDefinitionRepository
	areaRepo catalog.AreaRepository
}

func NewCatalogService(taskRepo catalog.TaskDefinitionRepository, areaRepo catalog.AreaRepository) *Service {
	return &Service{taskRepo: taskRepo, areaRepo: areaRepo}
}

func (s *Service) ListTasks(ctx context.Context) ([]dto.TaskCatalogItem, error) {
	tasks, err := s.taskRepo.GetAllTaskDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	areas, err := s.areaRepo.GetAllAreas(ctx)
	if err != nil {
		return nil, err
	}
	areaMap := make(map[int]*catalog.Area, len(areas))
	for _, a := range areas {
		areaMap[a.ID()] = a
	}
	out := make([]dto.TaskCatalogItem, 0, len(tasks))
	for _, t := range tasks {
		areaName := ""
		if a := areaMap[t.AreaID()]; a != nil {
			areaName = a.Name()
		}
		out = append(out, dto.TaskCatalogItem{
			ID:        t.ID(),
			AreaID:    t.AreaID(),
			AreaName:  areaName,
			Title:     t.Title(),
			Cost:      t.Cost(),
			Frequency: t.Frequency(),
		})
	}
	return out, nil
}

func (s *Service) ListAreas(ctx context.Context) ([]*catalog.Area, error) {
	return s.areaRepo.GetAllAreas(ctx)
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

func (s *Service) UpdateTask(ctx context.Context, id uuid.UUID, req dto.UpsertTaskCatalogRequest) error {
	if req.AreaID == 0 || req.Title == "" {
		return fmt.Errorf("area and title are required")
	}
	return s.taskRepo.Save(ctx, catalog.RestoreTaskDefinition(id, req.AreaID, req.Title, req.Cost, req.Frequency))
}

func (s *Service) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return s.taskRepo.Delete(ctx, id)
}
