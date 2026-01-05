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
	const query = `SELECT id, name, leader_id, city FROM dormitory WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	var name, city string
	var leaderIDBytes []byte

	if err := row.Scan(&id, &name, &leaderIDBytes, &city); err != nil {
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

	return structure.RestoreDormitory(id, name, leaderID, city), nil
}

func (r *DormitoryRepository) FindAll(ctx context.Context) ([]*structure.Dormitory, error) {
	const query = `SELECT id, name, leader_id, city FROM dormitory`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("FindAllDormitories: %w", err)
	}
	defer rows.Close()

	var dormitories []*structure.Dormitory
	for rows.Next() {
		var id int64
		var name, city string
		var leaderIDBytes []byte

		if err := rows.Scan(&id, &name, &leaderIDBytes, &city); err != nil {
			return nil, fmt.Errorf("FindAllDormitories: scan: %w", err)
		}

		var leaderID *uuid.UUID
		if len(leaderIDBytes) > 0 {
			uid, _ := uuid.FromBytes(leaderIDBytes)
			leaderID = &uid
		}

		dormitories = append(dormitories, structure.RestoreDormitory(id, name, leaderID, city))
	}

	return dormitories, nil
}
