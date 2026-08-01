package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	domain "dorm/pkg/core/domain/notification"
	"dorm/pkg/core/ports"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type NotificationRepository struct {
	db *sql.DB
}

var _ ports.NotificationRepository = (*NotificationRepository)(nil)

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) CreateIfAbsent(ctx context.Context, notification *domain.Notification) (bool, error) {
	const query = `
		INSERT INTO notification (
			id, user_id, type, title, body, target_url, deduplication_key, created_at, read_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	idBytes, err := marshalBinaryUUID(notification.ID(), "notification id")
	if err != nil {
		return false, err
	}
	userIDBytes, err := marshalBinaryUUID(notification.UserID(), "notification user id")
	if err != nil {
		return false, err
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		idBytes,
		userIDBytes,
		string(notification.Type()),
		notification.Title(),
		notification.Body(),
		notification.TargetURL(),
		notification.DeduplicationKey(),
		notification.CreatedAt(),
		nullTime(notification.ReadAt()),
	)
	if err != nil {
		if isNotificationDeduplicationConflict(err) {
			return false, nil
		}
		return false, fmt.Errorf("create notification: %w", err)
	}

	return true, nil
}

func isNotificationDeduplicationConflict(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) {
		return false
	}

	return mysqlErr.Number == 1062 &&
		strings.Contains(strings.ToLower(mysqlErr.Message), "uq_notification_user_deduplication")
}
