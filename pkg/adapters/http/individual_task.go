package http

import (
	"context"
	individual "dorm/pkg/core/domain/individualtask"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"strconv"
)

type IndividualTaskAPIHandler struct{ uc ports.IndividualTaskUseCase }

func NewIndividualTaskAPIHandler(uc ports.IndividualTaskUseCase) *IndividualTaskAPIHandler {
	return &IndividualTaskAPIHandler{uc}
}
func (h *IndividualTaskAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	a := app.Group("/api/v1/individual-tasks", auth)
	a.Get("/residents", h.search)
	a.Get("/areas", h.areas)
	a.Get("/residents/:residentId", h.listResident)
	a.Get("/me", h.mine)
	a.Get("/review", h.review)
	a.Get("/:taskId", h.get)
	a.Post("", h.create)
	a.Put("/:taskId", h.update)
	a.Delete("/:taskId", h.delete)
	a.Post("/:taskId/complete", h.complete)
	a.Post("/:taskId/reject", h.reject)
	a.Post("/:taskId/verify", h.verify)
}
func (h *IndividualTaskAPIHandler) user(c *fiber.Ctx) (uuid.UUID, error) { return currentUserID(c) }
func taskID(c *fiber.Ctx) (uuid.UUID, error)                             { return uuid.Parse(c.Params("taskId")) }
func (h *IndividualTaskAPIHandler) create(c *fiber.Ctx) error {
	u, e := h.user(c)
	if e != nil {
		return unauthorized(c)
	}
	var r dto.IndividualTaskRequest
	if e = c.BodyParser(&r); e != nil {
		return bad(c, "некорректный формат запроса")
	}
	x, e := h.uc.Create(c.Context(), u, r)
	if e != nil {
		return h.err(c, e)
	}
	return c.Status(201).JSON(x)
}
func (h *IndividualTaskAPIHandler) update(c *fiber.Ctx) error {
	u, e := h.user(c)
	if e != nil {
		return unauthorized(c)
	}
	id, e := taskID(c)
	if e != nil {
		return bad(c, "некорректный идентификатор задачи")
	}
	var r dto.IndividualTaskRequest
	if e = c.BodyParser(&r); e != nil {
		return bad(c, "некорректный формат запроса")
	}
	x, e := h.uc.Update(c.Context(), u, id, r)
	if e != nil {
		return h.err(c, e)
	}
	return c.JSON(x)
}
func (h *IndividualTaskAPIHandler) delete(c *fiber.Ctx) error {
	u, e := h.user(c)
	if e != nil {
		return unauthorized(c)
	}
	id, e := taskID(c)
	if e != nil {
		return bad(c, "некорректный идентификатор задачи")
	}
	if e = h.uc.Delete(c.Context(), u, id); e != nil {
		return h.err(c, e)
	}
	return c.SendStatus(204)
}
func (h *IndividualTaskAPIHandler) complete(c *fiber.Ctx) error { return h.action(c, h.uc.Complete) }
func (h *IndividualTaskAPIHandler) reject(c *fiber.Ctx) error   { return h.action(c, h.uc.Reject) }
func (h *IndividualTaskAPIHandler) verify(c *fiber.Ctx) error   { return h.action(c, h.uc.Verify) }
func (h *IndividualTaskAPIHandler) action(c *fiber.Ctx, f func(context.Context, uuid.UUID, uuid.UUID) (*dto.IndividualTaskItem, error)) error {
	u, e := h.user(c)
	if e != nil {
		return unauthorized(c)
	}
	id, e := taskID(c)
	if e != nil {
		return bad(c, "некорректный идентификатор задачи")
	}
	x, e := f(c.Context(), u, id)
	if e != nil {
		return h.err(c, e)
	}
	return c.JSON(x)
}
func (h *IndividualTaskAPIHandler) get(c *fiber.Ctx) error {
	u, e := h.user(c)
	if e != nil {
		return unauthorized(c)
	}
	id, e := taskID(c)
	if e != nil {
		return bad(c, "некорректный идентификатор задачи")
	}
	x, e := h.uc.Get(c.Context(), u, id)
	if e != nil {
		return h.err(c, e)
	}
	return c.JSON(x)
}
func (h *IndividualTaskAPIHandler) mine(c *fiber.Ctx) error {
	u, e := h.user(c)
	if e != nil {
		return unauthorized(c)
	}
	x, e := h.uc.ListMine(c.Context(), u)
	if e != nil {
		return h.err(c, e)
	}
	return c.JSON(x)
}
func (h *IndividualTaskAPIHandler) review(c *fiber.Ctx) error {
	u, e := h.user(c)
	if e != nil {
		return unauthorized(c)
	}
	x, e := h.uc.ListReview(c.Context(), u)
	if e != nil {
		return h.err(c, e)
	}
	return c.JSON(x)
}
func (h *IndividualTaskAPIHandler) listResident(c *fiber.Ctx) error {
	u, e := h.user(c)
	if e != nil {
		return unauthorized(c)
	}
	r, e := uuid.Parse(c.Params("residentId"))
	if e != nil {
		return bad(c, "некорректный идентификатор жителя")
	}
	x, e := h.uc.ListResident(c.Context(), u, r)
	if e != nil {
		return h.err(c, e)
	}
	return c.JSON(x)
}
func (h *IndividualTaskAPIHandler) search(c *fiber.Ctx) error {
	u, e := h.user(c)
	if e != nil {
		return unauthorized(c)
	}
	x, e := h.uc.SearchResidents(c.Context(), u, c.Query("q"))
	if e != nil {
		return h.err(c, e)
	}
	return c.JSON(x)
}
func (h *IndividualTaskAPIHandler) areas(c *fiber.Ctx) error {
	u, e := h.user(c)
	if e != nil {
		return unauthorized(c)
	}
	d, e := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if e != nil || d < 1 {
		return bad(c, "некорректный идентификатор общежития")
	}
	x, e := h.uc.ListAreas(c.Context(), u, d)
	if e != nil {
		return h.err(c, e)
	}
	return c.JSON(x)
}
func (h *IndividualTaskAPIHandler) err(c *fiber.Ctx, e error) error {
	switch {
	case errors.Is(e, individual.ErrAccessDenied):
		return c.Status(403).JSON(errorResponse("доступ запрещен"))
	case errors.Is(e, individual.ErrNotFound):
		return c.Status(404).JSON(errorResponse("задача не найдена"))
	case errors.Is(e, individual.ErrInvalidTitle), errors.Is(e, individual.ErrInvalidWeight), errors.Is(e, individual.ErrInvalidDeadline), errors.Is(e, individual.ErrInvalidArea), errors.Is(e, individual.ErrInvalidResident):
		return c.Status(400).JSON(errorResponse("некорректные данные индивидуальной задачи"))
	case errors.Is(e, individual.ErrCapacityExceeded):
		return c.Status(409).JSON(errorResponse("вес превышает доступный резерв предупреждений"))
	case errors.Is(e, individual.ErrVersionConflict), errors.Is(e, individual.ErrInvalidTransition):
		return c.Status(409).JSON(errorResponse("состояние задачи уже изменилось"))
	default:
		return c.Status(500).JSON(errorResponse("не удалось выполнить операцию с индивидуальной задачей"))
	}
}
func unauthorized(c *fiber.Ctx) error {
	return c.Status(401).JSON(errorResponse("требуется авторизация"))
}
func bad(c *fiber.Ctx, s string) error { return c.Status(400).JSON(errorResponse(s)) }
