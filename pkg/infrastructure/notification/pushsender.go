package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	domain "dorm/pkg/core/domain/notification"
	"dorm/pkg/core/ports"
	"dorm/pkg/infrastructure/config"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type NoopPushSender struct{}

func (NoopPushSender) Send(context.Context, *domain.PushSubscription, ports.PushMessage) error {
	return nil
}

type WebPushSender struct {
	options webpush.Options
}

func NewWebPushSender(cfg config.WebPushConfig) *WebPushSender {
	return &WebPushSender{
		options: webpush.Options{
			Subscriber:      webPushSubscriber(cfg.Subject),
			TTL:             60,
			VAPIDPublicKey:  cfg.PublicKey,
			VAPIDPrivateKey: cfg.PrivateKey,
		},
	}
}

func (s *WebPushSender) Send(ctx context.Context, subscription *domain.PushSubscription, message ports.PushMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal push payload: %w", err)
	}

	response, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: subscription.Endpoint(),
		Keys: webpush.Keys{
			Auth:   subscription.AuthSecret(),
			P256dh: subscription.P256DH(),
		},
	}, &s.options)
	if err != nil {
		return fmt.Errorf("send web push notification: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusGone || response.StatusCode == http.StatusNotFound {
		_, _ = io.Copy(io.Discard, response.Body)
		return ports.ErrPushSubscriptionExpired
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return fmt.Errorf("unexpected web push status: %d %s", response.StatusCode, string(body))
	}

	return nil
}

func webPushSubscriber(subject string) string {
	return strings.TrimPrefix(strings.TrimSpace(subject), "mailto:")
}
