package http

import (
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	userUC      ports.UserUseCase
	cleanUC     ports.CleaningUseCase
	dormitoryUC ports.DormitoryUseCase
	teamUC      ports.TeamUseCase
}

func NewAdminHandler(
	userUC ports.UserUseCase,
	cleanUC ports.CleaningUseCase,
	dormitoryUC ports.DormitoryUseCase,
	teamUC ports.TeamUseCase,
) *AdminHandler {
	return &AdminHandler{
		userUC:      userUC,
		cleanUC:     cleanUC,
		dormitoryUC: dormitoryUC,
		teamUC:      teamUC,
	}
}

func (h *AdminHandler) RegisterRoutes(app *fiber.App) {
	admin := app.Group("/admin")

	admin.Get("/users", h.HandleGetUsers)
	admin.Get("/users/create", h.HandleCreateUserModal)
	admin.Get("/users/teams-select", h.HandleTeamsDropdown)
	admin.Post("/users", h.HandleStoreUser)
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
		return c.Status(500).SendString(err.Error())
	}

	c.Response().Header.Set("HX-Trigger", "userCreated")
	return c.Redirect("/admin/users")
}
