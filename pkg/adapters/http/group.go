package http

import (
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	cleaninguc "dorm/pkg/core/usecase/cleaning"
	dutysettingsuc "dorm/pkg/core/usecase/dutysettings"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type GroupAPIHandler struct {
	dormitoryUC    ports.DormitoryUseCase
	cleaningUC     ports.CleaningUseCase
	dutySettingsUC ports.DutySettingsUseCase
}

type CreateGroupDutyRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func NewGroupAPIHandler(
	dormitoryUC ports.DormitoryUseCase,
	cleaningUC ports.CleaningUseCase,
	dutySettingsUC ports.DutySettingsUseCase,
) *GroupAPIHandler {
	return &GroupAPIHandler{
		dormitoryUC:    dormitoryUC,
		cleaningUC:     cleaningUC,
		dutySettingsUC: dutySettingsUC,
	}
}

func (h *GroupAPIHandler) RegisterRoutes(app *fiber.App, auth fiber.Handler) {
	api := app.Group("/api/v1/groups", auth)
	api.Get("", h.HandleGetGroups)
	api.Get("/options", h.HandleGetDormitoryUserOptions)
	api.Get("/:id", h.HandleGetGroup)
	api.Post("/:id/duties", h.HandleCreateGroupDuty)
	api.Get("/:id/duty-settings", h.HandleGetDutySettings)
	api.Get("/:id/duty-settings/teams/members", h.HandleGetDutySettingsTeamMemberOptions)
	api.Post("/:id/duty-settings/teams/reorder", h.HandleReorderDutySettingsTeams)
	api.Get("/:id/duty-settings/teams/:teamId", h.HandleGetDutySettingsTeam)
	api.Get("/:id/duty-settings/teams/:teamId/members", h.HandleGetDutySettingsTeamMembers)
	api.Get("/:id/duty-settings/teams/:teamId/member-search", h.HandleSearchDutySettingsTeamMembers)
	api.Post("/:id/duty-settings/teams/:teamId/members", h.HandleAddDutySettingsTeamMember)
	api.Put("/:id/duty-settings/teams/:teamId/leader", h.HandleAssignDutySettingsTeamLeader)
	api.Delete("/:id/duty-settings/teams/:teamId/members/:userId", h.HandleRemoveDutySettingsTeamMember)
	api.Post("/:id/duty-settings/teams", h.HandleCreateDutySettingsTeam)
	api.Put("/:id/duty-settings/teams/:teamId", h.HandleUpdateDutySettingsTeam)
	api.Delete("/:id/duty-settings/teams/:teamId", h.HandleDeleteDutySettingsTeam)
	api.Post("/:id/duty-settings/teams/:teamId/assign-active-duty", h.HandleAssignDutySettingsActiveTeam)
	api.Post("/:id/duty-settings/areas", h.HandleCreateDutySettingsArea)
	api.Put("/:id/duty-settings/areas/:areaId", h.HandleUpdateDutySettingsArea)
	api.Delete("/:id/duty-settings/areas/:areaId", h.HandleDeleteDutySettingsArea)
	api.Post("/:id/duty-settings/areas/:areaId/tasks", h.HandleCreateDutySettingsTask)
	api.Put("/:id/duty-settings/tasks/:taskId", h.HandleUpdateDutySettingsTask)
	api.Delete("/:id/duty-settings/tasks/:taskId", h.HandleDeleteDutySettingsTask)
	api.Post("/:id/duty-settings/tasks/:taskId/include", h.HandleIncludeDutySettingsTask)
	api.Delete("/:id/duty-settings/tasks/:taskId/include", h.HandleExcludeDutySettingsTask)
	api.Post("", h.HandleCreateGroup)
	api.Put("/:id", h.HandleUpdateGroup)
	api.Delete("/:id", h.HandleDeleteGroup)
}

func (h *GroupAPIHandler) HandleGetGroups(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	response, err := h.dormitoryUC.GetGroupsResponse(c.Context(), dormitoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить список групп"))
	}

	return c.JSON(response)
}

func (h *GroupAPIHandler) HandleGetDormitoryUserOptions(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	dormitoryID, err := strconv.ParseInt(c.Query("dormitory_id"), 10, 64)
	if err != nil || dormitoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор общежития"))
	}

	response, err := h.dormitoryUC.GetDormitoryUserOptionsResponse(c.Context(), dormitoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить список жителей"))
	}

	return c.JSON(response)
}

func (h *GroupAPIHandler) HandleGetGroup(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	group, err := h.dormitoryUC.GetGroupDetails(c.Context(), groupID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось загрузить группу"))
	}
	if group == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("группа не найдена"))
	}

	return c.JSON(group)
}

func (h *GroupAPIHandler) HandleCreateGroup(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	var req dto.CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	group, err := h.dormitoryUC.CreateGroupDetails(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(group)
}

func (h *GroupAPIHandler) HandleUpdateGroup(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	var req dto.UpdateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	group, err := h.dormitoryUC.UpdateGroupDetails(c.Context(), groupID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
	if group == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("группа не найдена"))
	}

	return c.JSON(group)
}

func (h *GroupAPIHandler) HandleDeleteGroup(c *fiber.Ctx) error {
	if err := h.requireDormitoryManagementAccess(c); err != nil {
		return err
	}

	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	if err := h.dormitoryUC.DeleteGroup(c.Context(), groupID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleCreateGroupDuty(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	var req CreateGroupDutyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	startDate, endDate, err := parseDutyPeriod(req.StartDate, req.EndDate)
	if err != nil {
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(errorResponse(fiberErr.Message))
		}

		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}

	if err := h.cleaningUC.StartNewDutyForGroup(c.Context(), userID, groupID, startDate, endDate); err != nil {
		return h.respondCreateGroupDutyError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleGetDutySettings(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	response, err := h.dutySettingsUC.GetDutySettings(c.Context(), userID, groupID)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.JSON(response)
}

func (h *GroupAPIHandler) HandleGetDutySettingsTeam(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	team, err := h.dutySettingsUC.GetTeamDetails(c.Context(), userID, groupID, teamID)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}
	if team == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("команда не найдена"))
	}

	return c.JSON(team)
}

func (h *GroupAPIHandler) HandleGetDutySettingsTeamMembers(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	response, err := h.dutySettingsUC.GetTeamMembers(c.Context(), userID, groupID, teamID)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.JSON(response)
}

func (h *GroupAPIHandler) HandleSearchDutySettingsTeamMembers(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	response, err := h.dutySettingsUC.SearchTeamMembers(c.Context(), userID, groupID, teamID, c.Query("q"))
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.JSON(response)
}

func (h *GroupAPIHandler) HandleGetDutySettingsTeamMemberOptions(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	teamID, err := parseOptionalUUIDQueryParam(c, "team_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	response, err := h.dutySettingsUC.GetTeamMemberOptions(c.Context(), userID, groupID, teamID)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.JSON(response)
}

func (h *GroupAPIHandler) HandleAddDutySettingsTeamMember(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	var req dto.DutySettingsTeamMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор жителя"))
	}

	if err := h.dutySettingsUC.AddTeamMember(c.Context(), userID, groupID, teamID, targetUserID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleAssignDutySettingsTeamLeader(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	var req dto.DutySettingsTeamLeaderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор жителя"))
	}

	if err := h.dutySettingsUC.AssignTeamLeader(c.Context(), userID, groupID, teamID, targetUserID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleRemoveDutySettingsTeamMember(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	targetUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор жителя"))
	}

	var req dto.DutySettingsRemoveTeamMemberRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
		}
	}
	var replacementLeaderID *uuid.UUID
	if req.ReplacementLeaderID != nil && *req.ReplacementLeaderID != "" {
		parsedReplacementLeaderID, err := uuid.Parse(*req.ReplacementLeaderID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор нового главы команды"))
		}
		replacementLeaderID = &parsedReplacementLeaderID
	}

	if err := h.dutySettingsUC.RemoveTeamMember(c.Context(), userID, groupID, teamID, targetUserID, replacementLeaderID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleCreateDutySettingsTeam(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	var req dto.CreateResidentTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	team, err := h.dutySettingsUC.CreateTeam(c.Context(), userID, groupID, req)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(team)
}

func (h *GroupAPIHandler) HandleUpdateDutySettingsTeam(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	var req dto.UpdateResidentTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	team, err := h.dutySettingsUC.UpdateTeam(c.Context(), userID, groupID, teamID, req)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}
	if team == nil {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("команда не найдена"))
	}

	return c.JSON(team)
}

func (h *GroupAPIHandler) HandleDeleteDutySettingsTeam(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	if err := h.dutySettingsUC.DeleteTeam(c.Context(), userID, groupID, teamID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleAssignDutySettingsActiveTeam(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	teamID, err := uuid.Parse(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
	}

	if err := h.dutySettingsUC.AssignActiveDutyTeam(c.Context(), userID, groupID, teamID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleReorderDutySettingsTeams(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	var req dto.ReorderDutySettingsTeamsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	teamIDs := make([]uuid.UUID, 0, len(req.TeamIDs))
	for _, rawID := range req.TeamIDs {
		teamID, err := uuid.Parse(rawID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор команды"))
		}
		teamIDs = append(teamIDs, teamID)
	}

	if err := h.dutySettingsUC.ReorderTeams(c.Context(), userID, groupID, teamIDs); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleCreateDutySettingsArea(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	var req dto.CreateAreaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	area, err := h.dutySettingsUC.CreateArea(c.Context(), userID, groupID, req)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(area)
}

func (h *GroupAPIHandler) HandleUpdateDutySettingsArea(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	areaID, err := strconv.Atoi(c.Params("areaId"))
	if err != nil || areaID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор территории"))
	}

	var req dto.UpdateAreaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	area, err := h.dutySettingsUC.UpdateArea(c.Context(), userID, groupID, areaID, req)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.JSON(area)
}

func (h *GroupAPIHandler) HandleDeleteDutySettingsArea(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	areaID, err := strconv.Atoi(c.Params("areaId"))
	if err != nil || areaID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор территории"))
	}

	if err := h.dutySettingsUC.DeleteArea(c.Context(), userID, groupID, areaID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleCreateDutySettingsTask(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	areaID, err := strconv.Atoi(c.Params("areaId"))
	if err != nil || areaID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор территории"))
	}

	var req dto.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	task, err := h.dutySettingsUC.CreateTask(c.Context(), userID, groupID, areaID, req)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

func (h *GroupAPIHandler) HandleUpdateDutySettingsTask(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	taskID, err := uuid.Parse(c.Params("taskId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор задачи"))
	}

	var req dto.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный формат запроса"))
	}

	task, err := h.dutySettingsUC.UpdateTask(c.Context(), userID, groupID, taskID, req)
	if err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.JSON(task)
}

func (h *GroupAPIHandler) HandleDeleteDutySettingsTask(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	taskID, err := uuid.Parse(c.Params("taskId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор задачи"))
	}

	if err := h.dutySettingsUC.DeleteTask(c.Context(), userID, groupID, taskID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleIncludeDutySettingsTask(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	taskID, err := uuid.Parse(c.Params("taskId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор задачи"))
	}

	if err := h.dutySettingsUC.IncludeTaskInActiveDuty(c.Context(), userID, groupID, taskID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) HandleExcludeDutySettingsTask(c *fiber.Ctx) error {
	userID, groupID, err := h.parseDutySettingsAccess(c)
	if err != nil {
		return err
	}

	taskID, err := uuid.Parse(c.Params("taskId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор задачи"))
	}

	if err := h.dutySettingsUC.ExcludeTaskFromActiveDuty(c.Context(), userID, groupID, taskID); err != nil {
		return h.respondDutySettingsError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupAPIHandler) parseDutySettingsAccess(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	userID, err := currentUserID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, c.Status(fiber.StatusBadRequest).JSON(errorResponse("некорректный идентификатор группы"))
	}

	return userID, groupID, nil
}

func (h *GroupAPIHandler) respondDutySettingsError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, dutysettingsuc.ErrAccessDenied):
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("доступ запрещен"))
	case errors.Is(err, structure.ErrReplacementLeaderRequired):
		return c.Status(fiber.StatusConflict).JSON(errorResponse("укажите нового главу команды"))
	case errors.Is(err, structure.ErrInvalidReplacementLeader):
		return c.Status(fiber.StatusConflict).JSON(errorResponse("новый глава должен быть другим участником этой команды"))
	case errors.Is(err, structure.ErrCannotRemoveOnlyLeader):
		return c.Status(fiber.StatusConflict).JSON(errorResponse("нельзя исключить единственного главу команды"))
	case err.Error() == "группа не найдена":
		return c.Status(fiber.StatusNotFound).JSON(errorResponse(err.Error()))
	case err.Error() == "территория не найдена":
		return c.Status(fiber.StatusNotFound).JSON(errorResponse(err.Error()))
	case err.Error() == "задача не найдена":
		return c.Status(fiber.StatusNotFound).JSON(errorResponse(err.Error()))
	case err.Error() == "В вашей группе пока нет дежурств":
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	case err.Error() == "В группе нет активного дежурства":
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	default:
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err.Error()))
	}
}

func (h *GroupAPIHandler) respondCreateGroupDutyError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, cleaninguc.ErrDutyCreationAccessDenied):
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("доступ запрещен"))
	case errors.Is(err, cleaninguc.ErrDutyGroupNotFound):
		return c.Status(fiber.StatusNotFound).JSON(errorResponse("группа не найдена"))
	case errors.Is(err, cleaninguc.ErrDutyAlreadyCreated):
		return c.Status(fiber.StatusConflict).JSON(errorResponse("Для группы уже было создано новое дежурство. Обновите страницу."))
	case errors.Is(err, duty.ErrDutyPeriodOverlap):
		return c.Status(fiber.StatusConflict).JSON(errorResponse("В выбранном периоде уже существует дежурство этой группы"))
	default:
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(errorResponse(fiberErr.Message))
		}

		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось создать дежурство"))
	}
}

func (h *GroupAPIHandler) requireDormitoryManagementAccess(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse("требуется авторизация"))
	}

	canManage, err := h.dormitoryUC.CanManageDormitories(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse("не удалось проверить доступ"))
	}
	if !canManage {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse("доступ запрещен"))
	}

	return nil
}
