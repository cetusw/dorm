package service

import (
	"dorm/internal/dorm/infrastructure/mysql/repository"
	"sort"
)

type TeamService struct {
	teamRepository *repository.TeamRepository
}

func NewTeamService(teamRepository *repository.TeamRepository) *TeamService {
	return &TeamService{
		teamRepository: teamRepository,
	}
}

func (s *TeamService) GetSortedTeamIDs() ([]int, error) {
	teams, err := s.teamRepository.FindAll()
	if err != nil {
		return nil, err
	}
	var teamIDs []int
	for _, team := range teams {
		teamIDs = append(teamIDs, team.TeamID)
	}

	sort.Ints(teamIDs)

	return teamIDs, nil
}

func (s *TeamService) GetTeamColor(teamID int) (string, error) {
	team, err := s.teamRepository.Find(teamID)
	if err != nil {
		return "", err
	}
	if team == nil {
		return "", nil
	}
	return team.Color, nil
}
