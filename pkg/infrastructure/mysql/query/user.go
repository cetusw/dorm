package query

import (
	"context"
	"database/sql"
	"dorm/pkg/core/ports/dto"
	"fmt"

	"github.com/google/uuid"
)

type UserQueryService struct {
	db *sql.DB
}

func NewUserQueryService(db *sql.DB) *UserQueryService {
	return &UserQueryService{db: db}
}

func (q *UserQueryService) GetResidentsByDormitoryID(ctx context.Context, dormitoryID int64) ([]dto.ResidentListItem, error) {
	const query = `
		SELECT
			u.id,
			TRIM(CONCAT(
				u.last_name, ' ',
				u.first_name,
				CASE
					WHEN u.middle_name IS NULL OR u.middle_name = '' THEN ''
					ELSE CONCAT(' ', u.middle_name)
				END
			)),
			COALESCE(u.room_number, 'Не указана')
		FROM user u
		WHERE u.deleted_at IS NULL
		  AND u.dormitory_id = ?
		ORDER BY u.last_name ASC, u.first_name ASC, u.middle_name ASC
	`

	rows, err := q.db.QueryContext(ctx, query, dormitoryID)
	if err != nil {
		return nil, fmt.Errorf("query residents by dormitory: %w", err)
	}
	defer rows.Close()

	items := make([]dto.ResidentListItem, 0)
	for rows.Next() {
		var item dto.ResidentListItem
		var idBytes []byte

		if err := rows.Scan(&idBytes, &item.Name, &item.RoomNumber); err != nil {
			return nil, fmt.Errorf("scan resident by dormitory: %w", err)
		}

		userID, err := uuid.FromBytes(idBytes)
		if err != nil {
			return nil, fmt.Errorf("decode resident id: %w", err)
		}
		item.ID = userID.String()
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate residents by dormitory: %w", err)
	}

	return items, nil
}
