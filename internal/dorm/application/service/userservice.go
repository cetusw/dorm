package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(userRepository *repository.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) RegisterUser(fullName string, chatID int64) error {
	user, err := s.userRepository.Find(chatID)
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

func (s *UserService) GetUser(id int64) (*model.User, error) {
	return s.userRepository.Find(id)
}

func (s *UserService) GetAllUsers() ([]model.User, error) {
	return s.userRepository.FindAll()
}
