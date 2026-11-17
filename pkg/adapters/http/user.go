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
	api.Get("", apiDeadline, h.HandleGetResidents)
	api.Get("/options", apiDeadline, h.HandleGetUserOptions)
	api.Get("/:id", apiDeadline, h.HandleGetResident)
	api.Post("", apiDeadline, h.HandleCreateResident)
	api.Put("/:id", apiDeadline, h.HandleUpdateResident)
	api.Delete("/:id", apiDeadline, h.HandleDeleteResident)
}

func (h *UserAPIHandler) HandleGetResidents(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	canManage, err := h.dormitoryUC.CanManageDormitories(c.UserContext(), userID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось проверить доступ"))
	}
	if !canManage {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("доступ запрещен"))
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	response, err := h.userUC.GetResidentsResponse(c.UserContext(), dormitoryID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
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

	resident, err := h.userUC.GetResidentDetails(c.UserContext(), residentID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
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

	resident, err := h.userUC.CreateResident(c.UserContext(), req)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
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

	resident, err := h.userUC.UpdateResident(c.UserContext(), residentID, req)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
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

	if err := h.userUC.DeleteResident(c.UserContext(), residentID); err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *UserAPIHandler) HandleGetUserOptions(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	response, err := h.dormitoryUC.GetUserOptionsResponse(c.UserContext())
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить список жителей"))
	}

	return c.JSON(response)
}

func (h *UserAPIHandler) requireDormitoryManagementAccess(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "требуется авторизация")
	}

	canManage, err := h.dormitoryUC.CanManageDormitories(c.UserContext(), userID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось проверить доступ")
	}
	if !canManage {
		return fiber.NewError(fiber.StatusForbidden, "доступ запрещен")
	}

	return nil
}
