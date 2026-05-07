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
	TelegramID  sql.NullInt64
	FirstName   string
	LastName    string
	TeamID      []byte
	RoomNumber  sql.NullString
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

	var roomNumber *string
	if dto.RoomNumber.Valid {
		room := dto.RoomNumber.String
		roomNumber = &room
	}

	return user.RestoreUser(
		id,
		nullInt64Ptr(dto.TelegramID),
		dto.FirstName,
		dto.LastName,
		teamID,
		roomNumber,
		dormID,
		dto.CreatedAt,
	)
}

func nullInt64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	const query = `
		INSERT INTO user (id, telegram_id, first_name, last_name, team_id, room_number, dormitory_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			first_name = VALUES(first_name),
			last_name = VALUES(last_name),
			team_id = VALUES(team_id),
			room_number = VALUES(room_number),
			dormitory_id = VALUES(dormitory_id)
	`

	var teamID interface{} = nil
	if u.TeamID() != nil {
		teamID, _ = u.TeamID().MarshalBinary()
	}

	var roomNumber interface{} = nil
	if u.RoomNumber() != nil {
		roomNumber = *u.RoomNumber()
	}

	var dormID interface{} = nil
	if u.DormitoryID() != nil {
		dormID = *u.DormitoryID()
	}

	idBytes, _ := u.ID().MarshalBinary()

	_, err := r.db.ExecContext(ctx, query,
		idBytes,
		u.TelegramIDValue(),
		u.FirstName(),
		u.LastName(),
		teamID,
		roomNumber,
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
		SELECT id, telegram_id, first_name, last_name, team_id, room_number, dormitory_id, created_at
		FROM user WHERE id = ? AND deleted_at IS NULL
	`
	idBytes, _ := id.MarshalBinary()
	row := r.db.QueryRowContext(ctx, query, idBytes)

	return r.scanUser(row)
}

func (r *UserRepository) FindByTelegramID(ctx context.Context, telegramID int64) (*user.User, error) {
	const query = `
		SELECT id, telegram_id, first_name, last_name, team_id, room_number, dormitory_id, created_at
		FROM user WHERE telegram_id = ? AND deleted_at IS NULL
	`
	row := r.db.QueryRowContext(ctx, query, telegramID)
	return r.scanUser(row)
}

func (r *UserRepository) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]*user.User, error) {
	const query = `
		SELECT id, telegram_id, first_name, last_name, team_id, room_number, dormitory_id, created_at
		FROM user WHERE team_id = ? AND deleted_at IS NULL
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
			&dto.ID,
			&dto.TelegramID,
			&dto.FirstName,
			&dto.LastName,
			&dto.TeamID,
			&dto.RoomNumber,
			&dto.DormitoryID,
			&dto.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, dto.toDomain())
	}
	return users, nil
}

func (r *UserRepository) FindByDormitoryID(ctx context.Context, dormitoryID int64) ([]*user.User, error) {
	const query = `
		SELECT id, telegram_id, first_name, last_name, team_id, room_number, dormitory_id, created_at
		FROM user
		WHERE dormitory_id = ? AND deleted_at IS NULL
		ORDER BY last_name, first_name
	`
	rows, err := r.db.QueryContext(ctx, query, dormitoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var dto userDTO
		err := rows.Scan(
			&dto.ID,
			&dto.TelegramID,
			&dto.FirstName,
			&dto.LastName,
			&dto.TeamID,
			&dto.RoomNumber,
			&dto.DormitoryID,
			&dto.CreatedAt,
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
		&dto.ID,
		&dto.TelegramID,
		&dto.FirstName,
		&dto.LastName,
		&dto.TeamID,
		&dto.RoomNumber,
		&dto.DormitoryID,
		&dto.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return dto.toDomain(), nil
}

func (r *UserRepository) MoveUserToTeam(ctx context.Context, userID uuid.UUID, teamID *uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("userRepo.MoveUserToTeam begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var userIDBytes []byte
	userIDBytes, _ = userID.MarshalBinary()

	var teamIDArg interface{} = nil
	if teamID != nil {
		var teamIDBytes []byte
		teamIDBytes, _ = teamID.MarshalBinary()
		teamIDArg = teamIDBytes
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE user SET team_id = ? WHERE id = ? AND deleted_at IS NULL",
		teamIDArg,
		userIDBytes,
	)
	if err != nil {
		return fmt.Errorf("userRepo.MoveUserToTeam update: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("userRepo.MoveUserToTeam commit: %w", err)
	}
	return nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE user
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`
	idBytes, _ := id.MarshalBinary()
	_, err := r.db.ExecContext(ctx, query, idBytes)
	if err != nil {
		return fmt.Errorf("userRepo.SoftDelete: %w", err)
	}
	return nil
}
