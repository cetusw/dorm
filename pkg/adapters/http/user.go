package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	api.Get("/:id", h.HandleGetResident)
	api.Post("", h.HandleCreateResident)
	api.Put("/:id", h.HandleUpdateResident)
	api.Delete("/:id", h.HandleDeleteResident)
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

func (h *UserAPIHandler) HandleGetResident(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	residentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор жителя"))
	}

	resident, err := h.userUC.GetResidentDetails(c.Context(), residentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить жителя"))
	}
	if resident == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("житель не найден"))
	}

	return c.JSON(resident)
}

func (h *UserAPIHandler) HandleCreateResident(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	var req dto.CreateResidentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	resident, err := h.userUC.CreateResident(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(resident)
}

func (h *UserAPIHandler) HandleUpdateResident(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	residentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор жителя"))
	}

	var req dto.UpdateResidentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	resident, err := h.userUC.UpdateResident(c.Context(), residentID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
	if resident == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("житель не найден"))
	}

	return c.JSON(resident)
}

func (h *UserAPIHandler) HandleDeleteResident(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	residentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор жителя"))
	}

	if err := h.userUC.DeleteResident(c.Context(), residentID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *UserAPIHandler) HandleGetUserOptions(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	response, err := h.dormitoryUC.GetUserOptionsResponse(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить список жителей"))
	}

	return c.JSON(response)
}

func (h *UserAPIHandler) requireDormitoryManagementAccess(c *fiber.Ctx) error {
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

	return nil
}
