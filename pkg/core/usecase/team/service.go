package team

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

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

func (s *Service) GetTeamsResponseByGroup(ctx context.Context, groupID uuid.UUID) (dto.TeamListResponse, error) {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return dto.TeamListResponse{}, fmt.Errorf("load group: %w", err)
	}
	if group == nil {
		return dto.TeamListResponse{}, fmt.Errorf("group not found")
	}

	teams, err := s.teamRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return dto.TeamListResponse{}, fmt.Errorf("load teams: %w", err)
	}

	residents, err := s.userRepo.FindByDormitoryID(ctx, group.DormitoryID())
	if err != nil {
		return dto.TeamListResponse{}, fmt.Errorf("load residents: %w", err)
	}
	residentNames := userNames(residents)

	items := make([]dto.TeamResponseItem, 0, len(teams))
	for _, team := range teams {
		members, err := s.userRepo.FindByTeamID(ctx, team.ID())
		if err != nil {
			return dto.TeamListResponse{}, fmt.Errorf("load team members: %w", err)
		}

		items = append(items, dto.TeamResponseItem{
			ID:           team.ID().String(),
			Leader:       userSummaryFromMap(residentNames, team.LeaderID()),
			MembersCount: len(members),
		})
	}

	return dto.TeamListResponse{Teams: items}, nil
}

func (s *Service) GetTeamDetails(ctx context.Context, id uuid.UUID) (*dto.TeamDetails, error) {
	team, err := s.teamRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load team: %w", err)
	}
	if team == nil {
		return nil, nil
	}

	group, err := s.groupRepo.FindByID(ctx, team.GroupID())
	if err != nil {
		return nil, fmt.Errorf("load group: %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("group not found")
	}

	residents, err := s.userRepo.FindByDormitoryID(ctx, group.DormitoryID())
	if err != nil {
		return nil, fmt.Errorf("load residents: %w", err)
	}
	residentNames := userNames(residents)

	members, err := s.userRepo.FindByTeamID(ctx, team.ID())
	if err != nil {
		return nil, fmt.Errorf("load team members: %w", err)
	}
	memberIDs := make([]string, 0, len(members))
	for _, member := range members {
		memberIDs = append(memberIDs, member.ID().String())
	}

	return &dto.TeamDetails{
		ID:        team.ID().String(),
		GroupID:   team.GroupID().String(),
		GroupName: group.Name(),
		Leader:    userSummaryFromMap(residentNames, team.LeaderID()),
		MemberIDs: memberIDs,
	}, nil
}

func (s *Service) CreateResidentTeam(ctx context.Context, req dto.CreateResidentTeamRequest) (*dto.TeamDetails, error) {
	groupID, err := uuid.Parse(req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	group, err := s.requireGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	memberIDs, err := parseMemberIDs(req.MemberIDs)
	if err != nil {
		return nil, err
	}
	leaderID, err := parseRequiredStringUUID(req.LeaderID)
	if err != nil {
		return nil, err
	}
	memberIDs = appendUniqueUUID(memberIDs, leaderID)
	if err := s.requireUsersInDormitory(ctx, group.DormitoryID(), memberIDs); err != nil {
		return nil, err
	}

	team := structure.NewTeam(groupID, leaderID, "", 1)
	if err := s.teamRepo.CreateWithLeader(ctx, team); err != nil {
		return nil, err
	}
	if err := s.moveMembers(ctx, team.ID(), memberIDs); err != nil {
		return nil, err
	}

	return s.GetTeamDetails(ctx, team.ID())
}

func (s *Service) UpdateResidentTeam(ctx context.Context, id uuid.UUID, req dto.UpdateResidentTeamRequest) (*dto.TeamDetails, error) {
	current, err := s.teamRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load team: %w", err)
	}
	if current == nil {
		return nil, nil
	}

	groupID, err := uuid.Parse(req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group id")
	}
	if current.GroupID() != groupID {
		return nil, fmt.Errorf("team does not belong to group")
	}

	group, err := s.requireGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	memberIDs, err := parseMemberIDs(req.MemberIDs)
	if err != nil {
		return nil, err
	}
	leaderID, err := parseRequiredStringUUID(req.LeaderID)
	if err != nil {
		return nil, err
	}
	memberIDs = appendUniqueUUID(memberIDs, leaderID)
	if err := s.requireUsersInDormitory(ctx, group.DormitoryID(), memberIDs); err != nil {
		return nil, err
	}

	updated := structure.RestoreTeam(id, groupID, leaderID, current.Color(), current.RotationPosition())
	if err := s.teamRepo.Save(ctx, updated); err != nil {
		return nil, err
	}
	if err := s.syncMembers(ctx, id, memberIDs); err != nil {
		return nil, err
	}

	return s.GetTeamDetails(ctx, id)
}

func (s *Service) DeleteTeam(ctx context.Context, id uuid.UUID) error {
	return s.teamRepo.Delete(ctx, id)
}

func (s *Service) GetTeamMemberOptionsResponse(ctx context.Context, groupID uuid.UUID, teamID *uuid.UUID) (dto.TeamMemberOptionsResponse, error) {
	currentTeamID := uuid.Nil
	if teamID != nil {
		currentTeamID = *teamID
	}

	group, err := s.requireGroup(ctx, groupID)
	if err != nil {
		return dto.TeamMemberOptionsResponse{}, err
	}

	items, err := s.getTeamMemberItems(ctx, group.DormitoryID(), currentTeamID)
	if err != nil {
		return dto.TeamMemberOptionsResponse{}, err
	}

	responseItems := make([]dto.TeamMemberOptionItem, 0, len(items))
	for _, item := range items {
		currentTeamIDValue := optionalUUIDString(item.CurrentTeamID)
		responseItems = append(responseItems, dto.TeamMemberOptionItem{
			ID:              item.UserID.String(),
			Name:            item.FullName,
			CurrentTeamID:   currentTeamIDValue,
			CurrentTeamName: item.CurrentTeamName,
			IsInCurrentTeam: item.IsInCurrentTeam,
			IsTeamLeader:    item.IsTeamLeader,
		})
	}

	return dto.TeamMemberOptionsResponse{Members: responseItems}, nil
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
	resident, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if resident == nil {
		return fmt.Errorf("user not found")
	}
	if resident.TeamID() != nil && *resident.TeamID() != teamID {
		sourceTeam, err := s.teamRepo.FindByID(ctx, *resident.TeamID())
		if err != nil {
			return err
		}
		if sourceTeam != nil && sourceTeam.LeaderID() == userID {
			return fmt.Errorf("assign a replacement team leader first")
		}
	}
	return s.userRepo.MoveUserToTeam(ctx, userID, &teamID)
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

func appendUniqueOptionalUUID(ids []uuid.UUID, optionalID *uuid.UUID) []uuid.UUID {
	if optionalID == nil {
		return ids
	}

	for _, id := range ids {
		if id == *optionalID {
			return ids
		}
	}

	return append(ids, *optionalID)
}

func optionalUUIDString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}

	stringValue := value.String()
	return &stringValue
}

func (s *Service) nextTeamOrder(ctx context.Context, groupID uuid.UUID) (int, error) {
	teams, err := s.teamRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return 0, err
	}

	maxOrder := 0
	for _, team := range teams {
		if team.RotationPosition() > maxOrder {
			maxOrder = team.RotationPosition()
		}
	}

	return maxOrder + 1, nil
}

func parseRequiredStringUUID(rawID string) (uuid.UUID, error) {
	trimmed := strings.TrimSpace(rawID)
	if trimmed == "" {
		return uuid.Nil, fmt.Errorf("глава команды обязателен")
	}

	id, err := uuid.Parse(trimmed)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user id")
	}

	return id, nil
}

func appendUniqueUUID(ids []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for _, currentID := range ids {
		if currentID == id {
			return ids
		}
	}
	return append(ids, id)
}

func userNames(users []*user.User) map[uuid.UUID]string {
	names := make(map[uuid.UUID]string, len(users))
	for _, resident := range users {
		names[resident.ID()] = strings.TrimSpace(fmt.Sprintf("%s %s", resident.LastName(), resident.FirstName()))
	}
	return names
}

func userSummaryFromMap(names map[uuid.UUID]string, leaderID uuid.UUID) *dto.UserSummary {
	name, ok := names[leaderID]
	if !ok || strings.TrimSpace(name) == "" {
		return nil
	}

	return &dto.UserSummary{
		ID:   leaderID.String(),
		Name: name,
	}
}

func validateResidentTeamName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("Введите название")
	}
	if utf8.RuneCountInString(strings.TrimSpace(name)) > 255 {
		return fmt.Errorf("Название не должно превышать 255 символов")
	}
	return nil
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
	teamLeaders := make(map[uuid.UUID]uuid.UUID, len(teams))
	for _, t := range teams {
		teamNames[t.ID()] = userSummaryFromMap(userNames(users), t.LeaderID()).Name
		teamLeaders[t.ID()] = t.LeaderID()
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
			IsTeamLeader:    u.TeamID() != nil && teamLeaders[*u.TeamID()] == u.ID(),
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
