package service

import (
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"
	"errors"
	"fmt"
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
	rows, err := s.userRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var users []model.User

	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.UserID,
			&user.TelegramID,
			&user.FirstName,
			&user.LastName,
			&user.MiddleName,
			&user.TeamID,
			&user.RoomNumber,
			&user.DormitoryID,
			&user.RoleID,
			&user.CreatedAt,
			&user.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan users row: %w", err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}
