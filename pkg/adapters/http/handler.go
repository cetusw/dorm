package http

import (
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AdminHandler struct {
	userUC      ports.UserUseCase
	dormitoryUC ports.DormitoryUseCase
	teamUC      ports.TeamUseCase
	dutyUC      ports.DutyUseCase
	taskUC      ports.TaskCatalogUseCase
	cleaningUC  ports.CleaningUseCase
	sheetsUC    ports.SheetsUseCase
}

type dutyGroupSettingsItem struct {
	Group        dto.GroupListItem
	Tasks        []dto.FutureDutyTaskGroup
	Teams        []dto.TeamListItem
	NextDutyTeam string
}

func NewAdminHandler(
	userUC ports.UserUseCase,
	dormitoryUC ports.DormitoryUseCase,
	teamUC ports.TeamUseCase,
	dutyUC ports.DutyUseCase,
	taskUC ports.TaskCatalogUseCase,
	cleaningUC ports.CleaningUseCase,
	sheetsUC ports.SheetsUseCase,
) *AdminHandler {
	return &AdminHandler{
		userUC:      userUC,
		dormitoryUC: dormitoryUC,
		teamUC:      teamUC,
		dutyUC:      dutyUC,
		taskUC:      taskUC,
		cleaningUC:  cleaningUC,
		sheetsUC:    sheetsUC,
	}
}

func (h *AdminHandler) RegisterRoutes(app *fiber.App) {
	admin := app.Group("/admin")

	admin.Get("/", h.HandleIndex)
	admin.Get("/users", h.HandleGetUsers)
	admin.Get("/users/create", h.HandleCreateUserModal)
	admin.Post("/users", h.HandleStoreUser)
	admin.Get("/users/:id/edit", h.HandleEditUserModal)
	admin.Put("/users/:id", h.HandleUpdateUser)
	admin.Delete("/users/:id", h.HandleDeleteUser)

	admin.Get("/dormitories", h.HandleGetDormitories)
	admin.Get("/dormitories/create", h.HandleCreateDormitoryModal)
	admin.Post("/dormitories", h.HandleStoreDormitory)
	admin.Get("/dormitories/:dormId/edit", h.HandleEditDormitoryModal)
	admin.Put("/dormitories/:dormId", h.HandleUpdateDormitory)
	admin.Delete("/dormitories/:dormId", h.HandleDeleteDormitory)

	admin.Get("/dormitories/:dormId/groups", h.HandleGetGroups)
	admin.Get("/dormitories/:dormIresidentDutyUCd/groups/create", h.HandleCreateGroupModal)
	admin.Post("/dormitories/:dormId/groups", h.HandleStoreGroup)
	admin.Get("/dormitories/:dormId/groups/:groupId/edit", h.HandleEditGroupModal)
	admin.Put("/dormitories/:dormId/groups/:groupId", h.HandleUpdateGroup)
	admin.Delete("/dormitories/:dormId/groups/:groupId", h.HandleDeleteGroup)

	admin.Get("/dormitories/:dormId/groups/:groupId/teams", h.HandleGetDormitoryGroupTeams)
	admin.Get("/dormitories/:dormId/groups/:groupId/teams/create", h.HandleCreateTeamModal)
	admin.Post("/dormitories/:dormId/groups/:groupId/teams", h.HandleStoreTeam)
	admin.Get("/dormitories/:dormId/groups/:groupId/teams/:teamId/edit", h.HandleEditTeamModal)
	admin.Put("/dormitories/:dormId/groups/:groupId/teams/:teamId", h.HandleUpdateTeam)
	admin.Delete("/dormitories/:dormId/groups/:groupId/teams/:teamId", h.HandleDeleteTeam)
	admin.Put("/dormitories/:dormId/groups/:groupId/teams/:teamId/members/:userID", h.HandleAddTeamMember)
	admin.Delete("/dormitories/:dormId/groups/:groupId/teams/members/:userID", h.HandleRemoveTeamMember)

	admin.Get("/dormitories/:dormId/groups/:groupId/duties", h.HandleGetGroupDuties)
	admin.Get("/dormitories/:dormId/groups/:groupId/duties/future", h.HandleFutureDutyModal)
	admin.Put("/dormitories/:dormId/groups/:groupId/duties/future", h.HandleUpdateFutureDuty)
	admin.Get("/dormitories/:dormId/groups/:groupId/duties/:dutyId", h.HandleGetDutyDetail)

	admin.Get("/duties", h.HandleGetDutyDormitories)
	admin.Get("/duties/:dormId/groups", h.HandleGetDutyGroups)
	admin.Put("/duties/:dormId/settings", h.HandleUpdateDormitoryDutySettings)
	admin.Get("/duties/:dormId/create", h.HandleCreateDormitoryDutiesModal)
	admin.Post("/duties/:dormId", h.HandleStoreDormitoryDuties)
	admin.Get("/duties/:dormId/common-settings", h.HandleCommonDutySettingsModal)
	admin.Put("/duties/:dormId/common-settings", h.HandleUpdateCommonDutySettings)

	admin.Get("/tasks", h.HandleGetTasks)
	admin.Get("/tasks/:dormId/groups", h.HandleGetTaskGroups)
	admin.Get("/tasks/:dormId/groups/common", h.HandleGetCommonTasks)
	admin.Get("/tasks/:dormId/groups/common/create", h.HandleCreateCommonTaskModal)
	admin.Post("/tasks/:dormId/groups/common", h.HandleStoreCommonTask)
	admin.Get("/tasks/:dormId/groups/common/items/:id/edit", h.HandleEditCommonTaskModal)
	admin.Put("/tasks/:dormId/groups/common/items/:id", h.HandleUpdateCommonTask)
	admin.Delete("/tasks/:dormId/groups/common/items/:id", h.HandleDeleteCommonTask)
	admin.Get("/tasks/:dormId/groups/:groupId", h.HandleGetGroupTasks)
	admin.Get("/tasks/:dormId/groups/:groupId/create", h.HandleCreateGroupTaskModal)
	admin.Post("/tasks/:dormId/groups/:groupId", h.HandleStoreGroupTask)
	admin.Get("/tasks/:dormId/groups/:groupId/items/:id/edit", h.HandleEditGroupTaskModal)
	admin.Put("/tasks/:dormId/groups/:groupId/items/:id", h.HandleUpdateGroupTask)
	admin.Delete("/tasks/:dormId/groups/:groupId/items/:id", h.HandleDeleteGroupTask)
	admin.Get("/tasks/:id/edit", h.HandleEditTaskModal)
	admin.Put("/tasks/:id", h.HandleUpdateTask)
	admin.Delete("/tasks/:id", h.HandleDeleteTask)
	admin.Post("/sheets/regenerate", h.HandleRegenerateSheet)
}

func (h *AdminHandler) HandleIndex(c *fiber.Ctx) error {
	return c.Redirect("/admin/users", fiber.StatusFound)
}

func (h *AdminHandler) HandleGetUsers(c *fiber.Ctx) error {
	users, err := h.userUC.GetUsersList(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении списка пользователей")
	}

	return c.Render("users", fiber.Map{
		"Users":         users,
		"ActiveSection": "users",
	}, "layouts/main")
}

func (h *AdminHandler) HandleCreateUserModal(c *fiber.Ctx) error {
	dorms, err := h.dormitoryUC.GetDormitories(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении общежитий")
	}
	return c.Render("partials/modal_create_user", fiber.Map{
		"Dormitories": dorms,
	})
}

func (h *AdminHandler) HandleStoreUser(c *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}

	if err := h.userUC.CreateUser(c.Context(), req); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	c.Response().Header.Set("HX-Redirect", "/admin/users")
	return c.SendString("")
}

func (h *AdminHandler) HandleEditUserModal(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid user id")
	}
	u, err := h.userUC.GetUserByID(c.Context(), id)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении пользователя")
	}
	if u == nil {
		return c.Status(404).SendString("User not found")
	}

	var dormID int64
	room := ""
	if u.DormitoryID() != nil {
		dormID = *u.DormitoryID()
	}
	if u.RoomNumber() != nil {
		room = *u.RoomNumber()
	}
	middleName := ""
	if u.MiddleName() != nil {
		middleName = *u.MiddleName()
	}
	floor := 0
	if u.FloorNumber() != nil {
		floor = *u.FloorNumber()
	}

	dorms, err := h.dormitoryUC.GetDormitories(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении общежитий")
	}

	return c.Render("partials/modal_edit_user", fiber.Map{
		"User": fiber.Map{
			"ID":          u.ID(),
			"FirstName":   u.FirstName(),
			"MiddleName":  middleName,
			"LastName":    u.LastName(),
			"RoomNumber":  room,
			"Floor":       floor,
			"DormitoryID": dormID,
		},
		"Dormitories": dorms,
	})
}

func (h *AdminHandler) HandleUpdateUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid user id")
	}

	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}

	if err := h.userUC.UpdateUser(c.Context(), id, req); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	c.Response().Header.Set("HX-Redirect", "/admin/users")
	return c.SendString("")
}

func (h *AdminHandler) HandleDeleteUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid user id")
	}
	if err := h.userUC.SoftDeleteUser(c.Context(), id); err != nil {
		return c.Status(500).SendString("Ошибка при удалении пользователя")
	}
	return c.SendString("")
}

func (h *AdminHandler) HandleGetDormitories(c *fiber.Ctx) error {
	dormitories, err := h.dormitoryUC.GetDormitoriesList(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении общежитий")
	}
	return c.Render("dormitories", fiber.Map{
		"Dormitories":   dormitories,
		"ActiveSection": "dormitories",
	}, "layouts/main")
}

func (h *AdminHandler) HandleCreateDormitoryModal(c *fiber.Ctx) error {
	users, err := h.dormitoryUC.GetUserOptions(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении пользователей")
	}
	return c.Render("partials/modal_create_dormitory", fiber.Map{"Users": users})
}

func (h *AdminHandler) HandleStoreDormitory(c *fiber.Ctx) error {
	var req dto.UpsertDormitoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.dormitoryUC.CreateDormitory(c.Context(), req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", "/admin/dormitories")
	return c.SendString("")
}

func (h *AdminHandler) HandleEditDormitoryModal(c *fiber.Ctx) error {
	dormID, err := parseIntParam(c, "dormId")
	if err != nil {
		return c.Status(400).SendString("Invalid dormitory id")
	}
	dormitory, err := h.dormitoryUC.GetDormitoryByID(c.Context(), dormID)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении общежития")
	}
	if dormitory == nil {
		return c.Status(404).SendString("Dormitory not found")
	}
	users, err := h.dormitoryUC.GetUserOptions(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении пользователей")
	}
	leaderID := uuid.Nil
	if dormitory.LeaderID() != nil {
		leaderID = *dormitory.LeaderID()
	}
	return c.Render("partials/modal_edit_dormitory", fiber.Map{
		"Dormitory": dormitory,
		"LeaderID":  leaderID,
		"Users":     users,
	})
}

func (h *AdminHandler) HandleUpdateDormitory(c *fiber.Ctx) error {
	dormID, err := parseIntParam(c, "dormId")
	if err != nil {
		return c.Status(400).SendString("Invalid dormitory id")
	}
	var req dto.UpsertDormitoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.dormitoryUC.UpdateDormitory(c.Context(), dormID, req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", "/admin/dormitories")
	return c.SendString("")
}

func (h *AdminHandler) HandleDeleteDormitory(c *fiber.Ctx) error {
	dormID, err := parseIntParam(c, "dormId")
	if err != nil {
		return c.Status(400).SendString("Invalid dormitory id")
	}
	if err := h.dormitoryUC.DeleteDormitory(c.Context(), dormID); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", "/admin/dormitories")
	return c.SendString("")
}

func (h *AdminHandler) HandleGetGroups(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	groups, err := h.dormitoryUC.GetGroupsList(c.Context(), dormitory.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении групп")
	}
	return c.Render("dormitory_groups", fiber.Map{
		"Dormitory":     dormitory,
		"Groups":        groups,
		"ActiveSection": "dormitories",
	}, "layouts/main")
}

func (h *AdminHandler) HandleCreateGroupModal(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	residents, err := h.dormitoryUC.GetDormitoryUserOptions(c.Context(), dormitory.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении жителей")
	}
	return c.Render("partials/modal_create_group", fiber.Map{
		"Dormitory": dormitory,
		"Residents": residents,
	})
}

func (h *AdminHandler) HandleStoreGroup(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	var req dto.UpsertGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.dormitoryUC.CreateGroup(c.Context(), dormitory.ID(), req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", groupsURL(dormitory.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleEditGroupModal(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	residents, err := h.dormitoryUC.GetDormitoryUserOptions(c.Context(), dormitory.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении жителей")
	}
	leaderID := uuid.Nil
	if group.LeaderID() != nil {
		leaderID = *group.LeaderID()
	}
	return c.Render("partials/modal_edit_group", fiber.Map{
		"Dormitory": dormitory,
		"Group":     group,
		"LeaderID":  leaderID,
		"Residents": residents,
	})
}

func (h *AdminHandler) HandleUpdateGroup(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	var req dto.UpsertGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.dormitoryUC.UpdateGroup(c.Context(), group.ID(), req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", groupsURL(dormitory.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleDeleteGroup(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	if err := h.dormitoryUC.DeleteGroup(c.Context(), group.ID()); err != nil {
		return c.Status(500).SendString("Ошибка при удалении группы")
	}
	c.Response().Header.Set("HX-Redirect", groupsURL(dormitory.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleGetDormitoryGroupTeams(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	teams, err := h.teamUC.GetTeamsListByGroup(c.Context(), group.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении команд")
	}
	return c.Render("dormitory_group_teams", fiber.Map{
		"Dormitory":     dormitory,
		"Group":         group,
		"Teams":         teams,
		"ActiveSection": "dormitories",
	}, "layouts/main")
}

func (h *AdminHandler) HandleCreateTeamModal(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	members, err := h.teamUC.GetTeamMembersForNewTeam(c.Context(), group.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении жителей")
	}
	residents, err := h.dormitoryUC.GetDormitoryUserOptions(c.Context(), dormitory.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении жителей")
	}
	return c.Render("partials/modal_create_team", fiber.Map{
		"Dormitory": dormitory,
		"Group":     group,
		"Members":   members,
		"Residents": residents,
	})
}

func (h *AdminHandler) HandleStoreTeam(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	var req dto.CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	req.MemberIDs = formValues(c, "member_ids")
	if err := h.teamUC.CreateTeamInGroup(c.Context(), group.ID(), req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", teamsURL(dormitory.ID(), group.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleEditTeamModal(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	teamID, err := parseUUIDParam(c, "teamId")
	if err != nil {
		return c.Status(400).SendString("Invalid team id")
	}

	teamItem, err := h.teamUC.GetTeamByID(c.Context(), teamID)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении команды")
	}
	if teamItem == nil {
		return c.Status(404).SendString("Team not found")
	}
	if teamItem.GroupID != group.ID() {
		return c.Status(404).SendString("Team not found")
	}

	members, err := h.teamUC.GetTeamMembersForEdit(c.Context(), teamID)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении участников команды")
	}
	residents, err := h.dormitoryUC.GetDormitoryUserOptions(c.Context(), dormitory.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении жителей")
	}
	leaderID := uuid.Nil
	if teamItem.LeaderID != nil {
		leaderID = *teamItem.LeaderID
	}
	return c.Render("partials/modal_edit_team", fiber.Map{
		"Dormitory": dormitory,
		"Group":     group,
		"Team":      teamItem,
		"Members":   members,
		"LeaderID":  leaderID,
		"Residents": residents,
	})
}

func (h *AdminHandler) HandleUpdateTeam(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	teamID, err := parseUUIDParam(c, "teamId")
	if err != nil {
		return c.Status(400).SendString("Invalid team id")
	}
	if err := h.requireTeamInGroup(c, teamID, group.ID()); err != nil {
		return err
	}

	var req dto.UpdateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	req.MemberIDs = formValues(c, "member_ids")
	if err := h.teamUC.UpdateTeamInGroup(c.Context(), group.ID(), teamID, req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", teamsURL(dormitory.ID(), group.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleDeleteTeam(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	teamID, err := parseUUIDParam(c, "teamId")
	if err != nil {
		return c.Status(400).SendString("Invalid team id")
	}
	if err := h.requireTeamInGroup(c, teamID, group.ID()); err != nil {
		return err
	}
	if err := h.teamUC.DeleteTeam(c.Context(), teamID); err != nil {
		return c.Status(500).SendString("Ошибка при удалении команды")
	}
	c.Response().Header.Set("HX-Redirect", teamsURL(dormitory.ID(), group.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleAddTeamMember(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	teamID, err := parseUUIDParam(c, "teamId")
	if err != nil {
		return c.Status(400).SendString("Invalid team id")
	}
	if err := h.requireTeamInGroup(c, teamID, group.ID()); err != nil {
		return err
	}
	userID, err := parseUUIDParam(c, "userID")
	if err != nil {
		return c.Status(400).SendString("Invalid user id")
	}
	if err := h.teamUC.MoveUserToTeam(c.Context(), teamID, userID); err != nil {
		return c.Status(500).SendString("Ошибка при добавлении участника")
	}
	c.Response().Header.Set("HX-Redirect", teamsURL(dormitory.ID(), group.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleRemoveTeamMember(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	userID, err := parseUUIDParam(c, "userID")
	if err != nil {
		return c.Status(400).SendString("Invalid user id")
	}
	if err := h.requireMemberFromGroup(c, userID, group.ID()); err != nil {
		return err
	}
	if err := h.teamUC.RemoveUserFromTeam(c.Context(), userID); err != nil {
		return c.Status(500).SendString("Ошибка при удалении участника")
	}
	c.Response().Header.Set("HX-Redirect", teamsURL(dormitory.ID(), group.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleGetGroupDuties(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	duties, err := h.dutyUC.GetGroupDuties(c.Context(), group.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении дежурств")
	}
	return c.Render("dormitory_group_duties", fiber.Map{
		"Dormitory":     dormitory,
		"Group":         group,
		"Duties":        duties,
		"ActiveSection": "duties",
	}, "layouts/main")
}

func (h *AdminHandler) HandleFutureDutyModal(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	tasks, err := h.dutyUC.GetFutureDutyTasks(c.Context(), group.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении задач")
	}
	teams, err := h.teamUC.GetTeamsListByGroup(c.Context(), group.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении команд")
	}
	nextDutyTeam := ""
	if group.NextDutyTeam() != nil {
		nextDutyTeam = strconv.Itoa(*group.NextDutyTeam())
	}
	return c.Render("partials/modal_future_duty", fiber.Map{
		"Dormitory":    dormitory,
		"Group":        group,
		"Tasks":        tasks,
		"Teams":        teams,
		"NextDutyTeam": nextDutyTeam,
	})
}

func (h *AdminHandler) HandleUpdateFutureDuty(c *fiber.Ctx) error {
	_, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	taskIDs, err := parseUUIDValues(formValues(c, "task_ids"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	nextDutyTeam, err := parseOptionalInt(c.FormValue("next_duty_team"))
	if err != nil {
		return c.Status(400).SendString("Invalid next duty team")
	}
	if err := h.dutyUC.UpdateGroupDutySettings(c.Context(), group.ID(), nextDutyTeam, taskIDs); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Trigger", "closeModal")
	return c.SendString("")
}

func (h *AdminHandler) HandleGetDutyDetail(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	dutyID, err := parseUUIDParam(c, "dutyId")
	if err != nil {
		return c.Status(400).SendString("Invalid duty id")
	}
	dutyItem, err := h.dutyUC.GetDutyDetail(c.Context(), group.ID(), dutyID)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении дежурства")
	}
	if dutyItem == nil {
		return c.Status(404).SendString("Duty not found")
	}
	return c.Render("duty_detail", fiber.Map{
		"Dormitory":     dormitory,
		"Group":         group,
		"Duty":          dutyItem,
		"ActiveSection": "duties",
	}, "layouts/main")
}

func (h *AdminHandler) HandleGetDutyDormitories(c *fiber.Ctx) error {
	dormitories, err := h.dormitoryUC.GetDormitoriesList(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении общежитий")
	}
	return c.Render("duties", fiber.Map{
		"Dormitories":   dormitories,
		"ActiveSection": "duties",
	}, "layouts/main")
}

func (h *AdminHandler) HandleGetDutyGroups(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	groups, err := h.dormitoryUC.GetGroupsList(c.Context(), dormitory.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении групп")
	}
	commonTasks, err := h.dutyUC.GetCommonFutureDutyTasks(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении общих задач")
	}
	groupSettings := make([]dutyGroupSettingsItem, 0, len(groups))
	for _, groupItem := range groups {
		group, err := h.dormitoryUC.GetGroupByID(c.Context(), groupItem.ID)
		if err != nil {
			return c.Status(500).SendString("Ошибка при получении группы")
		}
		if group == nil || group.DormitoryID() != dormitory.ID() {
			continue
		}
		tasks, err := h.dutyUC.GetFutureDutyTasks(c.Context(), group.ID())
		if err != nil {
			return c.Status(500).SendString("Ошибка при получении задач")
		}
		teams, err := h.teamUC.GetTeamsListByGroup(c.Context(), group.ID())
		if err != nil {
			return c.Status(500).SendString("Ошибка при получении команд")
		}
		nextDutyTeam := ""
		if group.NextDutyTeam() != nil {
			nextDutyTeam = strconv.Itoa(*group.NextDutyTeam())
		}
		groupSettings = append(groupSettings, dutyGroupSettingsItem{
			Group:        groupItem,
			Tasks:        tasks,
			Teams:        teams,
			NextDutyTeam: nextDutyTeam,
		})
	}
	return c.Render("duties_groups", fiber.Map{
		"Dormitory":     dormitory,
		"CommonTasks":   commonTasks,
		"GroupSettings": groupSettings,
		"ActiveSection": "duties",
	}, "layouts/main")
}

func (h *AdminHandler) HandleUpdateDormitoryDutySettings(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	groups, err := h.dormitoryUC.GetGroupsList(c.Context(), dormitory.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении групп")
	}
	commonTaskIDs, err := parseUUIDValues(formValues(c, "common_task_ids"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	req := dto.UpdateDormitoryDutySettingsRequest{CommonTaskIDs: commonTaskIDs}
	for _, group := range groups {
		taskIDs, err := parseUUIDValues(formValues(c, fmt.Sprintf("group_task_ids_%s", group.ID)))
		if err != nil {
			return c.Status(400).SendString("Invalid task id")
		}
		nextDutyTeam, err := parseOptionalInt(c.FormValue(fmt.Sprintf("next_duty_team_%s", group.ID)))
		if err != nil {
			return c.Status(400).SendString("Invalid next duty team")
		}
		req.Groups = append(req.Groups, dto.DutyGroupSettingsUpdate{
			GroupID:        group.ID,
			NextDutyTeam:   nextDutyTeam,
			IncludeTaskIDs: taskIDs,
		})
	}
	if err := h.dutyUC.UpdateDormitoryDutySettings(c.Context(), dormitory.ID(), req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", fmt.Sprintf("/admin/duties/%d/groups", dormitory.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleCreateDormitoryDutiesModal(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	return c.Render("partials/modal_create_dormitory_duties", fiber.Map{
		"Dormitory": dormitory,
	})
}

func (h *AdminHandler) HandleStoreDormitoryDuties(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}

	startDate, err := parseDateFormValue(c, "start_date")
	if err != nil {
		return c.Status(400).SendString("Invalid start date")
	}
	endDate, err := parseDateFormValue(c, "end_date")
	if err != nil {
		return c.Status(400).SendString("Invalid end date")
	}
	if err := h.cleaningUC.StartNewDutiesForDormitory(c.Context(), dormitory.ID(), startDate, endDate); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", fmt.Sprintf("/admin/duties/%d/groups", dormitory.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleCommonDutySettingsModal(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	tasks, err := h.dutyUC.GetCommonFutureDutyTasks(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении общих задач")
	}
	return c.Render("partials/modal_common_duty_settings", fiber.Map{
		"Dormitory": dormitory,
		"Tasks":     tasks,
	})
}

func (h *AdminHandler) HandleUpdateCommonDutySettings(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	taskIDs, err := parseUUIDValues(formValues(c, "common_task_ids"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	if err := h.dutyUC.UpdateCommonDutySettings(c.Context(), taskIDs); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", fmt.Sprintf("/admin/duties/%d/groups", dormitory.ID()))
	return c.SendString("")
}

func (h *AdminHandler) requireDormitory(c *fiber.Ctx) (*structure.Dormitory, error) {
	dormID, err := parseIntParam(c, "dormId")
	if err != nil {
		return nil, c.Status(400).SendString("Invalid dormitory id")
	}
	dormitory, err := h.dormitoryUC.GetDormitoryByID(c.Context(), dormID)
	if err != nil {
		return nil, c.Status(500).SendString("Ошибка при получении общежития")
	}
	if dormitory == nil {
		return nil, c.Status(404).SendString("Dormitory not found")
	}
	return dormitory, nil
}

func (h *AdminHandler) requireDormitoryGroup(c *fiber.Ctx) (*structure.Dormitory, *structure.Group, error) {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return nil, nil, err
	}
	groupID, err := parseUUIDParam(c, "groupId")
	if err != nil {
		return nil, nil, c.Status(400).SendString("Invalid group id")
	}
	group, err := h.dormitoryUC.GetGroupByID(c.Context(), groupID)
	if err != nil {
		return nil, nil, c.Status(500).SendString("Ошибка при получении группы")
	}
	if group == nil || group.DormitoryID() != dormitory.ID() {
		return nil, nil, c.Status(404).SendString("Group not found")
	}
	return dormitory, group, nil
}

func (h *AdminHandler) requireTeamInGroup(c *fiber.Ctx, teamID uuid.UUID, groupID uuid.UUID) error {
	team, err := h.teamUC.GetTeamByID(c.Context(), teamID)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении команды")
	}
	if team == nil || team.GroupID != groupID {
		return c.Status(404).SendString("Team not found")
	}
	return nil
}

func (h *AdminHandler) requireMemberFromGroup(c *fiber.Ctx, userID uuid.UUID, groupID uuid.UUID) error {
	resident, err := h.userUC.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении пользователя")
	}
	if resident == nil || resident.TeamID() == nil {
		return c.Status(404).SendString("Member not found")
	}
	team, err := h.teamUC.GetTeamByID(c.Context(), *resident.TeamID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении команды")
	}
	if team == nil || team.GroupID != groupID {
		return c.Status(404).SendString("Member not found")
	}
	return nil
}

func parseIntParam(c *fiber.Ctx, name string) (int64, error) {
	return strconv.ParseInt(c.Params(name), 10, 64)
}

func parseUUIDParam(c *fiber.Ctx, name string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(name))
}

func parseDateFormValue(c *fiber.Ctx, name string) (time.Time, error) {
	return time.Parse("02.01.2006", c.FormValue(name))
}

func parseOptionalInt(rawValue string) (*int, error) {
	if rawValue == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(rawValue)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func groupsURL(dormID int64) string {
	return fmt.Sprintf("/admin/dormitories/%d/groups", dormID)
}

func teamsURL(dormID int64, groupID uuid.UUID) string {
	return fmt.Sprintf("/admin/dormitories/%d/groups/%s/teams", dormID, groupID)
}

func dutiesURL(dormID int64, groupID uuid.UUID) string {
	return fmt.Sprintf("/admin/dormitories/%d/groups/%s/duties", dormID, groupID)
}

func taskGroupURL(dormID int64, groupID uuid.UUID) string {
	return fmt.Sprintf("/admin/tasks/%d/groups/%s", dormID, groupID)
}

func commonTaskURL(dormID int64) string {
	return fmt.Sprintf("/admin/tasks/%d/groups/common", dormID)
}

func formValues(c *fiber.Ctx, key string) []string {
	values := make([]string, 0)
	c.Request().PostArgs().VisitAll(func(argKey []byte, value []byte) {
		if string(argKey) == key {
			values = append(values, string(value))
		}
	})
	return values
}

func parseUUIDValues(values []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (h *AdminHandler) HandleGetTasks(c *fiber.Ctx) error {
	dormitories, err := h.dormitoryUC.GetDormitoriesList(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении общежитий")
	}
	return c.Render("tasks", fiber.Map{
		"Dormitories":   dormitories,
		"ActiveSection": "tasks",
	}, "layouts/main")
}

func (h *AdminHandler) HandleGetTaskGroups(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	groups, err := h.dormitoryUC.GetGroupsList(c.Context(), dormitory.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении групп")
	}
	return c.Render("task_groups", fiber.Map{
		"Dormitory":     dormitory,
		"Groups":        groups,
		"ActiveSection": "tasks",
	}, "layouts/main")
}

func (h *AdminHandler) HandleGetCommonTasks(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	taskGroups, err := h.taskUC.ListCommonTaskGroups(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении задач")
	}
	return c.Render("common_tasks", fiber.Map{
		"Dormitory":     dormitory,
		"TaskGroups":    taskGroups,
		"ActiveSection": "tasks",
	}, "layouts/main")
}

func (h *AdminHandler) HandleGetGroupTasks(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	taskGroups, err := h.taskUC.ListTaskGroupsByGroup(c.Context(), group.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении задач")
	}
	return c.Render("group_tasks", fiber.Map{
		"Dormitory":     dormitory,
		"Group":         group,
		"TaskGroups":    taskGroups,
		"ActiveSection": "tasks",
	}, "layouts/main")
}

func (h *AdminHandler) HandleCreateCommonTaskModal(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	return c.Render("partials/modal_create_task", fiber.Map{
		"Dormitory": dormitory,
		"Common":    true,
	})
}

func (h *AdminHandler) HandleCreateGroupTaskModal(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	areas, err := h.taskUC.ListAreasByGroup(c.Context(), group.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении зон")
	}
	return c.Render("partials/modal_create_task", fiber.Map{
		"Dormitory": dormitory,
		"Group":     group,
		"Areas":     areas,
	})
}

func (h *AdminHandler) HandleStoreCommonTask(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	var req dto.UpsertTaskCatalogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.taskUC.CreateCommonTask(c.Context(), req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", commonTaskURL(dormitory.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleStoreGroupTask(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	var req dto.UpsertTaskCatalogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.taskUC.CreateTaskInGroup(c.Context(), group.ID(), req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", taskGroupURL(dormitory.ID(), group.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleEditCommonTaskModal(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	task, err := h.taskUC.GetTask(c.Context(), id)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении задачи")
	}
	if task == nil {
		return c.Status(404).SendString("Task not found")
	}
	return c.Render("partials/modal_edit_task", fiber.Map{
		"Dormitory": dormitory,
		"Common":    true,
		"Task":      task,
	})
}

func (h *AdminHandler) HandleEditGroupTaskModal(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	task, err := h.taskUC.GetTask(c.Context(), id)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении задачи")
	}
	if task == nil {
		return c.Status(404).SendString("Task not found")
	}
	areas, err := h.taskUC.ListAreasByGroup(c.Context(), group.ID())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении зон")
	}
	return c.Render("partials/modal_edit_task", fiber.Map{
		"Dormitory": dormitory,
		"Group":     group,
		"Task":      task,
		"Areas":     areas,
	})
}

func (h *AdminHandler) HandleUpdateCommonTask(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	var req dto.UpsertTaskCatalogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.taskUC.UpdateCommonTask(c.Context(), id, req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", commonTaskURL(dormitory.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleUpdateGroupTask(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	var req dto.UpsertTaskCatalogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.taskUC.UpdateTaskInGroup(c.Context(), group.ID(), id, req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", taskGroupURL(dormitory.ID(), group.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleDeleteCommonTask(c *fiber.Ctx) error {
	dormitory, err := h.requireDormitory(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	if err := h.taskUC.DeleteCommonTask(c.Context(), id); err != nil {
		return c.Status(500).SendString("Ошибка при удалении задачи")
	}
	c.Response().Header.Set("HX-Redirect", commonTaskURL(dormitory.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleDeleteGroupTask(c *fiber.Ctx) error {
	dormitory, group, err := h.requireDormitoryGroup(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	if err := h.taskUC.DeleteTaskInGroup(c.Context(), group.ID(), id); err != nil {
		return c.Status(500).SendString("Ошибка при удалении задачи")
	}
	c.Response().Header.Set("HX-Redirect", taskGroupURL(dormitory.ID(), group.ID()))
	return c.SendString("")
}

func (h *AdminHandler) HandleEditTaskModal(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	task, err := h.taskUC.GetTask(c.Context(), id)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении задачи")
	}
	if task == nil {
		return c.Status(404).SendString("Task not found")
	}
	areas, err := h.taskUC.ListAreas(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении зон")
	}
	return c.Render("partials/modal_edit_task", fiber.Map{"Task": task, "Areas": areas})
}

func (h *AdminHandler) HandleUpdateTask(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	var req dto.UpsertTaskCatalogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.taskUC.UpdateTask(c.Context(), id, req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", "/admin/tasks")
	return c.SendString("")
}

func (h *AdminHandler) HandleDeleteTask(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid task id")
	}
	if err := h.taskUC.DeleteTask(c.Context(), id); err != nil {
		return c.Status(500).SendString("Ошибка при удалении задачи")
	}
	c.Response().Header.Set("HX-Redirect", "/admin/tasks")
	return c.SendString("")
}

func (h *AdminHandler) HandleRegenerateSheet(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.FormValue("team_id"))
	if err != nil {
		return c.Status(400).SendString("Invalid team id")
	}
	if err := h.sheetsUC.RegenerateCurrentDutySheet(c.Context(), teamID); err != nil {
		return c.Status(500).SendString("Ошибка при пересоздании Google Sheet")
	}
	c.Response().Header.Set("HX-Redirect", "/admin/users")
	return c.SendString("")
}
