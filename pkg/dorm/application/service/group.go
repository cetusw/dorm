package service

import (
	"dorm/pkg/dorm/application/model"
	"dorm/pkg/dorm/infrastructure/mysql/repository"

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
