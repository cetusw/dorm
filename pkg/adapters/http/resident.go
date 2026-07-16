package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ResidentAPIHandler struct {
	residentDutyUC ports.ResidentDutyUseCase
	dormitoryUC    ports.DormitoryUseCase
	cleaningUC     ports.CleaningUseCase
}

func NewResidentAPIHandler(
	residentDutyUC ports.ResidentDutyUseCase,
	dormitoryUC ports.DormitoryUseCase,
	cleaningUC ports.CleaningUseCase,
) *ResidentAPIHandler {
	return &ResidentAPIHandler{
		residentDutyUC: residentDutyUC,
		dormitoryUC:    dormitoryUC,
		cleaningUC:     cleaningUC,
	}
}

func (h *ResidentAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/resident", auth)

	api.Get("/current-duty", h.HandleGetCurrentDuty)
	api.Post("/duties", h.HandleCreateDormitoryDutyWeek)
	api.Post("/tasks/:taskId/take", h.HandleTakeTask)
	api.Post("/tasks/:taskId/return", h.HandleReturnTask)
	api.Post("/tasks/:taskId/complete", h.HandleCompleteTask)
	api.Post("/tasks/:taskId/open", h.HandleOpenTask)
	api.Post("/tasks/:taskId/verify", h.HandleVerifyTask)
}

func (h *ResidentAPIHandler) HandleGetCurrentDuty(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(err.Error()))
	}

	groupID, err := optionalGroupID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	currentDuty, err := h.residentDutyUC.GetCurrentDuty(c.Context(), userID, groupID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.JSON(currentDuty)
}

func (h *ResidentAPIHandler) HandleCreateDormitoryDutyWeek(c *fiber.Ctx) error {
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

	var req dto.CreateDormitoryDutyWeekRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	startDate, endDate, err := parseResidentDutyWeekDates(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.cleaningUC.StartNewDutiesForDormitory(c.Context(), req.DormitoryID, startDate, endDate); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ResidentAPIHandler) HandleTakeTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.TakeTask(c.Context(), userID, taskID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return h.respondWithCurrentDuty(c, userID)
}

func (h *ResidentAPIHandler) HandleReturnTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.ReturnTask(c.Context(), userID, taskID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return h.respondWithCurrentDuty(c, userID)
}

func (h *ResidentAPIHandler) HandleCompleteTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.CompleteTask(c.Context(), userID, taskID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return h.respondWithCurrentDuty(c, userID)
}

func (h *ResidentAPIHandler) HandleOpenTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.OpenTask(c.Context(), userID, taskID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return h.respondWithCurrentDuty(c, userID)
}

func (h *ResidentAPIHandler) HandleVerifyTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.VerifyTask(c.Context(), userID, taskID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return h.respondWithCurrentDuty(c, userID)
}

func parseResidentTaskAction(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	userID, err := currentUserID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	taskID, err := uuid.Parse(c.Params("taskId"))
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return userID, taskID, nil
}

func errorResponse(message string) fiber.Map {
	return fiber.Map{
		"error": message,
	}
}

func (h *ResidentAPIHandler) respondWithCurrentDuty(c *fiber.Ctx, userID uuid.UUID) error {
	groupID, err := optionalGroupID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	currentDuty, err := h.residentDutyUC.GetCurrentDuty(c.Context(), userID, groupID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.JSON(currentDuty)
}

func optionalGroupID(c *fiber.Ctx) (*uuid.UUID, error) {
	rawGroupID := c.Query("group_id")
	if rawGroupID == "" {
		return nil, nil
	}

	groupID, err := uuid.Parse(rawGroupID)
	if err != nil {
		return nil, err
	}

	return &groupID, nil
}

func parseResidentDutyWeekDates(req dto.CreateDormitoryDutyWeekRequest) (time.Time, time.Time, error) {
	if req.DormitoryID <= 0 {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "некорректный идентификатор общежития")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "некорректная дата начала")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "некорректная дата окончания")
	}

	return startDate, endDate, nil
}
