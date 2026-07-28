package http

import (
	"dorm/pkg/core/ports/dto"
	"dorm/pkg/infrastructure/config"

	"github.com/gofiber/fiber/v2"
)

type NotificationAPIHandler struct {
	webPush config.WebPushConfig
}

func NewNotificationAPIHandler(webPush config.WebPushConfig) *NotificationAPIHandler {
	return &NotificationAPIHandler{webPush: webPush}
}

func (h *NotificationAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/notifications", auth)
	api.Get("/config", h.HandleGetConfig)
}

func (h *NotificationAPIHandler) HandleGetConfig(c *fiber.Ctx) error {
	if !h.webPush.Enabled {
		return c.JSON(dto.WebPushConfigResponse{
			Enabled:   false,
			PublicKey: nil,
		})
	}

	publicKey := h.webPush.PublicKey
	return c.JSON(dto.WebPushConfigResponse{
		Enabled:   true,
		PublicKey: &publicKey,
	})
}
