package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"dorm/pkg/infrastructure/config"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type NotificationAPIHandler struct {
	webPush             config.WebPushConfig
	subscriptionService ports.NotificationUseCase
}

func NewNotificationAPIHandler(webPush config.WebPushConfig, subscriptionService ports.NotificationUseCase) *NotificationAPIHandler {
	return &NotificationAPIHandler{
		webPush:             webPush,
		subscriptionService: subscriptionService,
	}
}

func (h *NotificationAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/notifications", auth)
	api.Get("/config", h.HandleGetConfig)
	api.Post("/subscriptions", h.HandleSubscribe)
	api.Delete("/subscriptions", h.HandleUnsubscribe)
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

func (h *NotificationAPIHandler) HandleSubscribe(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	var request dto.CreatePushSubscriptionRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	userAgentValue := strings.TrimSpace(c.Get("User-Agent"))
	var userAgent *string
	if userAgentValue != "" {
		userAgent = &userAgentValue
	}

	if err := h.subscriptionService.Subscribe(c.Context(), userID, dto.SubscribeToPushRequest{
		Endpoint:   request.Endpoint,
		P256DH:     request.Keys.P256DH,
		AuthSecret: request.Keys.Auth,
		UserAgent:  userAgent,
	}); err != nil {
		return h.respondSubscriptionError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *NotificationAPIHandler) HandleUnsubscribe(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	var request dto.DeletePushSubscriptionRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	if err := h.subscriptionService.Unsubscribe(c.Context(), userID, request.Endpoint); err != nil {
		return h.respondSubscriptionError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *NotificationAPIHandler) respondSubscriptionError(c *fiber.Ctx, err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "required"),
		strings.Contains(message, "invalid"),
		strings.Contains(message, "too long"),
		strings.Contains(message, "cannot be empty"):
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректные данные подписки"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось сохранить подписку"))
	}
}
