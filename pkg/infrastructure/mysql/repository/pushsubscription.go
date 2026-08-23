package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"dorm/pkg/core/domain/notification"

	"github.com/google/uuid"
)

type PushSubscriptionRepository struct {
	db *sql.DB
}

type pushSubscriptionDTO struct {
	ID           []byte
	UserID       []byte
	Endpoint     string
	EndpointHash []byte
	P256DH       string
	AuthSecret   string
	UserAgent    sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var _ notification.PushSubscriptionRepository = (*PushSubscriptionRepository)(nil)

func NewPushSubscriptionRepository(db *sql.DB) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{db: db}
}

func (r *PushSubscriptionRepository) Save(ctx context.Context, subscription *notification.PushSubscription) error {
	const query = `
		INSERT INTO push_subscription (
			id, user_id, endpoint, endpoint_hash, p256dh, auth_secret, user_agent, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			endpoint = VALUES(endpoint),
			p256dh = VALUES(p256dh),
			auth_secret = VALUES(auth_secret),
			user_agent = VALUES(user_agent),
			updated_at = VALUES(updated_at)
	`

	idBytes, err := marshalBinaryUUID(subscription.ID(), "subscription id")
	if err != nil {
		return err
	}
	userIDBytes, err := marshalBinaryUUID(subscription.UserID(), "user id")
	if err != nil {
		return err
	}
	endpointHash, err := validEndpointHash(subscription.EndpointHash())
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		idBytes,
		userIDBytes,
		subscription.Endpoint(),
		endpointHash,
		subscription.P256DH(),
		subscription.AuthSecret(),
		nullStringValue(subscription.UserAgent()),
		subscription.CreatedAt(),
		subscription.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("save push subscription: %w", err)
	}

	return nil
}

func (r *PushSubscriptionRepository) FindByEndpointHash(ctx context.Context, endpointHash []byte) (*notification.PushSubscription, error) {
	const query = `
		SELECT id, user_id, endpoint, endpoint_hash, p256dh, auth_secret, user_agent, created_at, updated_at
		FROM push_subscription
		WHERE endpoint_hash = ?
	`

	validHash, err := validEndpointHash(endpointHash)
	if err != nil {
		return nil, err
	}

	subscription, err := scanPushSubscription(r.db.QueryRowContext(ctx, query, validHash))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find push subscription by endpoint hash: %w", err)
	}

	return subscription, nil
}

func (r *PushSubscriptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*notification.PushSubscription, error) {
	const query = `
		SELECT id, user_id, endpoint, endpoint_hash, p256dh, auth_secret, user_agent, created_at, updated_at
		FROM push_subscription
		WHERE user_id = ?
		ORDER BY created_at ASC, id ASC
	`

	userIDBytes, err := marshalBinaryUUID(userID, "user id")
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, userIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find push subscriptions by user id: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]*notification.PushSubscription, 0)
	for rows.Next() {
		subscription, err := scanPushSubscription(rows)
		if err != nil {
			return nil, fmt.Errorf("find push subscriptions by user id: %w", err)
		}
		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("find push subscriptions by user id: %w", err)
	}

	return subscriptions, nil
}

func (r *PushSubscriptionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM push_subscription WHERE id = ?`

	idBytes, err := marshalBinaryUUID(id, "subscription id")
	if err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, query, idBytes); err != nil {
		return fmt.Errorf("delete push subscription by id: %w", err)
	}

	return nil
}

func (r *PushSubscriptionRepository) DeleteByEndpointHash(ctx context.Context, userID uuid.UUID, endpointHash []byte) error {
	const query = `
		DELETE FROM push_subscription
		WHERE user_id = ? AND endpoint_hash = ?
	`

	userIDBytes, err := marshalBinaryUUID(userID, "user id")
	if err != nil {
		return err
	}
	validHash, err := validEndpointHash(endpointHash)
	if err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, query, userIDBytes, validHash); err != nil {
		return fmt.Errorf("delete push subscription by endpoint hash: %w", err)
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPushSubscription(scanner rowScanner) (*notification.PushSubscription, error) {
	var dto pushSubscriptionDTO
	if err := scanner.Scan(
		&dto.ID,
		&dto.UserID,
		&dto.Endpoint,
		&dto.EndpointHash,
		&dto.P256DH,
		&dto.AuthSecret,
		&dto.UserAgent,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return dto.toDomain()
}

func (dto pushSubscriptionDTO) toDomain() (*notification.PushSubscription, error) {
	id, err := uuid.FromBytes(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("map push subscription row: parse id: %w", err)
	}

	userID, err := uuid.FromBytes(dto.UserID)
	if err != nil {
		return nil, fmt.Errorf("map push subscription row: parse user id: %w", err)
	}

	endpointHash, err := validEndpointHash(dto.EndpointHash)
	if err != nil {
		return nil, fmt.Errorf("map push subscription row: %w", err)
	}

	return notification.RestorePushSubscription(notification.RestorePushSubscriptionParams{
		ID:           id,
		UserID:       userID,
		Endpoint:     dto.Endpoint,
		EndpointHash: endpointHash,
		P256DH:       dto.P256DH,
		AuthSecret:   dto.AuthSecret,
		UserAgent:    nullStringPtr(dto.UserAgent),
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
	}), nil
}

func marshalBinaryUUID(id uuid.UUID, field string) ([]byte, error) {
	bytes, err := id.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal %s: %w", field, err)
	}
	return bytes, nil
}

func validEndpointHash(endpointHash []byte) ([]byte, error) {
	if len(endpointHash) != sha256.Size {
		return nil, notification.ErrInvalidPushEndpointHash
	}
	return append([]byte(nil), endpointHash...), nil
}

func nullStringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
