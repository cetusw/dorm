package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
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
	query := `
		SELECT 
		    user_id, 
		    telegram_id, 
		    first_name, 
		    last_name, 
		    middle_name, 
		    team_id, 
		    room_number, 
		    dormitory_id, 
		    role_id,
		    created_at,
		    deleted_at 
		FROM user 
		WHERE telegram_id = ?`

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

func (r *UserRepository) FindAll() ([]model.User, error) {
	query := `
		SELECT 
			user_id, 
			telegram_id, 
			first_name, 
			last_name, 
			middle_name, 
			team_id, 
			room_number, 
			dormitory_id, 
			role_id, 
			created_at, 
			deleted_at 
		FROM user`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all users: %w", err)
	}
	defer rows.Close()

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
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

func (r *UserRepository) FindUsersByTeamID(teamID int) ([]model.User, error) {
	query := `
		SELECT 
			user_id, 
			telegram_id, 
			first_name, 
			last_name, 
			middle_name, 
			team_id, 
			room_number, 
			dormitory_id, 
			role_id, 
			created_at, 
			deleted_at 
		FROM user 
		WHERE team_id = ?`

	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query users by team id: %w", err)
	}
	defer rows.Close()

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
			return nil, fmt.Errorf("failed to scan user by team id row: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user by team id rows: %w", err)
	}

	return users, nil
}

func (r *UserRepository) FindUserTeamID(userID uuid.UUID) (int, error) {
	query := `
		SELECT 
			team_id 
		FROM user 
		WHERE user_id = UUID_TO_BIN(?)`

	var teamID *int
	err := r.db.QueryRow(query, userID).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, sql.ErrNoRows
		}
		return 0, fmt.Errorf("failed to get user team_id: %w", err)
	}

	if teamID == nil {
		return 0, nil
	}

	return *teamID, nil
}
