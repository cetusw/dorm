package repository

import (
	"database/sql"

	"dorm/internal/dorm/application/model"
)

type DormitoryRepository struct {
	db *sql.DB
}

func NewDormitoryRepository(db *sql.DB) *DormitoryRepository {
	return &DormitoryRepository{db: db}
}

func (r *DormitoryRepository) FindAll() ([]model.Dormitory, error) {
	const sqlQuery = `
		SELECT 
		    dormitory_id, 
		    leader_id,
		    dormitory_name, 
		    dormitory_city, 
		    dormitory_street_type, 
		    dormitory_street_name, 
		    dormitory_house_number
		FROM dormitory
		`
	rows, err := r.db.Query(sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dorms []model.Dormitory
	for rows.Next() {
		var d model.Dormitory
		if err := rows.Scan(
			&d.DormitoryID,
			&d.LeaderID,
			&d.Name,
			&d.City,
			&d.StreetType,
			&d.StreetName,
			&d.HouseNumber,
		); err != nil {
			return nil, err
		}
		dorms = append(dorms, d)
	}
	return dorms, nil
}
