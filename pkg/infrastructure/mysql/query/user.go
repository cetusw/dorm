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

func (q *UserQueryService) GetUsersDetailedList(ctx context.Context) ([]dto.UserListItem, error) {
	const query = `
		SELECT 
			u.id, u.login, COALESCE(u.telegram_id, 0), u.first_name, u.last_name, COALESCE(u.middle_name, ''),
			COALESCE(u.room_number, '-'),
			COALESCE(d.name, 'Не назначено'),
			COALESCE(g.name, 'Нет группы'),
			COALESCE(t.name, 'Без команды')
		FROM user u
		LEFT JOIN dormitory d ON u.dormitory_id = d.id
		LEFT JOIN team t ON u.team_id = t.id
		LEFT JOIN ` + "`group`" + ` g ON t.group_id = g.id
		WHERE u.deleted_at IS NULL
		ORDER BY u.last_name ASC
	`

	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.UserListItem
	for rows.Next() {
		var item dto.UserListItem
		var idBytes []byte
		var firstName, lastName, middleName string

		err := rows.Scan(
			&idBytes,
			&item.Login,
			&item.TelegramID,
			&firstName,
			&lastName,
			&middleName,
			&item.RoomNumber,
			&item.DormitoryName,
			&item.GroupName,
			&item.TeamName,
		)
		if err != nil {
			return nil, err
		}

		uid, _ := uuid.FromBytes(idBytes)
		item.ID = uid
		item.FullName = fmt.Sprintf("%s %s %s", lastName, firstName, middleName)

		result = append(result, item)
	}

	return result, rows.Err()
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
