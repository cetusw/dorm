package http

import (
	"dorm/pkg/core/ports"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ResidentAPIHandler struct {
	residentDutyUC ports.ResidentDutyUseCase
}

func NewResidentAPIHandler(residentDutyUC ports.ResidentDutyUseCase) *ResidentAPIHandler {
	return &ResidentAPIHandler{
		residentDutyUC: residentDutyUC,
	}
}

func (h *ResidentAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/resident", auth)

	api.Get("/current-duty", h.HandleGetCurrentDuty)
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
