package service

import (
	"dorm/internal/common/consts"
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(repository *repository.UserRepository) *UserService {
	return &UserService{
		userRepository: repository,
	}
}

func (s *UserService) RegisterUser(fullName string, chatID int64) error {
	user, err := s.userRepository.FindByTelegramID(chatID)
	if err != nil {
		return err
	}
	if user != nil {
		return nil
	}
	parts := strings.Split(fullName, " ")

	if len(parts) < 2 || len(parts) > 3 {
		return errors.New("invalid full name")
	}
	user = &model.User{
		UserID:     uuid.New(),
		TelegramID: chatID,
		FirstName:  parts[0],
		LastName:   parts[1],
		RoleID:     1,
	}
	if len(parts) == 3 {
		middleName := parts[2]
		user.MiddleName = &middleName
	}

	return s.userRepository.Store(user)
}

func (s *UserService) GetUserByTelegramID(telegramID int64) (*model.User, error) {
	return s.userRepository.FindByTelegramID(telegramID)
}

func (s *UserService) GetUserByID(id uuid.UUID) (*model.User, error) {
	return s.userRepository.FindByID(id)
}

func (s *UserService) GetUsersByTeamID(teamID uuid.UUID) ([]model.User, error) {
	return s.userRepository.FindUsersByTeamID(teamID)
}

func (s *UserService) GetRequiredUserPoints(userID uuid.UUID, dutyPoints int) (float64, error) {
	teamID, err := s.userRepository.FindUserTeamID(userID)
	if err != nil {
		return 0, err
	}
	if teamID == nil {
		return 0, nil
	}
	teamUsers, err := s.userRepository.FindUsersByTeamID(*teamID)
	if err != nil {
		return 0, err
	}
	teamSize := len(teamUsers)
	if teamSize == 0 {
		return 0, nil
	}
	return float64(dutyPoints) / float64(teamSize), nil
}

func (s *UserService) GetDormitoryHeads(dormID int64) ([]model.User, error) {
	roles := []int{consts.RoleFloorHead, consts.RoleColivingHead}
	headsRaw, err := s.userRepository.FindUsersByRole(roles)
	if err != nil {
		return nil, err
	}
	heads := make([]model.User, 0)
	for _, head := range headsRaw {
		if head.DormitoryID != nil && *head.DormitoryID == dormID {
			heads = append(heads, head)
		}
	}
	return heads, nil
}
