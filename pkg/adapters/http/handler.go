package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AdminHandler struct {
	userUC      ports.UserUseCase
	cleanUC     ports.CleaningUseCase
	dormitoryUC ports.DormitoryUseCase
	teamUC      ports.TeamUseCase
	taskUC      ports.TaskCatalogUseCase
	sheetsUC    ports.SheetsUseCase
}

type dashboardDutyCard struct {
	Duty            dto.DutyViewModel
	Total           int
	Assigned        int
	Done            int
	Verified        int
	AssignedPercent int
	DonePercent     int
	VerifiedPercent int
}

func NewAdminHandler(
	userUC ports.UserUseCase,
	cleanUC ports.CleaningUseCase,
	dormitoryUC ports.DormitoryUseCase,
	teamUC ports.TeamUseCase,
	taskUC ports.TaskCatalogUseCase,
	sheetsUC ports.SheetsUseCase,
) *AdminHandler {
	return &AdminHandler{
		userUC:      userUC,
		cleanUC:     cleanUC,
		dormitoryUC: dormitoryUC,
		teamUC:      teamUC,
		taskUC:      taskUC,
		sheetsUC:    sheetsUC,
	}
}

func (h *AdminHandler) RegisterRoutes(app *fiber.App) {
	admin := app.Group("/admin")

	admin.Get("/", h.HandleDashboard)
	admin.Get("/users", h.HandleGetUsers)
	admin.Get("/users/create", h.HandleCreateUserModal)
	admin.Get("/users/teams-select", h.HandleTeamsDropdown)
	admin.Post("/users", h.HandleStoreUser)
	admin.Get("/users/:id/edit", h.HandleEditUserModal)
	admin.Put("/users/:id", h.HandleUpdateUser)
	admin.Delete("/users/:id", h.HandleDeleteUser)

	admin.Get("/teams", h.HandleGetTeams)
	admin.Get("/teams/create", h.HandleCreateTeamModal)
	admin.Post("/teams", h.HandleStoreTeam)
	admin.Get("/teams/:id/edit", h.HandleEditTeamModal)
	admin.Put("/teams/:id", h.HandleUpdateTeam)
	admin.Delete("/teams/:id", h.HandleDeleteTeam)
	admin.Put("/teams/:teamID/members/:userID", h.HandleAddTeamMember)
	admin.Delete("/teams/members/:userID", h.HandleRemoveTeamMember)

	admin.Get("/tasks", h.HandleGetTasks)
	admin.Get("/tasks/create", h.HandleCreateTaskModal)
	admin.Post("/tasks", h.HandleStoreTask)
	admin.Get("/tasks/:id/edit", h.HandleEditTaskModal)
	admin.Put("/tasks/:id", h.HandleUpdateTask)
	admin.Delete("/tasks/:id", h.HandleDeleteTask)
	admin.Post("/sheets/regenerate", h.HandleRegenerateSheet)
}

func (h *AdminHandler) HandleDashboard(c *fiber.Ctx) error {
	duties, err := h.cleanUC.GetLatestDuties(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении текущих дежурств")
	}

	cards := make([]dashboardDutyCard, 0, len(duties))
	for _, d := range duties {
		sort.SliceStable(d.UsersStats, func(i, j int) bool {
			left := d.UsersStats[i]
			right := d.UsersStats[j]
			if left.IsTeamLeader != right.IsTeamLeader {
				return left.IsTeamLeader
			}
			leftLast := strings.ToLower(strings.TrimSpace(left.LastName))
			rightLast := strings.ToLower(strings.TrimSpace(right.LastName))
			if leftLast != rightLast {
				return leftLast < rightLast
			}
			leftFirst := strings.ToLower(strings.TrimSpace(left.FirstName))
			rightFirst := strings.ToLower(strings.TrimSpace(right.FirstName))
			return leftFirst < rightFirst
		})

		total := len(d.Tasks)
		assigned := 0
		done := 0
		verified := 0
		for _, t := range d.Tasks {
			if t.Assignee != nil {
				assigned++
			}
			if t.IsCompleted {
				done++
			}
			if t.IsVerified {
				verified++
			}
		}
		assignedPct := 0
		donePct := 0
		verifiedPct := 0
		if total > 0 {
			assignedPct = (assigned * 100) / total
			donePct = (done * 100) / total
			verifiedPct = (verified * 100) / total
		}

		cards = append(cards, dashboardDutyCard{
			Duty:            d,
			Total:           total,
			Assigned:        assigned,
			Done:            done,
			Verified:        verified,
			AssignedPercent: assignedPct,
			DonePercent:     donePct,
			VerifiedPercent: verifiedPct,
		})
	}

	return c.Render("dashboard", fiber.Map{
		"Cards": cards,
	}, "layouts/main")
}

func (h *AdminHandler) HandleGetUsers(c *fiber.Ctx) error {
	users, err := h.userUC.GetUsersList(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении списка пользователей")
	}

	return c.Render("users", fiber.Map{
		"Users": users,
	}, "layouts/main")
}

func (h *AdminHandler) HandleCreateUserModal(c *fiber.Ctx) error {
	dorms, _ := h.dormitoryUC.GetDormitories(c.Context())
	return c.Render("partials/modal_create_user", fiber.Map{
		"Dormitories": dorms,
	})
}

func (h *AdminHandler) HandleTeamsDropdown(c *fiber.Ctx) error {
	dormID := c.QueryInt("dormitory_id")
	teams, _ := h.teamUC.GetTeamsByDormitory(c.Context(), int64(dormID))
	return c.Render("partials/team_options", fiber.Map{
		"Teams": teams,
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
	teamID := ""
	if u.TeamID() != nil {
		teamID = u.TeamID().String()
	}

	dorms, _ := h.dormitoryUC.GetDormitories(c.Context())
	teams, _ := h.teamUC.GetTeamsByDormitory(c.Context(), dormID)

	return c.Render("partials/modal_edit_user", fiber.Map{
		"User": fiber.Map{
			"ID":          u.ID(),
			"FirstName":   u.FirstName(),
			"LastName":    u.LastName(),
			"RoomNumber":  room,
			"DormitoryID": dormID,
			"TeamID":      teamID,
		},
		"Dormitories": dorms,
		"Teams":       teams,
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

	// Simplest reliable way to reflect updated joins (team/group/dorm names).
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

func (h *AdminHandler) HandleGetTeams(c *fiber.Ctx) error {
	dorms, err := h.dormitoryUC.GetDormitories(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении общежитий")
	}
	selectedDormID := int64(c.QueryInt("dormitory_id"))
	if selectedDormID == 0 && len(dorms) > 0 {
		selectedDormID = dorms[0].ID()
	}
	teams, err := h.teamUC.GetTeamsList(c.Context(), selectedDormID)
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении команд")
	}
	return c.Render("teams", fiber.Map{
		"Dormitories":    dorms,
		"SelectedDormID": selectedDormID,
		"Teams":          teams,
	}, "layouts/main")
}

func (h *AdminHandler) HandleCreateTeamModal(c *fiber.Ctx) error {
	groups, _ := h.teamUC.GetGroups(c.Context())
	return c.Render("partials/modal_create_team", fiber.Map{
		"Groups": groups,
	})
}

func (h *AdminHandler) HandleStoreTeam(c *fiber.Ctx) error {
	var req dto.CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.teamUC.CreateTeam(c.Context(), req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", "/admin/teams")
	return c.SendString("")
}

func (h *AdminHandler) HandleEditTeamModal(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
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

	groups, _ := h.teamUC.GetGroups(c.Context())
	members, _ := h.teamUC.GetTeamMembersForEdit(c.Context(), teamID)
	return c.Render("partials/modal_edit_team", fiber.Map{
		"Team":    teamItem,
		"Groups":  groups,
		"Members": members,
	})
}

func (h *AdminHandler) HandleUpdateTeam(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid team id")
	}

	var req dto.UpdateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.teamUC.UpdateTeam(c.Context(), teamID, req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", "/admin/teams")
	return c.SendString("")
}

func (h *AdminHandler) HandleDeleteTeam(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid team id")
	}
	if err := h.teamUC.DeleteTeam(c.Context(), teamID); err != nil {
		return c.Status(500).SendString("Ошибка при удалении команды")
	}
	c.Response().Header.Set("HX-Redirect", "/admin/teams")
	return c.SendString("")
}

func (h *AdminHandler) HandleAddTeamMember(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("teamID"))
	if err != nil {
		return c.Status(400).SendString("Invalid team id")
	}
	userID, err := uuid.Parse(c.Params("userID"))
	if err != nil {
		return c.Status(400).SendString("Invalid user id")
	}
	if err := h.teamUC.MoveUserToTeam(c.Context(), teamID, userID); err != nil {
		return c.Status(500).SendString("Ошибка при добавлении участника")
	}
	c.Response().Header.Set("HX-Redirect", "/admin/teams")
	return c.SendString("")
}

func (h *AdminHandler) HandleRemoveTeamMember(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("userID"))
	if err != nil {
		return c.Status(400).SendString("Invalid user id")
	}
	if err := h.teamUC.RemoveUserFromTeam(c.Context(), userID); err != nil {
		return c.Status(500).SendString("Ошибка при удалении участника")
	}
	c.Response().Header.Set("HX-Redirect", "/admin/teams")
	return c.SendString("")
}

func (h *AdminHandler) HandleGetTasks(c *fiber.Ctx) error {
	tasks, err := h.taskUC.ListTasks(c.Context())
	if err != nil {
		return c.Status(500).SendString("Ошибка при получении задач")
	}
	return c.Render("tasks", fiber.Map{"Tasks": tasks}, "layouts/main")
}

func (h *AdminHandler) HandleCreateTaskModal(c *fiber.Ctx) error {
	areas, _ := h.taskUC.ListAreas(c.Context())
	return c.Render("partials/modal_create_task", fiber.Map{"Areas": areas})
}

func (h *AdminHandler) HandleStoreTask(c *fiber.Ctx) error {
	var req dto.UpsertTaskCatalogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid data")
	}
	if err := h.taskUC.CreateTask(c.Context(), req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	c.Response().Header.Set("HX-Redirect", "/admin/tasks")
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
	areas, _ := h.taskUC.ListAreas(c.Context())
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
	c.Response().Header.Set("HX-Redirect", "/admin")
	return c.SendString("")
}
