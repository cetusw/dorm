package http

import (
	"dorm/pkg/core/ports"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type UserAPIHandler struct {
	userUC      ports.UserUseCase
	dormitoryUC ports.DormitoryUseCase
}

func NewUserAPIHandler(userUC ports.UserUseCase, dormitoryUC ports.DormitoryUseCase) *UserAPIHandler {
	return &UserAPIHandler{
		userUC:      userUC,
		dormitoryUC: dormitoryUC,
	}
}

func (h *UserAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/users", auth)
	api.Get("", h.HandleGetResidents)
	api.Get("/options", h.HandleGetUserOptions)
}

func (h *UserAPIHandler) HandleGetResidents(c *fiber.Ctx) error {
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

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	response, err := h.userUC.GetResidentsResponse(c.Context(), dormitoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить список жителей"))
	}

	return c.JSON(response)
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
