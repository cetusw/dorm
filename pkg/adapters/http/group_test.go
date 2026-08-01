package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cleaninguc "dorm/pkg/core/usecase/cleaning"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type cleaningUseCaseStub struct {
	startNewDutyForGroupFunc func(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, startDate, endDate time.Time) error
}

func (s cleaningUseCaseStub) AssignTask(context.Context, uuid.UUID, uuid.UUID) error   { return nil }
func (s cleaningUseCaseStub) UnassignTask(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (s cleaningUseCaseStub) CompleteTask(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (s cleaningUseCaseStub) OpenTask(context.Context, uuid.UUID, uuid.UUID) error     { return nil }
func (s cleaningUseCaseStub) CancelCompletion(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (s cleaningUseCaseStub) VerifyTask(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (s cleaningUseCaseStub) ReopenTask(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (s cleaningUseCaseStub) StartNewWeek(context.Context) error                     { return nil }
func (s cleaningUseCaseStub) StartNewDutyForGroup(ctx context.Context, currentUserID uuid.UUID, groupID uuid.UUID, startDate, endDate time.Time) error {
	if s.startNewDutyForGroupFunc == nil {
		return nil
	}
	return s.startNewDutyForGroupFunc(ctx, currentUserID, groupID, startDate, endDate)
}

func TestHandleCreateGroupDuty_Unauthorized(t *testing.T) {
	app := fiber.New()
	handler := &GroupAPIHandler{cleaningUC: cleaningUseCaseStub{}}
	app.Post("/api/v1/groups/:id/duties", handler.HandleCreateGroupDuty)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/"+uuid.NewString()+"/duties", strings.NewReader(`{"start_date":"2026-07-29","end_date":"2026-08-05"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandleCreateGroupDuty_InvalidGroupID(t *testing.T) {
	app := fiber.New()
	handler := &GroupAPIHandler{cleaningUC: cleaningUseCaseStub{}}
	app.Post("/api/v1/groups/:id/duties", withUser(uuid.New(), handler.HandleCreateGroupDuty))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/invalid/duties", strings.NewReader(`{"start_date":"2026-07-29","end_date":"2026-08-05"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleCreateGroupDuty_InvalidPeriod(t *testing.T) {
	app := fiber.New()
	handler := &GroupAPIHandler{cleaningUC: cleaningUseCaseStub{}}
	app.Post("/api/v1/groups/:id/duties", withUser(uuid.New(), handler.HandleCreateGroupDuty))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/"+uuid.NewString()+"/duties", strings.NewReader(`{"start_date":"2026-08-05","end_date":"2026-07-29"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleCreateGroupDuty_Forbidden(t *testing.T) {
	groupID := uuid.New()
	app := fiber.New()
	handler := &GroupAPIHandler{
		cleaningUC: cleaningUseCaseStub{
			startNewDutyForGroupFunc: func(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) error {
				return cleaninguc.ErrDutyCreationAccessDenied
			},
		},
	}
	app.Post("/api/v1/groups/:id/duties", withUser(uuid.New(), handler.HandleCreateGroupDuty))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/"+groupID.String()+"/duties", strings.NewReader(`{"start_date":"2026-07-29","end_date":"2026-08-05"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestHandleCreateGroupDuty_Success(t *testing.T) {
	userID := uuid.New()
	groupID := uuid.New()
	app := fiber.New()
	called := false
	handler := &GroupAPIHandler{
		cleaningUC: cleaningUseCaseStub{
			startNewDutyForGroupFunc: func(_ context.Context, currentUserID uuid.UUID, currentGroupID uuid.UUID, startDate, endDate time.Time) error {
				called = true
				assert.Equal(t, userID, currentUserID)
				assert.Equal(t, groupID, currentGroupID)
				assert.Equal(t, time.Date(2026, 7, 29, 9, 0, 0, 0, time.Local), startDate)
				assert.Equal(t, time.Date(2026, 8, 5, 9, 0, 0, 0, time.Local), endDate)
				return nil
			},
		},
	}
	app.Post("/api/v1/groups/:id/duties", withUser(userID, handler.HandleCreateGroupDuty))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/"+groupID.String()+"/duties", strings.NewReader(`{"start_date":"2026-07-29","end_date":"2026-08-05"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.True(t, called)
}

func withUser(userID uuid.UUID, next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(currentUserIDLocalKey, userID)
		return next(c)
	}
}
