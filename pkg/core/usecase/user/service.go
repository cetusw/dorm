package user

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/user"
)

type Service struct {
	repo user.Repository
}

func NewUserService(repo user.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterUser(ctx context.Context, telegramID int64, fullName string) error {
	parts := strings.Split(fullName, " ")
	u, err := user.NewUser(telegramID, parts[0], parts[1])
	if err != nil {
		return err
	}
	return s.repo.Save(ctx, u)
}

func (s *Service) GetUserByTelegramID(ctx context.Context, telegramID int64) (*user.User, error) {
	return s.repo.FindByTelegramID(ctx, telegramID)
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.repo.FindByID(ctx, id)
}
