package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func parseIntParam(c *fiber.Ctx, key string) (int64, error) {
	return strconv.ParseInt(c.Params(key), 10, 64)
}
