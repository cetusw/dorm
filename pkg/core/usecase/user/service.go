package user

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
)

type Service struct {
	userRepo      user.Repository
	teamRepo      structure.TeamRepository
	groupRepo     structure.GroupRepository
	dormitoryRepo structure.DormitoryRepository
}

func NewUserService(
	userRepo user.Repository,
	teamRepo structure.TeamRepository,
	groupRepo structure.GroupRepository,
	dormitoryRepo structure.DormitoryRepository,
) *Service {
	return &Service{
		userRepo:      userRepo,
		teamRepo:      teamRepo,
		groupRepo:     groupRepo,
		dormitoryRepo: dormitoryRepo,
	}
}

func (s *Service) RegisterUser(ctx context.Context, telegramID int64, fullName string) error {
	parts := strings.Split(fullName, " ")
	u, err := user.NewUser(telegramID, parts[0], parts[1])
	if err != nil {
		return err
	}
	return s.userRepo.Save(ctx, u)
}

func (s *Service) GetUserByTelegramID(ctx context.Context, telegramID int64) (*user.User, error) {
	return s.userRepo.FindByTelegramID(ctx, telegramID)
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *Service) GetUserProfile(ctx context.Context, userID uuid.UUID) (*ports.ProfileViewModel, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}

	room := "нет"
	if u.RoomNumber() != nil {
		room = *u.RoomNumber()
	}

	profile := &ports.ProfileViewModel{
		FirstName:     u.FirstName(),
		LastName:      u.LastName(),
		RoomNumber:    room,
		DormitoryName: "нет",
		GroupName:     "нет",
		TeamName:      "нет",
	}

	team, err := s.fillTeamData(ctx, profile, u)
	if err != nil || team == nil {
		return profile, err
	}

	group, err := s.fillGroupData(ctx, profile, team)
	if err != nil || group == nil {
		return profile, err
	}

	if err := s.fillDormData(ctx, profile, group); err != nil {
		return profile, err
	}

	return profile, nil
}

func (s *Service) fillTeamData(ctx context.Context, p *ports.ProfileViewModel, u *user.User) (*structure.Team, error) {
	if u.TeamID() == nil {
		return nil, nil
	}
	team, err := s.teamRepo.FindByID(ctx, *u.TeamID())
	if err != nil || team == nil {
		return nil, err
	}
	p.TeamID = team.ID()
	p.TeamName = team.Name()
	return team, nil
}

func (s *Service) fillGroupData(ctx context.Context, p *ports.ProfileViewModel, t *structure.Team) (*structure.Group, error) {
	group, err := s.groupRepo.FindByID(ctx, t.GroupID())
	if err != nil || group == nil {
		return nil, err
	}
	p.GroupName = group.Name()
	return group, nil
}

func (s *Service) fillDormData(ctx context.Context, p *ports.ProfileViewModel, g *structure.Group) error {
	dorm, err := s.dormitoryRepo.FindByID(ctx, g.DormitoryID())
	if err != nil || dorm == nil {
		return err
	}
	p.DormitoryName = dorm.Name()
	return nil
}
