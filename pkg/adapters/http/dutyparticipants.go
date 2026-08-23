package http

import (
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/ports"
	participants "dorm/pkg/core/usecase/dutyparticipants"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type DutyParticipantsAPIHandler struct{ participants ports.DutyParticipantsUseCase }

func NewDutyParticipantsAPIHandler(participants ports.DutyParticipantsUseCase) *DutyParticipantsAPIHandler {
	return &DutyParticipantsAPIHandler{participants: participants}
}
func (h *DutyParticipantsAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/resident/duties", auth)
	api.Get("/:dutyId/participants", h.get)
	api.Get("/:dutyId/participant-candidates", h.candidates)
	api.Post("/:dutyId/participants", h.add)
	api.Post("/:dutyId/participants/:participantId/exclude", h.exclude)
	api.Post("/:dutyId/participants/:participantId/restore", h.restore)
	api.Patch("/:dutyId/leader", h.leader)
}
func (h *DutyParticipantsAPIHandler) get(c *fiber.Ctx) error {
	actor, id, err := participantIDs(c, false)
	if err != nil {
		return c.Status(400).JSON(errorResponse(err.Error()))
	}
	v, err := h.participants.GetParticipants(c.Context(), actor, id)
	if err != nil {
		return h.err(c, err)
	}
	return c.JSON(fiber.Map{"participants": v})
}
func (h *DutyParticipantsAPIHandler) candidates(c *fiber.Ctx) error {
	actor, id, err := participantIDs(c, false)
	if err != nil {
		return c.Status(400).JSON(errorResponse(err.Error()))
	}
	v, err := h.participants.GetCandidates(c.Context(), actor, id)
	if err != nil {
		return h.err(c, err)
	}
	return c.JSON(fiber.Map{"participants": v})
}
func (h *DutyParticipantsAPIHandler) add(c *fiber.Ctx) error {
	actor, id, err := participantIDs(c, false)
	if err != nil {
		return c.Status(400).JSON(errorResponse(err.Error()))
	}
	var body struct {
		ParticipantID string `json:"participant_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(errorResponse("некорректный запрос"))
	}
	p, err := uuid.Parse(body.ParticipantID)
	if err != nil {
		return c.Status(400).JSON(errorResponse("некорректный идентификатор жителя"))
	}
	return h.err(c, h.participants.AddParticipant(c.Context(), actor, id, p))
}
func (h *DutyParticipantsAPIHandler) exclude(c *fiber.Ctx) error {
	actor, id, err := participantIDs(c, true)
	if err != nil {
		return c.Status(400).JSON(errorResponse(err.Error()))
	}
	p, _ := uuid.Parse(c.Params("participantId"))
	var body struct {
		ReplacementLeaderID *string `json:"replacement_leader_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(errorResponse("некорректный запрос"))
	}
	var replacement *uuid.UUID
	if body.ReplacementLeaderID != nil && *body.ReplacementLeaderID != "" {
		v, err := uuid.Parse(*body.ReplacementLeaderID)
		if err != nil {
			return c.Status(400).JSON(errorResponse("некорректный идентификатор нового лидера"))
		}
		replacement = &v
	}
	return h.err(c, h.participants.ExcludeParticipant(c.Context(), actor, id, p, replacement))
}
func (h *DutyParticipantsAPIHandler) restore(c *fiber.Ctx) error {
	actor, id, err := participantIDs(c, true)
	if err != nil {
		return c.Status(400).JSON(errorResponse(err.Error()))
	}
	p, _ := uuid.Parse(c.Params("participantId"))
	return h.err(c, h.participants.RestoreParticipant(c.Context(), actor, id, p))
}
func (h *DutyParticipantsAPIHandler) leader(c *fiber.Ctx) error {
	actor, id, err := participantIDs(c, false)
	if err != nil {
		return c.Status(400).JSON(errorResponse(err.Error()))
	}
	var body struct {
		LeaderID string `json:"leader_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(errorResponse("некорректный запрос"))
	}
	leader, err := uuid.Parse(body.LeaderID)
	if err != nil {
		return c.Status(400).JSON(errorResponse("некорректный идентификатор лидера"))
	}
	return h.err(c, h.participants.ChangeLeader(c.Context(), actor, id, leader))
}
func participantIDs(c *fiber.Ctx, withParticipant bool) (uuid.UUID, uuid.UUID, error) {
	actor, err := currentUserID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	id, err := uuid.Parse(c.Params("dutyId"))
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	if withParticipant {
		if _, err := uuid.Parse(c.Params("participantId")); err != nil {
			return uuid.Nil, uuid.Nil, err
		}
	}
	return actor, id, nil
}
func (h *DutyParticipantsAPIHandler) err(c *fiber.Ctx, err error) error {
	if err == nil {
		return c.SendStatus(fiber.StatusNoContent)
	}
	switch {
	case errors.Is(err, participants.ErrDutyNotFound):
		return c.Status(404).JSON(errorResponse("дежурство не найдено"))
	case errors.Is(err, participants.ErrAccessDenied):
		return c.Status(403).JSON(errorResponse("доступ запрещен"))
	case errors.Is(err, participants.ErrNotCurrent):
		return c.Status(409).JSON(errorResponse("состав можно менять только у текущего дежурства"))
	case errors.Is(err, duty.ErrParticipantAlreadyActive), errors.Is(err, duty.ErrParticipantAlreadyExcluded), errors.Is(err, duty.ErrParticipantPeriodConflict), errors.Is(err, duty.ErrParticipantNotFound), errors.Is(err, duty.ErrLeaderReplacementRequired):
		return c.Status(409).JSON(errorResponse(err.Error()))
	default:
		return c.Status(400).JSON(errorResponse(err.Error()))
	}
}
