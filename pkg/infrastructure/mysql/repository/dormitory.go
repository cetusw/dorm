package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/structure"
)

type DormitoryRepository struct {
	db *sql.DB
}

func NewDormitoryRepository(db *sql.DB) *DormitoryRepository {
	return &DormitoryRepository{db: db}
}

func (r *DormitoryRepository) FindByID(ctx context.Context, id int64) (*structure.Dormitory, error) {
	const query = `SELECT id, name, leader_id, city, street_type, street_name, house_number FROM dormitory WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	var name, city, streetType, streetName, houseNumber string
	var leaderIDBytes []byte

	if err := row.Scan(&id, &name, &leaderIDBytes, &city, &streetType, &streetName, &houseNumber); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindDormitoryByID: %w", err)
	}

	var leaderID *uuid.UUID
	if len(leaderIDBytes) > 0 {
		uid, _ := uuid.FromBytes(leaderIDBytes)
		leaderID = &uid
	}

	return structure.RestoreDormitory(id, name, leaderID, city, streetType, streetName, houseNumber), nil
}

func (r *DormitoryRepository) FindAll(ctx context.Context) ([]*structure.Dormitory, error) {
	const query = `SELECT id, name, leader_id, city, street_type, street_name, house_number FROM dormitory`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("FindAllDormitories: %w", err)
	}
	defer rows.Close()

	var dormitories []*structure.Dormitory
	for rows.Next() {
		var id int64
		var name, city, streetType, streetName, houseNumber string
		var leaderIDBytes []byte

		if err := rows.Scan(&id, &name, &leaderIDBytes, &city, &streetType, &streetName, &houseNumber); err != nil {
			return nil, fmt.Errorf("FindAllDormitories: scan: %w", err)
		}

		var leaderID *uuid.UUID
		if len(leaderIDBytes) > 0 {
			uid, _ := uuid.FromBytes(leaderIDBytes)
			leaderID = &uid
		}

		dormitories = append(dormitories, structure.RestoreDormitory(id, name, leaderID, city, streetType, streetName, houseNumber))
	}

	return dormitories, nil
}

func (r *DormitoryRepository) Save(ctx context.Context, dormitory *structure.Dormitory) error {
	if dormitory.ID() == 0 {
		const query = `
			INSERT INTO dormitory (leader_id, name, city, street_type, street_name, house_number)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		_, err := r.db.ExecContext(ctx, query, leaderIDBytes(dormitory.LeaderID()), dormitory.Name(), dormitory.City(), dormitory.StreetType(), dormitory.StreetName(), dormitory.HouseNumber())
		if err != nil {
			return fmt.Errorf("SaveDormitory insert: %w", err)
		}
		return nil
	}

	const query = `
		UPDATE dormitory
		SET leader_id = ?, name = ?, city = ?, street_type = ?, street_name = ?, house_number = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, leaderIDBytes(dormitory.LeaderID()), dormitory.Name(), dormitory.City(), dormitory.StreetType(), dormitory.StreetName(), dormitory.HouseNumber(), dormitory.ID())
	if err != nil {
		return fmt.Errorf("SaveDormitory update: %w", err)
	}
	return nil
}

func leaderIDBytes(id *uuid.UUID) interface{} {
	if id == nil {
		return nil
	}
	value, _ := id.MarshalBinary()
	return value
}

func (r *DormitoryRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM dormitory WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("DeleteDormitory: %w", err)
	}
	return nil
}
