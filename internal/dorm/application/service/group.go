package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"

	"github.com/google/uuid"
)

type GroupService struct {
	groupRepository *repository.GroupRepository
}

func NewGroupService(repository *repository.GroupRepository) *GroupService {
	return &GroupService{
		groupRepository: repository,
	}
}

func (s *GroupService) GetAllGroups() ([]model.Group, error) {
	return s.groupRepository.FindAll()
}

func (s *GroupService) GetGroup(groupID uuid.UUID) (*model.Group, error) {
	return s.groupRepository.Find(groupID)
}
