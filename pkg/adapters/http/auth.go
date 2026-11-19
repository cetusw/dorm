package http

import (
	"context"
	"errors"
	"fmt"
	nethttp "net/http"
	"time"

	"dorm/pkg/core/ports"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	currentUserIDLocalKey                  = "current_user_id"
	residentSessionCookie                  = "resident_session"
	sessionRejectionKey                    = "auth.session_rejection"
	sessionMissingCookie  sessionRejection = "missing_cookie"
)

type sessionRejection = ports.SessionRejection

func sessionHTTPError(c *fiber.Ctx, err error) error {
	var reason ports.SessionRejection
	if errors.As(err, &reason) {
		c.Locals(sessionRejectionKey, reason)
		return fiber.NewError(401, "требуется авторизация")
	}
	if isContextError(err) {
		return err
	}
	c.Locals("auth.infrastructure_error", true)
	return fiber.NewError(503, "не удалось проверить сессию")
}

func ResidentAuthMiddleware(sessions ports.SessionUseCase, secure bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		session, err := sessions.Validate(ctx, c.Cookies(residentSessionCookie))
		cancel()
		if err != nil {
			if errors.Is(err, ports.SessionRejection("user_inactive")) {
				clearResidentSession(c, secure)
			}
			return sessionHTTPError(c, err)
		}
		c.Locals(currentUserIDLocalKey, session.UserID)
		return c.Next()
	}
}

func residentCookie(secure bool) *nethttp.Cookie {
	return &nethttp.Cookie{Name: residentSessionCookie, Path: "/", HttpOnly: true, Secure: secure, SameSite: nethttp.SameSiteLaxMode}
}

func setResidentSession(c *fiber.Ctx, token string, expires time.Time, secure bool) {
	cookie := residentCookie(secure)
	cookie.Value = token
	cookie.Expires = expires
	cookie.MaxAge = int(time.Until(expires).Seconds())
	if cookie.MaxAge < 1 {
		cookie.MaxAge = 1
	}
	c.Response().Header.DelCookie(residentSessionCookie)
	c.Response().Header.Add(fiber.HeaderSetCookie, cookie.String())
}

func clearResidentSession(c *fiber.Ctx, secure bool) {
	cookie := residentCookie(secure)
	cookie.Expires = time.Unix(0, 0)
	cookie.MaxAge = -1
	c.Response().Header.DelCookie(residentSessionCookie)
	c.Response().Header.Add(fiber.HeaderSetCookie, cookie.String())
}

func currentUserID(c *fiber.Ctx) (uuid.UUID, error) {
	value := c.Locals(currentUserIDLocalKey)
	id, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("current user is not set")
	}
	return id, nil
}
