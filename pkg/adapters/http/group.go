package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type GroupAPIHandler struct {
	dormitoryUC ports.DormitoryUseCase
}

func NewGroupAPIHandler(dormitoryUC ports.DormitoryUseCase) *GroupAPIHandler {
	return &GroupAPIHandler{
		dormitoryUC: dormitoryUC,
	}
}

func (h *GroupAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/groups", auth)
	api.Get("", h.HandleGetGroups)
	api.Get("/options", h.HandleGetDormitoryUserOptions)
	api.Get("/:id", h.HandleGetGroup)
	api.Post("", h.HandleCreateGroup)
	api.Put("/:id", h.HandleUpdateGroup)
	api.Delete("/:id", h.HandleDeleteGroup)
}

func (h *GroupAPIHandler) HandleGetGroups(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	response, err := h.dormitoryUC.GetGroupsResponse(c.Context(), dormitoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить список групп"))
	}

	return c.JSON(response)
}

func (h *GroupAPIHandler) HandleGetDormitoryUserOptions(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	response, err := h.dormitoryUC.GetDormitoryUserOptionsResponse(c.Context(), dormitoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить список жителей"))
	}

	return c.JSON(response)
}

func (h *GroupAPIHandler) HandleGetGroup(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	group, err := h.dormitoryUC.GetGroupDetails(c.Context(), groupID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить группу"))
	}
	if group == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("группа не найдена"))
	}

	return c.JSON(group)
}

func (h *GroupAPIHandler) HandleCreateGroup(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	var req dto.CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	group, err := h.dormitoryUC.CreateGroupDetails(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(group)
}

func (h *GroupAPIHandler) HandleUpdateGroup(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	var req dto.UpdateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	group, err := h.dormitoryUC.UpdateGroupDetails(c.Context(), groupID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
	if group == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("группа не найдена"))
	}

	return c.JSON(group)
}

func (h *GroupAPIHandler) HandleDeleteGroup(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	if err := h.dormitoryUC.DeleteGroup(c.Context(), groupID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) requireDormitoryManagementAccess(c *fiber.Ctx) error {
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
