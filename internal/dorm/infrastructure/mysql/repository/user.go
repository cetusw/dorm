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

func NewUserRepository(dbUser, dbPass, dbHost, dbPort, dbName string) (*UserRepository, error) {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &UserRepository{db: db}, nil
}

func (r *UserRepository) Close() {
	r.db.Close()
}

func (r *UserRepository) Store(user *model.User) error {
	query := "INSERT INTO user (user_id, telegram_id, first_name, last_name) VALUES (UUID_TO_BIN(?), ?, ?, ?)"
	_, err := r.db.Exec(query, user.UserId, user.TelegramId, user.FirstName, user.LastName)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

func (r *UserRepository) Find(telegramId int64) (*model.User, error) {
	user := &model.User{}
	query := "SELECT user_id, telegram_id, first_name, last_name, middle_name, team_id, room_number, dormitory_id, created_at, deleted_at FROM user WHERE telegram_id = ?"

	err := r.db.QueryRow(query, telegramId).Scan(
		&user.UserId,
		&user.TelegramId,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.TeamId,
		&user.RoomNumber,
		&user.DormitoryId,
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
