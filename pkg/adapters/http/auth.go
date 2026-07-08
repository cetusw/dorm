package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const currentUserIDLocalKey = "current_user_id"

func DevCurrentUserMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawUserID := c.Get("X-User-ID")
		if rawUserID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("missing X-User-ID header"))
		}

		userID, err := uuid.Parse(rawUserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("invalid X-User-ID header"))
		}

		c.Locals(currentUserIDLocalKey, userID)
		return c.Next()
	}
}

func currentUserID(c *fiber.Ctx) (uuid.UUID, error) {
	value := c.Locals(currentUserIDLocalKey)
	if value == nil {
		return uuid.Nil, fmt.Errorf("current user is not set")
	}

	userID, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("current user has invalid type")
	}

	return userID, nil
}
