package team

import (
	"context"
	"fmt"
	"sort"
	"strings"

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

func (s *Service) GetTeamsListByGroup(ctx context.Context, groupID uuid.UUID) ([]dto.TeamListItem, error) {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("group not found")
	}

	teams, err := s.teamRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.TeamListItem, 0, len(teams))
	for _, team := range teams {
		members, err := s.userRepo.FindByTeamID(ctx, team.ID())
		if err != nil {
			return nil, err
		}
		items = append(items, dto.TeamListItem{
			ID:           team.ID(),
			Name:         team.Name(),
			Color:        team.Color(),
			Order:        team.Order(),
			GroupID:      team.GroupID(),
			GroupName:    group.Name(),
			DormitoryID:  group.DormitoryID(),
			MembersCount: len(members),
			LeaderID:     team.LeaderID(),
		})
	}
	return items, nil
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
		ID:       team.ID(),
		Name:     team.Name(),
		Color:    team.Color(),
		Order:    team.Order(),
		GroupID:  team.GroupID(),
		LeaderID: team.LeaderID(),
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
	return s.CreateTeamInGroup(ctx, groupID, req)
}

func (s *Service) CreateTeamInGroup(ctx context.Context, groupID uuid.UUID, req dto.CreateTeamRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("team name is required")
	}
	group, err := s.requireGroup(ctx, groupID)
	if err != nil {
		return err
	}
	memberIDs, err := parseMemberIDs(req.MemberIDs)
	if err != nil {
		return err
	}
	leaderID, err := parseOptionalUUID(req.LeaderID)
	if err != nil {
		return err
	}
	if err := s.requireUsersInDormitory(ctx, group.DormitoryID(), appendOptionalUUID(memberIDs, leaderID)); err != nil {
		return err
	}

	team := structure.NewTeam(req.Name, groupID, normalizeColor(req.Color), req.Order)
	if leaderID != nil {
		team = structure.RestoreTeam(team.ID(), team.Name(), team.GroupID(), leaderID, team.Color(), team.Order())
	}
	if err := s.teamRepo.Save(ctx, team); err != nil {
		return err
	}
	return s.moveMembers(ctx, team.ID(), memberIDs)
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

	return s.UpdateTeamInGroup(ctx, groupID, id, req)
}

func (s *Service) UpdateTeamInGroup(ctx context.Context, groupID uuid.UUID, id uuid.UUID, req dto.UpdateTeamRequest) error {
	current, err := s.teamRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return fmt.Errorf("team not found")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("team name is required")
	}
	group, err := s.requireGroup(ctx, groupID)
	if err != nil {
		return err
	}
	memberIDs, err := parseMemberIDs(req.MemberIDs)
	if err != nil {
		return err
	}
	leaderID, err := parseOptionalUUID(req.LeaderID)
	if err != nil {
		return err
	}
	if err := s.requireUsersInDormitory(ctx, group.DormitoryID(), appendOptionalUUID(memberIDs, leaderID)); err != nil {
		return err
	}

	updated := structure.RestoreTeam(id, req.Name, groupID, leaderID, normalizeColor(req.Color), req.Order)
	if err := s.teamRepo.Save(ctx, updated); err != nil {
		return err
	}
	return s.syncMembers(ctx, id, memberIDs)
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

	return s.getTeamMemberItems(ctx, group.DormitoryID(), teamID)
}

func (s *Service) GetTeamMembersForNewTeam(ctx context.Context, groupID uuid.UUID) ([]dto.TeamMemberItem, error) {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("group not found")
	}
	return s.getTeamMemberItems(ctx, group.DormitoryID(), uuid.Nil)
}

func (s *Service) MoveUserToTeam(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return fmt.Errorf("team not found")
	}
	group, err := s.requireGroup(ctx, team.GroupID())
	if err != nil {
		return err
	}
	if err := s.requireUsersInDormitory(ctx, group.DormitoryID(), []uuid.UUID{userID}); err != nil {
		return err
	}
	return s.userRepo.MoveUserToTeam(ctx, userID, &teamID)
}

func (s *Service) RemoveUserFromTeam(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.MoveUserToTeam(ctx, userID, nil)
}

func (s *Service) requireGroup(ctx context.Context, id uuid.UUID) (*structure.Group, error) {
	group, err := s.groupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("group not found")
	}
	return group, nil
}

func (s *Service) moveMembers(ctx context.Context, teamID uuid.UUID, memberIDs []uuid.UUID) error {
	for _, userID := range memberIDs {
		if err := s.userRepo.MoveUserToTeam(ctx, userID, &teamID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) syncMembers(ctx context.Context, teamID uuid.UUID, memberIDs []uuid.UUID) error {
	selected := make(map[uuid.UUID]struct{}, len(memberIDs))
	for _, userID := range memberIDs {
		selected[userID] = struct{}{}
	}

	currentMembers, err := s.userRepo.FindByTeamID(ctx, teamID)
	if err != nil {
		return err
	}
	for _, resident := range currentMembers {
		if _, ok := selected[resident.ID()]; ok {
			continue
		}
		if err := s.userRepo.MoveUserToTeam(ctx, resident.ID(), nil); err != nil {
			return err
		}
	}

	for _, userID := range memberIDs {
		if err := s.userRepo.MoveUserToTeam(ctx, userID, &teamID); err != nil {
			return err
		}
	}
	return nil
}

func parseMemberIDs(rawIDs []string) ([]uuid.UUID, error) {
	memberIDs := make([]uuid.UUID, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		userID, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("invalid member id")
		}
		memberIDs = append(memberIDs, userID)
	}
	return memberIDs, nil
}

func parseOptionalUUID(rawID string) (*uuid.UUID, error) {
	if strings.TrimSpace(rawID) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	return &id, nil
}

func appendOptionalUUID(ids []uuid.UUID, optionalID *uuid.UUID) []uuid.UUID {
	if optionalID == nil {
		return ids
	}
	return append(ids, *optionalID)
}

func (s *Service) requireUsersInDormitory(ctx context.Context, dormitoryID int64, userIDs []uuid.UUID) error {
	for _, userID := range userIDs {
		resident, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		if resident == nil {
			return fmt.Errorf("user not found")
		}
		if resident.DormitoryID() == nil || *resident.DormitoryID() != dormitoryID {
			return fmt.Errorf("user does not belong to dormitory")
		}
	}
	return nil
}

func normalizeColor(color string) string {
	color = strings.TrimSpace(color)
	return strings.TrimPrefix(color, "#")
}

func (s *Service) getTeamMemberItems(ctx context.Context, dormitoryID int64, currentTeamID uuid.UUID) ([]dto.TeamMemberItem, error) {
	users, err := s.userRepo.FindByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return nil, err
	}

	teams, err := s.GetTeamsByDormitory(ctx, dormitoryID)
	if err != nil {
		return nil, err
	}
	teamNames := make(map[uuid.UUID]string, len(teams))
	for _, t := range teams {
		teamNames[t.ID()] = t.Name()
	}

	items := make([]dto.TeamMemberItem, 0, len(users))
	for _, u := range users {
		fullName := strings.TrimSpace(fmt.Sprintf("%s %s", u.LastName(), u.FirstName()))
		var currentTeamName string
		if u.TeamID() != nil {
			currentTeamName = teamNames[*u.TeamID()]
		}
		items = append(items, dto.TeamMemberItem{
			UserID:          u.ID(),
			FullName:        fullName,
			CurrentTeamID:   u.TeamID(),
			CurrentTeamName: currentTeamName,
			IsInCurrentTeam: currentTeamID != uuid.Nil && u.TeamID() != nil && *u.TeamID() == currentTeamID,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsInCurrentTeam != items[j].IsInCurrentTeam {
			return items[i].IsInCurrentTeam
		}
		return strings.ToLower(items[i].FullName) < strings.ToLower(items[j].FullName)
	})
	return items, nil
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
