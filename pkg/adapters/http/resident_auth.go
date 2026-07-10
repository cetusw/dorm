package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"

	"github.com/gofiber/fiber/v2"
)

type ResidentAuthHandler struct {
	userUC     ports.UserUseCase
	authSecret string
}

func NewResidentAuthHandler(userUC ports.UserUseCase, authSecret string) *ResidentAuthHandler {
	return &ResidentAuthHandler{
		userUC:     userUC,
		authSecret: authSecret,
	}
}

func (h *ResidentAuthHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/api/v1/auth")

	auth.Post("/login", h.HandleLogin)
	auth.Post("/logout", h.HandleLogout)
}

func (h *ResidentAuthHandler) HandleLogin(c *fiber.Ctx) error {
	var req dto.ResidentLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	u, err := h.userUC.AuthenticateResident(c.Context(), req.Login, req.Password)
	if err != nil {
		clearResidentSession(c)
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(err.Error()))
	}

	if err := setResidentSession(c, u.ID(), h.authSecret); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось создать сессию"))
	}

	return c.JSON(fiber.Map{
		"redirect_url": "/app/tasks",
	})
}

func (h *ResidentAuthHandler) HandleLogout(c *fiber.Ctx) error {
	clearResidentSession(c)
	return c.SendStatus(fiber.StatusNoContent)
}
