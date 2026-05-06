package team

import (
	"context"
	"fmt"

	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"
	ports "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
)

type Service struct {
	teamRepo      structure.TeamRepository
	groupRepo     structure.GroupRepository
	dormitoryRepo structure.DormitoryRepository
	userRepo      user.Repository
	queryService  ports.TeamQueryService
}

func NewTeamService(
	teamRepo structure.TeamRepository,
	groupRepo structure.GroupRepository,
	dormitoryRepo structure.DormitoryRepository,
	userRepo user.Repository,
	queryService ports.TeamQueryService,
) *Service {
	return &Service{
		teamRepo:      teamRepo,
		groupRepo:     groupRepo,
		dormitoryRepo: dormitoryRepo,
		userRepo:      userRepo,
		queryService:  queryService,
	}
}

func (s *Service) GetTeamsByDormitory(ctx context.Context, dormID int64) ([]*structure.Team, error) {
	return s.queryService.FindTeamsByDormitoryID(ctx, dormID)
}

func (s *Service) GetGroups(ctx context.Context) ([]*structure.Group, error) {
	return s.groupRepo.FindAll(ctx)
}

func (s *Service) GetTeamsList(ctx context.Context, dormID int64) ([]dto.TeamListItem, error) {
	return s.queryService.GetTeamsDetailedList(ctx, dormID)
}

func (s *Service) GetTeamByID(ctx context.Context, id uuid.UUID) (*dto.TeamListItem, error) {
	team, err := s.teamRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}
	groupMap, dormMap, err := s.loadGroupDormitoryMaps(ctx)
	if err != nil {
		return nil, err
	}
	group := groupMap[team.GroupID()]
	item := &dto.TeamListItem{
		ID:      team.ID(),
		Name:    team.Name(),
		Color:   team.Color(),
		Order:   team.Order(),
		GroupID: team.GroupID(),
	}
	if group != nil {
		item.GroupName = group.Name()
		item.DormitoryID = group.DormitoryID()
		if dorm := dormMap[group.DormitoryID()]; dorm != nil {
			item.DormitoryName = dorm.Name()
		}
	}
	members, err := s.userRepo.FindByTeamID(ctx, team.ID())
	if err != nil {
		return nil, err
	}
	item.MembersCount = len(members)
	return item, nil
}

func (s *Service) CreateTeam(ctx context.Context, req dto.CreateTeamRequest) error {
	groupID, err := uuid.Parse(req.GroupID)
	if err != nil {
		return fmt.Errorf("invalid group id")
	}
	team := structure.NewTeam(req.Name, groupID, req.Color, req.Order)
	return s.teamRepo.Save(ctx, team)
}

func (s *Service) UpdateTeam(ctx context.Context, id uuid.UUID, req dto.UpdateTeamRequest) error {
	current, err := s.teamRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return fmt.Errorf("team not found")
	}

	groupID, err := uuid.Parse(req.GroupID)
	if err != nil {
		return fmt.Errorf("invalid group id")
	}

	updated := structure.RestoreTeam(id, req.Name, groupID, current.LeaderID(), req.Color, req.Order)
	return s.teamRepo.Save(ctx, updated)
}

func (s *Service) DeleteTeam(ctx context.Context, id uuid.UUID) error {
	return s.teamRepo.Delete(ctx, id)
}

func (s *Service) GetTeamMembersForEdit(ctx context.Context, teamID uuid.UUID) ([]dto.TeamMemberItem, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, fmt.Errorf("team not found")
	}
	group, err := s.groupRepo.FindByID(ctx, team.GroupID())
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("group not found")
	}

	users, err := s.userRepo.FindByDormitoryID(ctx, group.DormitoryID())
	if err != nil {
		return nil, err
	}

	teams, err := s.GetTeamsByDormitory(ctx, group.DormitoryID())
	if err != nil {
		return nil, err
	}
	teamNames := make(map[uuid.UUID]string, len(teams))
	for _, t := range teams {
		teamNames[t.ID()] = t.Name()
	}

	items := make([]dto.TeamMemberItem, 0, len(users))
	for _, u := range users {
		fullName := fmt.Sprintf("%s %s", u.FirstName(), u.LastName())
		var currentTeamName string
		if u.TeamID() != nil {
			currentTeamName = teamNames[*u.TeamID()]
		}
		items = append(items, dto.TeamMemberItem{
			UserID:          u.ID(),
			FullName:        fullName,
			CurrentTeamID:   u.TeamID(),
			CurrentTeamName: currentTeamName,
			IsInCurrentTeam: u.TeamID() != nil && *u.TeamID() == teamID,
		})
	}
	return items, nil
}

func (s *Service) MoveUserToTeam(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	return s.userRepo.MoveUserToTeam(ctx, userID, &teamID)
}

func (s *Service) RemoveUserFromTeam(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.MoveUserToTeam(ctx, userID, nil)
}

func (s *Service) loadGroupDormitoryMaps(ctx context.Context) (map[uuid.UUID]*structure.Group, map[int64]*structure.Dormitory, error) {
	groups, err := s.groupRepo.FindAll(ctx)
	if err != nil {
		return nil, nil, err
	}
	groupMap := make(map[uuid.UUID]*structure.Group, len(groups))
	dormitoryIDs := make(map[int64]struct{})
	for _, g := range groups {
		groupMap[g.ID()] = g
		dormitoryIDs[g.DormitoryID()] = struct{}{}
	}

	dormMap := make(map[int64]*structure.Dormitory, len(dormitoryIDs))
	for dormID := range dormitoryIDs {
		dorm, err := s.dormitoryRepo.FindByID(ctx, dormID)
		if err != nil {
			return nil, nil, err
		}
		if dorm != nil {
			dormMap[dormID] = dorm
		}
	}
	return groupMap, dormMap, nil
}
