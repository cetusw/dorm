package user

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"dorm/pkg/core/ports/dto"
	ports "dorm/pkg/core/ports/query"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
)

type Service struct {
	userRepo      user.Repository
	teamRepo      structure.TeamRepository
	groupRepo     structure.GroupRepository
	dormitoryRepo structure.DormitoryRepository
	queryService  ports.UserQueryService
}

func NewUserService(
	userRepo user.Repository,
	teamRepo structure.TeamRepository,
	groupRepo structure.GroupRepository,
	dormitoryRepo structure.DormitoryRepository,
	queryService ports.UserQueryService,
) *Service {
	return &Service{
		userRepo:      userRepo,
		teamRepo:      teamRepo,
		groupRepo:     groupRepo,
		dormitoryRepo: dormitoryRepo,
		queryService:  queryService,
	}
}

func (s *Service) CreateUser(ctx context.Context, request dto.CreateUserRequest) error {
	if request.DormitoryID == 0 {
		return fmt.Errorf("dormitory is required")
	}
	passwordHash, err := hashRequiredResidentPassword(request.Password)
	if err != nil {
		return err
	}

	u, err := user.NewManualUser(request.FirstName, request.LastName, request.Login, passwordHash)
	if err != nil {
		return err
	}

	u.SetMiddleName(request.MiddleName)
	u.MoveInto(request.DormitoryID, request.RoomNumber)
	u.SetFloorNumber(request.Floor)
	if request.TeamID != "" {
		teamID, err := uuid.Parse(request.TeamID)
		if err != nil {
			return fmt.Errorf("invalid team_id")
		}
		u.JoinTeam(teamID)
	}

	return s.userRepo.Save(ctx, u)
}

func (s *Service) AuthenticateResident(ctx context.Context, login string, password string) (*user.User, error) {
	login = strings.TrimSpace(login)
	password = strings.TrimSpace(password)
	if login == "" || password == "" {
		return nil, fmt.Errorf("логин и пароль обязательны")
	}

	u, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("find user by login: %w", err)
	}
	if u == nil {
		return nil, fmt.Errorf("неверный логин или пароль")
	}

	passwordHash := hashResidentPassword(password)
	if subtle.ConstantTimeCompare([]byte(u.PasswordHash()), []byte(passwordHash)) != 1 {
		return nil, fmt.Errorf("неверный логин или пароль")
	}

	return u, nil
}

func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, request dto.UpdateUserRequest) error {
	if request.DormitoryID == 0 {
		return fmt.Errorf("dormitory is required")
	}
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("user not found")
	}

	if err := u.Rename(request.FirstName, request.LastName); err != nil {
		return err
	}
	if err := u.SetLogin(request.Login); err != nil {
		return err
	}
	if strings.TrimSpace(request.Password) != "" {
		passwordHash, err := hashRequiredResidentPassword(request.Password)
		if err != nil {
			return err
		}
		if err := u.SetPasswordHash(passwordHash); err != nil {
			return err
		}
	}
	u.SetMiddleName(request.MiddleName)
	u.MoveInto(request.DormitoryID, request.RoomNumber)
	u.SetFloorNumber(request.Floor)

	return s.userRepo.Save(ctx, u)
}

func (s *Service) SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.SoftDelete(ctx, id)
}

func (s *Service) RegisterUser(ctx context.Context, telegramID int64, fullName string) error {
	parts := strings.Split(fullName, " ")
	u, err := user.NewUser(
		telegramID,
		parts[0],
		parts[1],
		fmt.Sprintf("tg_%d", telegramID),
		telegramPlaceholderPasswordHash(telegramID),
	)
	if err != nil {
		return err
	}
	return s.userRepo.Save(ctx, u)
}

func telegramPlaceholderPasswordHash(telegramID int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("telegram:%d", telegramID)))
	return hex.EncodeToString(sum[:])
}

func hashResidentPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func hashRequiredResidentPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", fmt.Errorf("пароль обязателен")
	}

	return hashResidentPassword(password), nil
}

func (s *Service) GetUserByTelegramID(ctx context.Context, telegramID int64) (*user.User, error) {
	return s.userRepo.FindByTelegramID(ctx, telegramID)
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *Service) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.CurrentUserResponse, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find current user: %w", err)
	}
	if u == nil {
		return nil, nil
	}

	canManageDormitories, err := s.dormitoryRepo.ExistsByLeaderID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check dormitory management access: %w", err)
	}

	return &dto.CurrentUserResponse{
		ID:                   u.ID().String(),
		FirstName:            u.FirstName(),
		LastName:             u.LastName(),
		CanManageDormitories: canManageDormitories,
	}, nil
}

func (s *Service) GetUserProfile(ctx context.Context, userID uuid.UUID) (*dto.ProfileViewModel, error) {
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

	profile := &dto.ProfileViewModel{
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

func (s *Service) GetUsersList(ctx context.Context) ([]dto.UserListItem, error) {
	return s.queryService.GetUsersDetailedList(ctx)
}

func (s *Service) GetResidentsResponse(ctx context.Context, dormitoryID int64) (dto.ResidentListResponse, error) {
	items, err := s.queryService.GetResidentsByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return dto.ResidentListResponse{}, fmt.Errorf("load residents by dormitory: %w", err)
	}

	return dto.ResidentListResponse{
		Users: items,
	}, nil
}

func (s *Service) fillTeamData(ctx context.Context, p *dto.ProfileViewModel, u *user.User) (*structure.Team, error) {
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

func (s *Service) fillGroupData(ctx context.Context, p *dto.ProfileViewModel, t *structure.Team) (*structure.Group, error) {
	group, err := s.groupRepo.FindByID(ctx, t.GroupID())
	if err != nil || group == nil {
		return nil, err
	}
	p.GroupName = group.Name()
	return group, nil
}

func (s *Service) fillDormData(ctx context.Context, p *dto.ProfileViewModel, g *structure.Group) error {
	dorm, err := s.dormitoryRepo.FindByID(ctx, g.DormitoryID())
	if err != nil || dorm == nil {
		return err
	}
	p.DormitoryName = dorm.Name()
	return nil
}
