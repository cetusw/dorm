package http

import (
	"crypto/sha256"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"dorm/pkg/infrastructure/config"

	"github.com/gofiber/fiber/v2"
)

func OriginProtection(origins ...string) fiber.Handler {
	allowed := make(map[string]bool, len(origins))
	for _, origin := range origins {
		normalized, err := config.NormalizeOrigin(origin)
		if err == nil {
			allowed[normalized] = true
		}
	}
	return func(c *fiber.Ctx) error {
		switch c.Method() {
		case "POST", "PUT", "PATCH", "DELETE":
			actual, err := config.NormalizeOrigin(c.Get("Origin"))
			if err != nil || !allowed[actual] {
				c.Locals("auth.csrf_rejected", true)
				return fiber.NewError(403, "недопустимый источник запроса")
			}
		}
		return c.Next()
	}
}

type loginLimitReason string

const (
	accountSourceLimit loginLimitReason = "account_source_limit"
	sourceRateLimit    loginLimitReason = "source_rate_limit"
	concurrencyLimit   loginLimitReason = "concurrency_limit"
	limiterCapacity    loginLimitReason = "limiter_capacity"
)

type loginWindow struct {
	started time.Time
	count   int
	tokens  float64
	updated time.Time
	pending int
}
type loginLimiter struct {
	mu      sync.Mutex
	windows map[[32]byte]*loginWindow
	slots   chan struct{}
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{windows: make(map[[32]byte]*loginWindow), slots: make(chan struct{}, 4)}
}
func (l *loginLimiter) reserve(source, login string) (func(bool), loginLimitReason, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if len(l.windows) > 9000 {
		for key, w := range l.windows {
			if w.pending == 0 && now.Sub(w.updated) >= 15*time.Minute {
				delete(l.windows, key)
			}
		}
	}
	sourceKey := sha256.Sum256([]byte("source:" + source))
	accountKey := sha256.Sum256([]byte("account:" + source + ":" + strings.ToLower(strings.TrimSpace(login))))
	sourceWindow := l.windows[sourceKey]
	accountWindow := l.windows[accountKey]
	needed := 0
	if sourceWindow == nil {
		needed++
	}
	if accountWindow == nil {
		needed++
	}
	if len(l.windows)+needed > 10000 {
		retry := 15 * time.Minute
		for _, window := range l.windows {
			if window.pending == 0 {
				retry = min(retry, 15*time.Minute-now.Sub(window.updated))
			}
		}
		return nil, limiterCapacity, max(1, int(math.Ceil(retry.Seconds())))
	}
	if sourceWindow == nil {
		sourceWindow = &loginWindow{tokens: 250, updated: now}
	}
	sourceWindow.tokens = math.Min(250, sourceWindow.tokens+now.Sub(sourceWindow.updated).Seconds()*250/900)
	sourceWindow.updated = now
	if accountWindow == nil || now.Sub(accountWindow.started) >= 15*time.Minute {
		accountWindow = &loginWindow{started: now, updated: now}
	}
	l.windows[sourceKey] = sourceWindow
	l.windows[accountKey] = accountWindow
	if accountWindow.count >= 5 {
		return nil, accountSourceLimit, max(1, int(math.Ceil((15*time.Minute - now.Sub(accountWindow.started)).Seconds())))
	}
	if sourceWindow.tokens < 1 {
		return nil, sourceRateLimit, max(1, int(math.Ceil((1-sourceWindow.tokens)*900/250)))
	}
	select {
	case l.slots <- struct{}{}:
	default:
		return nil, concurrencyLimit, 2
	}
	sourceWindow.tokens--
	sourceWindow.pending++
	accountWindow.count++
	accountWindow.pending++
	accountWindow.updated = now
	var once sync.Once
	return func(failed bool) {
		once.Do(func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			<-l.slots
			sourceWindow.pending--
			accountWindow.pending--
			if !failed {
				sourceWindow.tokens = math.Min(250, sourceWindow.tokens+1)
				accountWindow.count--
			}
		})
	}, "", 0
}

func loginSource(c *fiber.Ctx, production bool) (string, error) {
	if c.App().Config().EnableTrustedProxyCheck && c.IsProxyTrusted() {
		source := net.ParseIP(c.Get("X-Real-IP"))
		if source == nil {
			return "", fiber.NewError(503, "не удалось определить источник запроса")
		}
		return source.String(), nil
	}
	if production {
		return "", fiber.NewError(503, "не настроен доверенный прокси")
	}
	return c.Context().RemoteIP().String(), nil
}
