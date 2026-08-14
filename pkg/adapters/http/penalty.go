package http

import (
	"errors"

	penaltydomain "dorm/pkg/core/domain/penalty"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PenaltyAPIHandler struct {
	penaltyUC ports.PenaltyUseCase
}

func NewPenaltyAPIHandler(penaltyUC ports.PenaltyUseCase) *PenaltyAPIHandler {
	return &PenaltyAPIHandler{penaltyUC: penaltyUC}
}

func (h *PenaltyAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/penalties", auth)
	api.Get("", h.HandleListResidents)
	api.Get("/me", h.HandleGetCurrentUserPenalties)
	api.Get("/residents", h.HandleSearchResidents)
	api.Get("/residents/:userId", h.HandleGetResidentPenalties)
	api.Post("", h.HandleCreatePenalty)
	api.Delete("/:penaltyId", h.HandleResolvePenalty)
}

func (h *PenaltyAPIHandler) HandleListResidents(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	response, err := h.penaltyUC.ListResidents(c.Context(), userID)
	if err != nil {
		return h.respondPenaltyError(c, err)
	}

	return c.JSON(response)
}

func (h *PenaltyAPIHandler) HandleSearchResidents(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	response, err := h.penaltyUC.SearchResidents(c.Context(), userID, c.Query("q"))
	if err != nil {
		return h.respondPenaltyError(c, err)
	}

	return c.JSON(response)
}

func (h *PenaltyAPIHandler) HandleGetCurrentUserPenalties(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	response, err := h.penaltyUC.GetCurrentUserPenalties(c.Context(), userID)
	if err != nil {
		return h.respondPenaltyError(c, err)
	}

	return c.JSON(response)
}

func (h *PenaltyAPIHandler) HandleGetResidentPenalties(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	residentID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор жителя"))
	}

	response, err := h.penaltyUC.GetResidentPenalties(c.Context(), userID, residentID)
	if err != nil {
		return h.respondPenaltyError(c, err)
	}

	return c.JSON(response)
}

func (h *PenaltyAPIHandler) HandleCreatePenalty(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	var request dto.CreatePenaltyRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	item, err := h.penaltyUC.CreatePenalty(c.Context(), userID, request)
	if err != nil {
		return h.respondPenaltyError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *PenaltyAPIHandler) HandleResolvePenalty(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	penaltyID, err := uuid.Parse(c.Params("penaltyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор предупреждения"))
	}

	if err := h.penaltyUC.ResolvePenalty(c.Context(), userID, penaltyID); err != nil {
		return h.respondPenaltyError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PenaltyAPIHandler) respondPenaltyError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, penaltydomain.ErrInvalidResidentID):
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор жителя"))
	case errors.Is(err, penaltydomain.ErrInvalidReason):
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректная причина предупреждения"))
	case errors.Is(err, penaltydomain.ErrInvalidWeight):
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный вес предупреждения"))
	case errors.Is(err, penaltydomain.ErrInvalidIssuedOn):
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректная дата предупреждения"))
	case errors.Is(err, penaltydomain.ErrIssuedOnInFuture):
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("дата предупреждения не может быть в будущем"))
	case errors.Is(err, penaltydomain.ErrAccessDenied):
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("доступ запрещен"))
	case errors.Is(err, penaltydomain.ErrResidentNotFound):
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("житель не найден"))
	case errors.Is(err, penaltydomain.ErrPenaltyNotFound):
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("предупреждение не найдено"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось выполнить операцию с предупреждениями"))
	}
}
