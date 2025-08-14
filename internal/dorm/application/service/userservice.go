package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure"
	"dorm/internal/dorm/infrastructure/mysql/repository"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type UserService struct {
	userRepository *repository.UserRepository
	telegram       *infrastructure.Telegram
}

func NewUserService(userRepository *repository.UserRepository, telegram *infrastructure.Telegram) *UserService {
	return &UserService{
		userRepository: userRepository,
		telegram:       telegram,
	}
}

func (s *UserService) RegisterUser(fullName string, chatId int64) error {
	user, err := s.userRepository.Find(chatId)
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
		UserId:     uuid.New(),
		TelegramId: chatId,
		FirstName:  parts[0],
		LastName:   parts[1],
	}
	if len(parts) == 3 {
		user.MiddleName = parts[2]
	}

	return s.userRepository.Store(user)
}

func (s *UserService) GetUser(id int64) (*model.User, error) {
	return s.userRepository.Find(id)
}
