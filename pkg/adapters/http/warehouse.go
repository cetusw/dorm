package http

import (
	"errors"

	warehousedomain "dorm/pkg/core/domain/warehouse"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WarehouseAPIHandler struct {
	warehouseUC ports.WarehouseUseCase
}

func NewWarehouseAPIHandler(warehouseUC ports.WarehouseUseCase) *WarehouseAPIHandler {
	return &WarehouseAPIHandler{warehouseUC: warehouseUC}
}

func (h *WarehouseAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/warehouse", auth)
	api.Get("", h.HandleGetWarehouse)
	api.Get("/items/:itemId/history", h.HandleGetItemHistory)
	api.Post("/items", h.HandleCreateItem)
	api.Put("/items/:itemId", h.HandleUpdateItem)
	api.Delete("/items/:itemId", h.HandleDeleteItem)
	api.Post("/items/:itemId/add", h.HandleAddItems)
	api.Post("/items/:itemId/write-off", h.HandleWriteOffItems)
}

func (h *WarehouseAPIHandler) HandleGetWarehouse(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	response, err := h.warehouseUC.GetWarehouse(c.Context(), userID)
	if err != nil {
		return h.respondWarehouseError(c, err)
	}

	return c.JSON(response)
}

func (h *WarehouseAPIHandler) HandleGetItemHistory(c *fiber.Ctx) error {
	userID, itemID, err := h.parseWarehouseItemRequest(c)
	if err != nil {
		return err
	}

	response, err := h.warehouseUC.GetItemHistory(c.Context(), userID, itemID)
	if err != nil {
		return h.respondWarehouseError(c, err)
	}

	return c.JSON(response)
}

func (h *WarehouseAPIHandler) HandleCreateItem(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	var request dto.CreateWarehouseItemRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	item, err := h.warehouseUC.CreateItem(c.Context(), userID, request)
	if err != nil {
		return h.respondWarehouseError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *WarehouseAPIHandler) HandleUpdateItem(c *fiber.Ctx) error {
	userID, itemID, err := h.parseWarehouseItemRequest(c)
	if err != nil {
		return err
	}

	var request dto.UpdateWarehouseItemRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	item, err := h.warehouseUC.UpdateItem(c.Context(), userID, itemID, request)
	if err != nil {
		return h.respondWarehouseError(c, err)
	}

	return c.JSON(item)
}

func (h *WarehouseAPIHandler) HandleDeleteItem(c *fiber.Ctx) error {
	userID, itemID, err := h.parseWarehouseItemRequest(c)
	if err != nil {
		return err
	}

	if err := h.warehouseUC.DeleteItem(c.Context(), userID, itemID); err != nil {
		return h.respondWarehouseError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *WarehouseAPIHandler) HandleAddItems(c *fiber.Ctx) error {
	return h.handleMovement(c, true)
}

func (h *WarehouseAPIHandler) HandleWriteOffItems(c *fiber.Ctx) error {
	return h.handleMovement(c, false)
}

func (h *WarehouseAPIHandler) handleMovement(c *fiber.Ctx, isAdd bool) error {
	userID, itemID, err := h.parseWarehouseItemRequest(c)
	if err != nil {
		return err
	}

	var request dto.WarehouseMovementRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	var item *dto.WarehouseItem
	if isAdd {
		item, err = h.warehouseUC.AddItems(c.Context(), userID, itemID, request)
	} else {
		item, err = h.warehouseUC.WriteOffItems(c.Context(), userID, itemID, request)
	}
	if err != nil {
		return h.respondWarehouseError(c, err)
	}

	return c.JSON(item)
}

func (h *WarehouseAPIHandler) parseWarehouseItemRequest(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	userID, err := currentUserID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return uuid.Nil, uuid.Nil, c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор позиции"))
	}

	return userID, itemID, nil
}

func (h *WarehouseAPIHandler) respondWarehouseError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, warehousedomain.ErrInvalidItemName):
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректное название позиции"))
	case errors.Is(err, warehousedomain.ErrInvalidMovementQuantity):
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректное количество"))
	case errors.Is(err, warehousedomain.ErrCommentTooLong):
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("комментарий не должен превышать 256 символов"))
	case errors.Is(err, warehousedomain.ErrAccessDenied):
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("доступ запрещен"))
	case errors.Is(err, warehousedomain.ErrItemNotFound):
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("позиция не найдена"))
	case errors.Is(err, warehousedomain.ErrDuplicateItemName):
		return c.Status(fiber.StatusConflict).JSON(errorResponse("позиция с таким названием уже существует"))
	case errors.Is(err, warehousedomain.ErrInsufficientItems):
		return c.Status(fiber.StatusConflict).JSON(errorResponse("недостаточно предметов для списания"))
	case errors.Is(err, warehousedomain.ErrItemHasNonZeroBalance):
		return c.Status(fiber.StatusConflict).JSON(errorResponse("нельзя удалить позицию с ненулевым остатком"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось выполнить операцию со складом"))
	}
}
