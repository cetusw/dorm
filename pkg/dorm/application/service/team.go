package service

import (
	"dorm/pkg/dorm/application/model"
	"dorm/pkg/dorm/infrastructure/mysql/repository"

	"github.com/google/uuid"
)

type TeamService struct {
	teamRepository *repository.TeamRepository
}

func NewTeamService(repository *repository.TeamRepository) *TeamService {
	return &TeamService{
		teamRepository: repository,
	}
}

func (s *TeamService) GetTeam(teamID uuid.UUID) (*model.Team, error) {
	return s.teamRepository.Find(teamID)
}

func (s *TeamService) GetTeamByGroupIDAndOrder(groupID uuid.UUID, order int) (*model.Team, error) {
	return s.teamRepository.FindTeamByGroupIDAndOrder(groupID, order)
}

func (s *TeamService) GetTeamColor(teamID uuid.UUID) (string, error) {
	team, err := s.teamRepository.Find(teamID)
	if err != nil {
		return "", err
	}
	if team == nil {
		return "", nil
	}
	return team.Color, nil
}

func (s *TeamService) GetGroupTeams(groupID uuid.UUID) ([]model.Team, error) {
	return s.teamRepository.FindTeamsByGroupID(groupID)
}

func (s *TeamService) GetFirstGroupTeam(groupID uuid.UUID) (*model.Team, error) {
	return s.teamRepository.FindGroupTeamByOrder(groupID, 1)
}
