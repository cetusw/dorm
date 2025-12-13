package repository

import (
	"database/sql"
	"dorm/internal/dorm/application/model"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type GroupRepository struct {
	db *sql.DB
}

func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Find(groupID uuid.UUID) (*model.Group, error) {
	const sqlQuery = "SELECT group_id, leader_id, group_name, dormitory_id, spreadsheet_id FROM `group` WHERE group_id = UUID_TO_BIN(?)"

	group := &model.Group{}
	err := r.db.QueryRow(sqlQuery, groupID).Scan(
		&group.GroupID,
		&group.LeaderID,
		&group.Name,
		&group.DormitoryID,
		&group.SpreadsheetID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	return group, nil
}

func (r *GroupRepository) FindAll() ([]model.Group, error) {
	const sqlQuery = "SELECT group_id, leader_id, group_name, dormitory_id, spreadsheet_id FROM `group`"

	rows, err := r.db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query all team groups: %w", err)
	}
	defer rows.Close()

	var groups []model.Group

	for rows.Next() {
		var group model.Group
		if err := rows.Scan(
			&group.GroupID,
			&group.LeaderID,
			&group.Name,
			&group.DormitoryID,
			&group.SpreadsheetID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan team group row: %w", err)
		}
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team group rows: %w", err)
	}

	return groups, nil
}
