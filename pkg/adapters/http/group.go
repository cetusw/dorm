package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	dutysettingsuc "dorm/pkg/core/usecase/dutysettings"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type GroupAPIHandler struct {
	dormitoryUC    ports.DormitoryUseCase
	dutySettingsUC ports.DutySettingsUseCase
}

func NewGroupAPIHandler(dormitoryUC ports.DormitoryUseCase, dutySettingsUC ports.DutySettingsUseCase) *GroupAPIHandler {
	return &GroupAPIHandler{
		dormitoryUC:    dormitoryUC,
		dutySettingsUC: dutySettingsUC,
	}
}

func (h *GroupAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/groups", auth)
	api.Get("", h.HandleGetGroups)
	api.Get("/options", h.HandleGetDormitoryUserOptions)
	api.Get("/:id", h.HandleGetGroup)
	api.Get("/:id/duty-settings", h.HandleGetDutySettings)
	api.Post("/:id/duty-settings/areas", h.HandleCreateDutySettingsArea)
	api.Put("/:id/duty-settings/areas/:areaId", h.HandleUpdateDutySettingsArea)
	api.Delete("/:id/duty-settings/areas/:areaId", h.HandleDeleteDutySettingsArea)
	api.Post("/:id/duty-settings/areas/:areaId/tasks", h.HandleCreateDutySettingsTask)
	api.Put("/:id/duty-settings/tasks/:taskId", h.HandleUpdateDutySettingsTask)
	api.Delete("/:id/duty-settings/tasks/:taskId", h.HandleDeleteDutySettingsTask)
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

func (h *GroupAPIHandler) HandleGetDutySettings(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	response, err := h.dutySettingsUC.GetDutySettings(c.Context(), userID, groupID)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.JSON(response)
}

func (h *GroupAPIHandler) HandleCreateDutySettingsArea(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	var req dto.CreateAreaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	area, err := h.dutySettingsUC.CreateArea(c.Context(), userID, groupID, req)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(area)
}

func (h *GroupAPIHandler) HandleUpdateDutySettingsArea(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	areaID, err := strconv.Atoi(c.Params("areaId"))
	if err != nil || areaID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор территории"))
	}

	var req dto.UpdateAreaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	area, err := h.dutySettingsUC.UpdateArea(c.Context(), userID, groupID, areaID, req)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.JSON(area)
}

func (h *GroupAPIHandler) HandleDeleteDutySettingsArea(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	areaID, err := strconv.Atoi(c.Params("areaId"))
	if err != nil || areaID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор территории"))
	}

	if err := h.dutySettingsUC.DeleteArea(c.Context(), userID, groupID, areaID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleCreateDutySettingsTask(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	areaID, err := strconv.Atoi(c.Params("areaId"))
	if err != nil || areaID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор территории"))
	}

	var req dto.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	task, err := h.dutySettingsUC.CreateTask(c.Context(), userID, groupID, areaID, req)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

func (h *GroupAPIHandler) HandleUpdateDutySettingsTask(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	taskID, err := uuid.Parse(c.Params("taskId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор задачи"))
	}

	var req dto.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	task, err := h.dutySettingsUC.UpdateTask(c.Context(), userID, groupID, taskID, req)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.JSON(task)
}

func (h *GroupAPIHandler) HandleDeleteDutySettingsTask(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	taskID, err := uuid.Parse(c.Params("taskId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор задачи"))
	}

	if err := h.dutySettingsUC.DeleteTask(c.Context(), userID, groupID, taskID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) parseDutySettingsAccess(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	userID, err := currentUserID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	return userID, groupID, nil
}

func (h *GroupAPIHandler) respondDutySettingsError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, dutysettingsuc.ErrAccessDenied):
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("доступ запрещен"))
	case err.Error() == "группа не найдена":
		return c.Status(fiber.StatusNotFound).JSON(errorResponse(err.Error()))
	case err.Error() == "территория не найдена":
		return c.Status(fiber.StatusNotFound).JSON(errorResponse(err.Error()))
	case err.Error() == "задача не найдена":
		return c.Status(fiber.StatusNotFound).JSON(errorResponse(err.Error()))
	default:
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
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
