package http

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})
	app.Get("/health/ready", h.ready)
}

func (h *HealthHandler) ready(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).Send(nil)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
