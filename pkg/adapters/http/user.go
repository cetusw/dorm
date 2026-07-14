package http

import (
	"dorm/pkg/core/ports"

	"github.com/gofiber/fiber/v2"
)

type UserAPIHandler struct {
	dormitoryUC ports.DormitoryUseCase
}

func NewUserAPIHandler(dormitoryUC ports.DormitoryUseCase) *UserAPIHandler {
	return &UserAPIHandler{
		dormitoryUC: dormitoryUC,
	}
}

func (h *UserAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/users", auth)
	api.Get("/options", h.HandleGetUserOptions)
}

func (h *UserAPIHandler) HandleGetUserOptions(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	canManage, err := h.dormitoryUC.CanManageDormitories(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось проверить доступ"))
	}
	if !canManage {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("доступ запрещен"))
	}

	response, err := h.dormitoryUC.GetUserOptionsResponse(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить список жителей"))
	}

	return c.JSON(response)
}
