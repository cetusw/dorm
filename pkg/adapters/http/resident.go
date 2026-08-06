package http

import (
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	residentuc "dorm/pkg/core/usecase/resident"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ResidentAPIHandler struct {
	residentDutyUC ports.ResidentDutyUseCase
}

type residentTaskAssignedConflictResponse struct {
	Error string                `json:"error"`
	Task  *dto.ResidentDutyTask `json:"task,omitempty"`
}

func NewResidentAPIHandler(
	residentDutyUC ports.ResidentDutyUseCase,
) *ResidentAPIHandler {
	return &ResidentAPIHandler{
		residentDutyUC: residentDutyUC,
	}
}

func (h *ResidentAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/resident", auth)

	api.Get("/current-duty", h.HandleGetCurrentDuty)
	api.Get("/duties", h.HandleGetDutyHistory)
	api.Get("/duties/:dutyId", h.HandleGetDutyDetails)
	api.Post("/tasks/:taskId/take", h.HandleTakeTask)
	api.Post("/tasks/:taskId/return", h.HandleReturnTask)
	api.Post("/tasks/:taskId/complete", h.HandleCompleteTask)
	api.Post("/tasks/:taskId/open", h.HandleOpenTask)
	api.Post("/tasks/:taskId/verify", h.HandleVerifyTask)
	api.Post("/tasks/:taskId/reopen", h.HandleReopenTask)
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
		return h.respondResidentDutyError(c, err)
	}

	return c.JSON(currentDuty)
}

func (h *ResidentAPIHandler) HandleGetDutyHistory(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(err.Error()))
	}

	groupID, err := optionalGroupID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	history, err := h.residentDutyUC.GetDutyHistory(c.Context(), userID, groupID)
	if err != nil {
		return h.respondResidentDutyError(c, err)
	}

	return c.JSON(history)
}

func (h *ResidentAPIHandler) HandleGetDutyDetails(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(err.Error()))
	}

	dutyID, err := uuid.Parse(c.Params("dutyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор дежурства"))
	}

	dutyDetails, err := h.residentDutyUC.GetDutyDetails(c.Context(), userID, dutyID)
	if err != nil {
		return h.respondResidentDutyError(c, err)
	}

	return c.JSON(dutyDetails)
}

func (h *ResidentAPIHandler) HandleTakeTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.TakeTask(c.Context(), userID, taskID); err != nil {
		if errors.Is(err, duty.ErrTaskAssigned) {
			return h.respondResidentTaskAssignedConflict(c, userID, taskID)
		}

		return h.respondResidentTaskError(c, err)
	}

	return h.respondWithCurrentDuty(c, userID)
}

func (h *ResidentAPIHandler) HandleReturnTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.ReturnTask(c.Context(), userID, taskID); err != nil {
		return h.respondResidentTaskError(c, err)
	}

	return h.respondWithCurrentDuty(c, userID)
}

func (h *ResidentAPIHandler) HandleCompleteTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.CompleteTask(c.Context(), userID, taskID); err != nil {
		return h.respondResidentTaskError(c, err)
	}

	return h.respondWithCurrentDuty(c, userID)
}

func (h *ResidentAPIHandler) HandleOpenTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.OpenTask(c.Context(), userID, taskID); err != nil {
		return h.respondResidentTaskError(c, err)
	}

	return h.respondWithCurrentDuty(c, userID)
}

func (h *ResidentAPIHandler) HandleVerifyTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.VerifyTask(c.Context(), userID, taskID); err != nil {
		return h.respondResidentTaskError(c, err)
	}

	return h.respondWithCurrentDuty(c, userID)
}

func (h *ResidentAPIHandler) HandleReopenTask(c *fiber.Ctx) error {
	userID, taskID, err := parseResidentTaskAction(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.residentDutyUC.ReopenTask(c.Context(), userID, taskID); err != nil {
		return h.respondResidentTaskError(c, err)
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

func (h *ResidentAPIHandler) respondResidentTaskAssignedConflict(c *fiber.Ctx, userID, taskID uuid.UUID) error {
	const message = "эту задачу уже взял другой пользователь"

	groupID, err := optionalGroupID(c)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(errorResponse(message))
	}

	currentDuty, err := h.residentDutyUC.GetCurrentDuty(c.Context(), userID, groupID)
	if err != nil || currentDuty == nil {
		return c.Status(fiber.StatusConflict).JSON(errorResponse(message))
	}

	for _, task := range currentDuty.Tasks {
		if task.ID != taskID.String() {
			continue
		}

		taskCopy := task
		return c.Status(fiber.StatusConflict).JSON(residentTaskAssignedConflictResponse{
			Error: message,
			Task:  &taskCopy,
		})
	}

	return c.Status(fiber.StatusConflict).JSON(errorResponse(message))
}

func (h *ResidentAPIHandler) respondResidentTaskError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, duty.ErrTaskAssigned):
		return c.Status(fiber.StatusConflict).JSON(errorResponse("эту задачу уже взял другой пользователь"))
	case errors.Is(err, duty.ErrTaskStateConflict):
		return c.Status(fiber.StatusConflict).JSON(errorResponse("состояние задачи уже изменилось, обновите список"))
	case errors.Is(err, duty.ErrDutyActionsUnavailable):
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("Действия с этим дежурством недоступны"))
	case errors.Is(err, duty.ErrTaskNotFound):
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("задача не найдена"))
	case errors.Is(err, duty.ErrTaskAccessDenied):
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("доступ запрещен"))
	default:
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
}

func (h *ResidentAPIHandler) respondWithCurrentDuty(c *fiber.Ctx, userID uuid.UUID) error {
	groupID, err := optionalGroupID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	currentDuty, err := h.residentDutyUC.GetCurrentDuty(c.Context(), userID, groupID)
	if err != nil {
		return h.respondResidentDutyError(c, err)
	}

	return c.JSON(currentDuty)
}

func (h *ResidentAPIHandler) respondResidentDutyError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, residentuc.ErrResidentDutyGroupNotFound):
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("группа не найдена"))
	case errors.Is(err, residentuc.ErrResidentDutyAccessDenied):
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("доступ запрещен"))
	case errors.Is(err, residentuc.ErrResidentDutyNotFound):
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("дежурство не найдено"))
	default:
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
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
