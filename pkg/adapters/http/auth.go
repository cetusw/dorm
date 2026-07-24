package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	currentUserIDLocalKey = "current_user_id"
	residentSessionCookie = "resident_session"
	residentSessionTTL    = 365 * 24 * time.Hour
)

func ResidentAuthMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseResidentSession(c.Cookies(residentSessionCookie), secret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
		}

		c.Locals(currentUserIDLocalKey, userID)
		return c.Next()
	}
}

func setResidentSession(c *fiber.Ctx, userID uuid.UUID, secret string) error {
	sessionValue, err := buildResidentSession(userID, time.Now().Add(residentSessionTTL), secret)
	if err != nil {
		return fmt.Errorf("build resident session: %w", err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     residentSessionCookie,
		Value:    sessionValue,
		Path:     "/",
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  time.Now().Add(residentSessionTTL),
	})

	return nil
}

func clearResidentSession(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     residentSessionCookie,
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
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

func buildResidentSession(userID uuid.UUID, expiresAt time.Time, secret string) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("auth secret is empty")
	}

	payload := fmt.Sprintf("%s|%d", userID.String(), expiresAt.UTC().Unix())
	mac := signResidentSession(payload, secret)

	token := fmt.Sprintf("%s|%s", payload, mac)
	return base64.RawURLEncoding.EncodeToString([]byte(token)), nil
}

func parseResidentSession(rawValue string, secret string) (uuid.UUID, error) {
	if rawValue == "" {
		return uuid.Nil, fmt.Errorf("missing resident session")
	}
	if strings.TrimSpace(secret) == "" {
		return uuid.Nil, fmt.Errorf("auth secret is empty")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(rawValue)
	if err != nil {
		return uuid.Nil, fmt.Errorf("decode resident session: %w", err)
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 3 {
		return uuid.Nil, fmt.Errorf("invalid resident session format")
	}

	payload := strings.Join(parts[:2], "|")
	expectedMAC := signResidentSession(payload, secret)
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(expectedMAC)) != 1 {
		return uuid.Nil, fmt.Errorf("invalid resident session signature")
	}

	userID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse resident session user id: %w", err)
	}

	expiresAtUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse resident session expiry: %w", err)
	}
	if time.Now().UTC().After(time.Unix(expiresAtUnix, 0).UTC()) {
		return uuid.Nil, fmt.Errorf("resident session expired")
	}

	return userID, nil
}

func signResidentSession(payload string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
