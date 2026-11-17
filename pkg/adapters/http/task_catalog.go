package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TaskCatalogAPIHandler struct {
	dormitoryUC ports.DormitoryUseCase
	taskUC      ports.TaskCatalogUseCase
}

func NewTaskCatalogAPIHandler(dormitoryUC ports.DormitoryUseCase, taskUC ports.TaskCatalogUseCase) *TaskCatalogAPIHandler {
	return &TaskCatalogAPIHandler{
		dormitoryUC: dormitoryUC,
		taskUC:      taskUC,
	}
}

func (h *TaskCatalogAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/task-definitions", auth)
	api.Get("", apiDeadline, h.HandleGetTasks)
	api.Get("/:id", apiDeadline, h.HandleGetTask)
	api.Post("", apiDeadline, h.HandleCreateTask)
	api.Put("/:id", apiDeadline, h.HandleUpdateTask)
	api.Delete("/:id", apiDeadline, h.HandleDeleteTask)
}

func (h *TaskCatalogAPIHandler) HandleGetTasks(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	response, err := h.taskUC.GetTasksResponse(c.UserContext(), dormitoryID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.JSON(response)
}

func (h *TaskCatalogAPIHandler) HandleGetTask(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор задачи"))
	}

	task, err := h.taskUC.GetTaskDetails(c.UserContext(), dormitoryID, taskID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
	if task == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("задача не найдена"))
	}

	return c.JSON(task)
}

func (h *TaskCatalogAPIHandler) HandleCreateTask(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	var req dto.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	task, err := h.taskUC.CreateTaskDetails(c.UserContext(), dormitoryID, req)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

func (h *TaskCatalogAPIHandler) HandleUpdateTask(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор задачи"))
	}

	var req dto.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	task, err := h.taskUC.UpdateTaskDetails(c.UserContext(), dormitoryID, taskID, req)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
	if task == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("задача не найдена"))
	}

	return c.JSON(task)
}

func (h *TaskCatalogAPIHandler) HandleDeleteTask(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор задачи"))
	}

	if err := h.taskUC.DeleteTaskDetails(c.UserContext(), dormitoryID, taskID); err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *TaskCatalogAPIHandler) requireDormitoryManagementAccess(c *fiber.Ctx) error {
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

	return nil
}
