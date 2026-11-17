package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TeamAPIHandler struct {
	dormitoryUC ports.DormitoryUseCase
	teamUC      ports.TeamUseCase
}

func NewTeamAPIHandler(dormitoryUC ports.DormitoryUseCase, teamUC ports.TeamUseCase) *TeamAPIHandler {
	return &TeamAPIHandler{
		dormitoryUC: dormitoryUC,
		teamUC:      teamUC,
	}
}

func (h *TeamAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/teams", auth)
	api.Get("", apiDeadline, h.HandleGetTeams)
	api.Get("/members", apiDeadline, h.HandleGetTeamMemberOptions)
	api.Get("/:id", apiDeadline, h.HandleGetTeam)
	api.Post("", apiDeadline, h.HandleCreateTeam)
	api.Put("/:id", apiDeadline, h.HandleUpdateTeam)
	api.Delete("/:id", apiDeadline, h.HandleDeleteTeam)
}

func (h *TeamAPIHandler) HandleGetTeams(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	groupID, err := parseUUIDQueryParam(c, "group_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	response, err := h.teamUC.GetTeamsResponseByGroup(c.UserContext(), groupID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.JSON(response)
}

func (h *TeamAPIHandler) HandleGetTeamMemberOptions(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	groupID, err := parseUUIDQueryParam(c, "group_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	teamID, err := parseOptionalUUIDQueryParam(c, "team_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	response, err := h.teamUC.GetTeamMemberOptionsResponse(c.UserContext(), groupID, teamID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.JSON(response)
}

func (h *TeamAPIHandler) HandleGetTeam(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	team, err := h.teamUC.GetTeamDetails(c.UserContext(), teamID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
	if team == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("команда не найдена"))
	}

	return c.JSON(team)
}

func (h *TeamAPIHandler) HandleCreateTeam(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	var req dto.CreateResidentTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	team, err := h.teamUC.CreateResidentTeam(c.UserContext(), req)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(team)
}

func (h *TeamAPIHandler) HandleUpdateTeam(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	var req dto.UpdateResidentTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	team, err := h.teamUC.UpdateResidentTeam(c.UserContext(), teamID, req)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
	if team == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("команда не найдена"))
	}

	return c.JSON(team)
}

func (h *TeamAPIHandler) HandleDeleteTeam(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	if err := h.teamUC.DeleteTeam(c.UserContext(), teamID); err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *TeamAPIHandler) requireDormitoryManagementAccess(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "требуется авторизация")
	}

	canManage, err := h.dormitoryUC.CanManageDormitories(c.UserContext(), userID)
	if err != nil {
		if isInfrastructureError(err) {
			return err
		}
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось проверить доступ")
	}
	if !canManage {
		return fiber.NewError(fiber.StatusForbidden, "доступ запрещен")
	}

	return nil
}

func parseUUIDQueryParam(c *fiber.Ctx, key string) (uuid.UUID, error) {
	return uuid.Parse(c.Query(key))
}

func parseOptionalUUIDQueryParam(c *fiber.Ctx, key string) (*uuid.UUID, error) {
	rawValue := c.Query(key)
	if rawValue == "" {
		return nil, nil
	}

	value, err := uuid.Parse(rawValue)
	if err != nil {
		return nil, err
	}

	return &value, nil
}
