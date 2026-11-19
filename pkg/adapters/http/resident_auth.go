package http

import (
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ResidentAuthHandler struct {
	userUC   ports.UserUseCase
	sessions ports.SessionUseCase
	secure   bool
	limiter  *loginLimiter
}

func NewResidentAuthHandler(userUC ports.UserUseCase, sessions ports.SessionUseCase, secure bool) *ResidentAuthHandler {
	return &ResidentAuthHandler{userUC: userUC, sessions: sessions, secure: secure, limiter: newLoginLimiter()}
}

func (h *ResidentAuthHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/api/v1/auth")

	auth.Post("/login", apiDeadline, h.HandleLogin)
	auth.Post("/logout", apiDeadline, h.HandleLogout)
	auth.Get("/me", ResidentAuthMiddleware(h.sessions, h.secure), apiDeadline, h.HandleCurrentUser)
}

func (h *ResidentAuthHandler) HandleLogin(c *fiber.Ctx) error {
	var req dto.ResidentLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	source, err := loginSource(c, h.secure)
	if err != nil {
		return err
	}
	finish, reason, retryAfter := h.limiter.reserve(source, req.Login)
	if reason != "" {
		c.Locals("auth.login_rate_limited", reason)
		c.Set("Retry-After", strconv.Itoa(retryAfter))
		return fiber.NewError(429, "слишком много попыток входа")
	}
	failed := false
	defer func() { finish(failed) }()

	u, err := h.userUC.AuthenticateResident(c.UserContext(), req.Login, req.Password)
	if err != nil {
		if !errors.Is(err, user.ErrInvalidCredentials) {
			return sessionHTTPError(c, err)
		}
		failed = true
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("неверный логин или пароль"))
	}

	session, err := h.sessions.Create(c.UserContext(), u.ID(), u.PasswordHash())
	if err != nil {
		return sessionHTTPError(c, err)
	}
	setResidentSession(c, session.Token, session.ExpiresAt, h.secure)

	return c.JSON(fiber.Map{
		"redirect_url": "/app/tasks",
	})
}

func (h *ResidentAuthHandler) HandleLogout(c *fiber.Ctx) error {
	if err := h.sessions.Revoke(c.UserContext(), c.Cookies(residentSessionCookie)); err != nil {
		return sessionHTTPError(c, err)
	}
	clearResidentSession(c, h.secure)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ResidentAuthHandler) HandleCurrentUser(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	currentUser, err := h.userUC.GetCurrentUser(c.UserContext(), userID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить текущего пользователя"))
	}
	if currentUser == nil {
		clearResidentSession(c, h.secure)
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	return c.JSON(currentUser)
}
