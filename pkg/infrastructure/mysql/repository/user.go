package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/user"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

type userDTO struct {
	ID          []byte
	TelegramID  int64
	FirstName   string
	LastName    string
	TeamID      []byte
	DormitoryID sql.NullInt64
	CreatedAt   time.Time
}

func (dto *userDTO) toDomain() *user.User {
	id, _ := uuid.FromBytes(dto.ID)

	var teamID *uuid.UUID
	if len(dto.TeamID) > 0 {
		uid, _ := uuid.FromBytes(dto.TeamID)
		teamID = &uid
	}

	var dormID *int64
	if dto.DormitoryID.Valid {
		val := dto.DormitoryID.Int64
		dormID = &val
	}

	return user.RestoreUser(
		id,
		dto.TelegramID,
		dto.FirstName,
		dto.LastName,
		teamID,
		dormID,
		dto.CreatedAt,
	)
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	const query = `
		INSERT INTO user (id, telegram_id, first_name, last_name, team_id, dormitory_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			first_name = VALUES(first_name),
			last_name = VALUES(last_name),
			team_id = VALUES(team_id),
			dormitory_id = VALUES(dormitory_id)
	`

	var teamID interface{}
	if u.TeamID() != nil {
		teamID, _ = u.TeamID().MarshalBinary()
	}

	var dormID interface{}
	if u.DormitoryID() != nil {
		dormID = *u.DormitoryID()
	}

	idBytes, _ := u.ID().MarshalBinary()

	_, err := r.db.ExecContext(ctx, query,
		idBytes,
		u.TelegramID(),
		u.FirstName(),
		u.LastName(),
		teamID,
		dormID,
		u.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("userRepo.Save: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	const query = `
		SELECT id, telegram_id, first_name, last_name, team_id, dormitory_id, created_at
		FROM user WHERE id = ?
	`
	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	return r.scanUser(row)
}

func (r *UserRepository) FindByTelegramID(ctx context.Context, telegramID int64) (*user.User, error) {
	const query = `
		SELECT id, telegram_id, first_name, last_name, team_id, dormitory_id, created_at
		FROM user WHERE telegram_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, telegramID)
	return r.scanUser(row)
}

func (r *UserRepository) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]*user.User, error) {
	const query = `
		SELECT id, telegram_id, first_name, last_name, team_id, dormitory_id, created_at
		FROM user WHERE team_id = ?
	`
	idBytes, _ := teamID.MarshalBinary()
	rows, err := r.db.QueryContext(ctx, query, idBytes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var dto userDTO
		err := rows.Scan(
			&dto.ID, &dto.TelegramID, &dto.FirstName, &dto.LastName,
			&dto.TeamID, &dto.DormitoryID, &dto.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, dto.toDomain())
	}
	return users, nil
}

func (r *UserRepository) scanUser(row *sql.Row) (*user.User, error) {
	var dto userDTO
	err := row.Scan(
		&dto.ID, &dto.TelegramID, &dto.FirstName, &dto.LastName,
		&dto.TeamID, &dto.DormitoryID, &dto.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return dto.toDomain(), nil
}
