package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type AreaAPIHandler struct {
	dormitoryUC ports.DormitoryUseCase
	taskUC      ports.TaskCatalogUseCase
}

func NewAreaAPIHandler(dormitoryUC ports.DormitoryUseCase, taskUC ports.TaskCatalogUseCase) *AreaAPIHandler {
	return &AreaAPIHandler{
		dormitoryUC: dormitoryUC,
		taskUC:      taskUC,
	}
}

func (h *AreaAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/areas", auth)
	api.Get("", h.HandleGetAreas)
	api.Get("/:id", h.HandleGetArea)
	api.Post("", h.HandleCreateArea)
	api.Put("/:id", h.HandleUpdateArea)
	api.Delete("/:id", h.HandleDeleteArea)
}

func (h *AreaAPIHandler) HandleGetAreas(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	response, err := h.taskUC.GetAreasResponse(c.Context(), dormitoryID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.JSON(response)
}

func (h *AreaAPIHandler) HandleGetArea(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	areaID, err := strconv.Atoi(c.Params("id"))
	if err != nil || areaID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор территории"))
	}

	area, err := h.taskUC.GetAreaDetails(c.Context(), areaID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
	if area == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("территория не найдена"))
	}

	return c.JSON(area)
}

func (h *AreaAPIHandler) HandleCreateArea(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	var req dto.CreateAreaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	area, err := h.taskUC.CreateArea(c.Context(), dormitoryID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(area)
}

func (h *AreaAPIHandler) HandleUpdateArea(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	areaID, err := strconv.Atoi(c.Params("id"))
	if err != nil || areaID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор территории"))
	}

	var req dto.UpdateAreaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	area, err := h.taskUC.UpdateArea(c.Context(), dormitoryID, areaID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
	if area == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("территория не найдена"))
	}

	return c.JSON(area)
}

func (h *AreaAPIHandler) HandleDeleteArea(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	areaID, err := strconv.Atoi(c.Params("id"))
	if err != nil || areaID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор территории"))
	}

	if err := h.taskUC.DeleteArea(c.Context(), dormitoryID, areaID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AreaAPIHandler) requireDormitoryManagementAccess(c *fiber.Ctx) error {
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
