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
			u.id, COALESCE(u.telegram_id, 0), u.first_name, u.last_name, COALESCE(u.middle_name, ''),
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
