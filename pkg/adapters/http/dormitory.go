package http

import (
	"dorm/pkg/core/ports"

	"github.com/gofiber/fiber/v2"
)

type DormitoryAPIHandler struct {
	dormitoryUC ports.DormitoryUseCase
}

func NewDormitoryAPIHandler(dormitoryUC ports.DormitoryUseCase) *DormitoryAPIHandler {
	return &DormitoryAPIHandler{
		dormitoryUC: dormitoryUC,
	}
}

func (h *DormitoryAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/dormitories", auth)
	api.Get("", h.HandleGetDormitories)
}

func (h *DormitoryAPIHandler) HandleGetDormitories(c *fiber.Ctx) error {
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

	response, err := h.dormitoryUC.GetDormitoriesResponse(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить общежития"))
	}

	return c.JSON(response)
}
