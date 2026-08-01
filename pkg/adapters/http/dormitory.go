package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"

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
	api.Get("/:id", h.HandleGetDormitory)
	api.Post("", h.HandleCreateDormitory)
	api.Put("/:id", h.HandleUpdateDormitory)
	api.Delete("/:id", h.HandleDeleteDormitory)
}

func (h *DormitoryAPIHandler) HandleGetDormitories(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	response, err := h.dormitoryUC.GetDormitoriesResponse(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить общежития"))
	}

	return c.JSON(response)
}

func (h *DormitoryAPIHandler) HandleGetDormitory(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := parseIntParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	dormitory, err := h.dormitoryUC.GetDormitoryDetails(c.Context(), dormitoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить общежитие"))
	}
	if dormitory == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("общежитие не найдено"))
	}

	return c.JSON(dormitory)
}

func (h *DormitoryAPIHandler) HandleCreateDormitory(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	var req dto.CreateDormitoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	dormitory, err := h.dormitoryUC.CreateDormitoryDetails(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(dormitory)
}

func (h *DormitoryAPIHandler) HandleUpdateDormitory(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := parseIntParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	var req dto.UpdateDormitoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	dormitory, err := h.dormitoryUC.UpdateDormitoryDetails(c.Context(), dormitoryID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
	if dormitory == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("общежитие не найдено"))
	}

	return c.JSON(dormitory)
}

func (h *DormitoryAPIHandler) HandleDeleteDormitory(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := parseIntParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	dormitory, err := h.dormitoryUC.GetDormitoryDetails(c.Context(), dormitoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить общежитие"))
	}
	if dormitory == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("общежитие не найдено"))
	}

	if err := h.dormitoryUC.DeleteDormitory(c.Context(), dormitoryID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *DormitoryAPIHandler) requireDormitoryManagementAccess(c *fiber.Ctx) error {
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
