package http

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"log/slog"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	recovermw "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
)

type requestIDKey struct{}

const panicStackKey = "http.panic_stack"
const apiRouteKey = "http.api_route"

// requestID returns a server-generated ID whose storage is independent of Fiber buffers.
func requestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// RegisterDiagnostics surrounds recovery with access logging so recovered errors
// are rendered before their final status is logged.
func RegisterDiagnostics(app *fiber.App, lifetime context.Context) {
	app.Use(func(c *fiber.Ctx) error {
		ctx, cancel := context.WithCancel(c.UserContext())
		stop := context.AfterFunc(lifetime, cancel)
		defer stop()
		defer cancel()
		id := uuid.NewString()
		c.SetUserContext(context.WithValue(ctx, requestIDKey{}, id))
		c.Set("X-Request-ID", id)
		return c.Next()
	})
	app.Use(accessLog())
	app.Use(recovermw.New(recovermw.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, _ interface{}) {
			// Panic values can contain request data; record the stack without the value.
			c.Locals(panicStackKey, string(debug.Stack()))
		},
	}))
}

func apiDeadline(c *fiber.Ctx) error          { return withRequestDeadline(c, 30*time.Second) }
func dutyCreationDeadline(c *fiber.Ctx) error { return withRequestDeadline(c, 90*time.Second) }
func withRequestDeadline(c *fiber.Ctx, budget time.Duration) error {
	c.Locals(apiRouteKey, c.Route().Path)
	ctx, cancel := context.WithTimeout(c.UserContext(), budget)
	defer cancel()
	c.SetUserContext(ctx)
	err := c.Next()
	// Current API handlers buffer their responses. Preserve business failures,
	// but do not report success if cooperative work returned after cancellation.
	if err == nil && c.Response().StatusCode() < 400 && ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}
func isContextError(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}

// Infrastructure failures must bypass business response mappings, some of which
// intentionally expose validation messages. SQL/transport details stay server-side.
func isInfrastructureError(err error) bool {
	var sqlErr *mysqlDriver.MySQLError
	var networkErr net.Error
	return isContextError(err) || errors.As(err, &sqlErr) || errors.As(err, &networkErr) ||
		errors.Is(err, driver.ErrBadConn) || errors.Is(err, mysqlDriver.ErrInvalidConn) ||
		errors.Is(err, sql.ErrConnDone) || errors.Is(err, sql.ErrTxDone)
}

// ErrorHandler preserves the API error envelope without exposing internal errors.
func ErrorHandler(c *fiber.Ctx, err error) error {
	// fasthttp can reject an incomplete request before app middleware runs. Its
	// method/route metadata and handler duration may be unavailable. Do not
	// invent them or log the raw URL.
	if requestID(c.UserContext()) == "" {
		id := uuid.NewString()
		c.SetUserContext(context.WithValue(c.UserContext(), requestIDKey{}, id))
		c.Set("X-Request-ID", id)
		defer func() {
			level := slog.LevelWarn
			if c.Response().StatusCode() >= 500 {
				level = slog.LevelError
			}
			slog.Log(c.UserContext(), level, "http.protocol_error", "event", "http.protocol_error", "request_id", id, "method", "unknown", "route", "unmatched", "status", c.Response().StatusCode(), "duration_ms", nil)
		}()
	}
	code := fiber.StatusInternalServerError
	message := "внутренняя ошибка сервера"
	if c.Locals(panicStackKey) == nil {
		var httpErr *fiber.Error
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			code, message = fiber.StatusGatewayTimeout, "время выполнения операции истекло"
		case errors.Is(err, context.Canceled):
			code, message = fiber.StatusServiceUnavailable, "операция отменена"
		case errors.As(err, &httpErr):
			code = httpErr.Code
			if code < 500 {
				message = httpErr.Message
			}
		}
	}
	if !strings.HasPrefix(c.Path(), "/api/") && c.Locals(panicStackKey) == nil && !isContextError(err) {
		return fiber.DefaultErrorHandler(c, err)
	}
	return c.Status(code).JSON(errorResponse(message))
}

func accessLog() fiber.Handler {
	var healthMu sync.Mutex
	healthLast := make(map[string]time.Time)
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		operationErr := err
		if err != nil {
			err = c.App().ErrorHandler(c, err)
		}
		elapsed := time.Since(start)
		status := c.Response().StatusCode()
		route := "unmatched"
		if r := c.Route(); r != nil && r.Path != "/" {
			route = r.Path
		}
		api := strings.HasPrefix(c.Path(), "/api/")
		if api && status == 404 && c.Locals(apiRouteKey) == nil {
			route = "unmatched"
		}
		stack, panicOccurred := c.Locals(panicStackKey).(string)
		slow := elapsed >= 2*time.Second
		if !api && status < 500 && !slow && !panicOccurred {
			return err
		}
		if (route == "/health/live" || route == "/health/ready") && elapsed < 10*time.Second && !panicOccurred {
			healthMu.Lock()
			last := healthLast[route]
			emit := time.Since(last) >= time.Minute
			if emit {
				healthLast[route] = time.Now()
			}
			healthMu.Unlock()
			if !emit {
				return err
			}
		}
		level, event := slog.LevelInfo, "http.request"
		fields := []any{"event", event, "request_id", requestID(c.UserContext()), "method", c.Method(), "route", route, "status", status, "duration_ms", elapsed.Milliseconds()}
		if errors.Is(operationErr, context.DeadlineExceeded) {
			fields = append(fields, "error_kind", "deadline_exceeded")
		} else if errors.Is(operationErr, context.Canceled) {
			fields = append(fields, "error_kind", "canceled")
		} else if isInfrastructureError(operationErr) {
			fields = append(fields, "error_kind", "infrastructure_error")
		}
		if slow {
			level, event = slog.LevelWarn, "http.slow_request"
			fields = append(fields, "threshold_ms", 2000)
			if elapsed >= 10*time.Second {
				level = slog.LevelError
			}
		}
		if status >= 500 {
			level = slog.LevelError
		}
		if panicOccurred {
			level, event = slog.LevelError, "http.panic"
			fields = append(fields, "stack", stack)
		}
		fields[1] = event
		slog.Log(c.UserContext(), level, event, fields...)
		return err
	}
}
