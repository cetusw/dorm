package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByTelegramID(telegramID int64) (*model.User, error) {
	const sqlQuery = `
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
		WHERE telegram_id = ?
		`
	user := &model.User{}
	err := r.db.QueryRow(sqlQuery, telegramID).Scan(
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

func (r *UserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	const sqlQuery = `
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
		WHERE user_id = UUID_TO_BIN(?)`

	user := &model.User{}
	err := r.db.QueryRow(sqlQuery, id).Scan(
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
	const sqlQuery = `
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

	rows, err := r.db.Query(sqlQuery)
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

func (r *UserRepository) FindUsersByTeamID(teamID uuid.UUID) ([]model.User, error) {
	const sqlQuery = `
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
		WHERE team_id = UUID_TO_BIN(?)`

	rows, err := r.db.Query(sqlQuery, teamID)
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

func (r *UserRepository) FindUserTeamID(userID uuid.UUID) (*uuid.UUID, error) {
	const sqlQuery = `
		SELECT team_id 
		FROM user 
		WHERE user_id = UUID_TO_BIN(?)`

	var teamID *uuid.UUID
	err := r.db.QueryRow(sqlQuery, userID).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get user team_id: %w", err)
	}

	if teamID == nil {
		return nil, nil
	}

	return teamID, nil
}

func (r *UserRepository) FindUsersByRole(roleIDs []int) ([]model.User, error) {
	if len(roleIDs) == 0 {
		return []model.User{}, nil
	}

	placeholders := make([]string, len(roleIDs))
	args := make([]interface{}, len(roleIDs))

	for i, id := range roleIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	sqlQuery := fmt.Sprintf(`
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
		WHERE role_id IN (%s) AND deleted_at IS NULL`,
		strings.Join(placeholders, ","))

	rows, err := r.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users by roles: %w", err)
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

func (r *UserRepository) Store(user *model.User) error {
	const sqlQuery = `
		INSERT INTO user (user_id, telegram_id, first_name, last_name, role_id) 
		VALUES (UUID_TO_BIN(?), ?, ?, ?, ?)`

	_, err := r.db.Exec(sqlQuery, user.UserID, user.TelegramID, user.FirstName, user.LastName, user.RoleID)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}
