package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Store(user *model.User) error {
	query := "INSERT INTO user (user_id, telegram_id, first_name, last_name, role_id) VALUES (UUID_TO_BIN(?), ?, ?, ?, ?)"
	_, err := r.db.Exec(query, user.UserID, user.TelegramID, user.FirstName, user.LastName, user.RoleID)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

func (r *UserRepository) Find(telegramID int64) (*model.User, error) {
	user := &model.User{}
	query := "SELECT user_id, telegram_id, first_name, last_name, middle_name, team_id, room_number, dormitory_id, role_id, deleted_at FROM user WHERE telegram_id = ?"

	err := r.db.QueryRow(query, telegramID).Scan(
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
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}
