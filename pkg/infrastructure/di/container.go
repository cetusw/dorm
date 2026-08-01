package di

import (
	"database/sql"
	"dorm/pkg/adapters/http"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/usecase/dormitory"
	dutysettingsuc "dorm/pkg/core/usecase/dutysettings"
	notificationuc "dorm/pkg/core/usecase/notification"
	residentusecase "dorm/pkg/core/usecase/resident"
	"dorm/pkg/core/usecase/team"
	"dorm/pkg/infrastructure/mysql/query"
	notificationinfra "dorm/pkg/infrastructure/notification"
	"fmt"
	"log"

	"dorm/pkg/core/ports"
	cataloguc "dorm/pkg/core/usecase/catalog"
	"dorm/pkg/core/usecase/cleaning"
	"dorm/pkg/core/usecase/user"
	"dorm/pkg/infrastructure/config"
	"dorm/pkg/infrastructure/eventbus"
	"dorm/pkg/infrastructure/mysql"
	"dorm/pkg/infrastructure/mysql/repository"
	"dorm/pkg/infrastructure/scheduler"

	"github.com/gofiber/fiber/v2"
)

type Container struct {
	Config          *config.AppConfig
	DB              *sql.DB
	EventBus        ports.EventBus
	CleaningService ports.CleaningUseCase
	Scheduler       *scheduler.Scheduler
	HTTPServer      *fiber.App
}

func NewContainer(configPath string) (*Container, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := mysql.NewConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	bus := eventbus.NewInMemoryEventBus()

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	dutyRepo := repository.NewDutyRepository(db)
	dutyTaskRepo := repository.NewDutyTaskRepository(db)
	areaRepo := repository.NewAreaRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	taskOverrideRepo := repository.NewDutyTaskOverrideRepository(db)
	dormitoryRepo := repository.NewDormitoryRepository(db)
	pushSubscriptionRepo := repository.NewPushSubscriptionRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	cleaningService := cleaning.NewCleaningService(
		userRepo,
		teamRepo,
		groupRepo,
		dutyRepo,
		dutyTaskRepo,
		taskRepo,
		taskOverrideRepo,
		areaRepo,
		bus,
	)

	userQueryService := query.NewUserQueryService(db)
	teamQueryService := query.NewTeamQueryService(db)
	notificationQueryService := query.NewNotificationQueryService(db)

	userService := user.NewUserService(
		userRepo,
		dormitoryRepo,
		userQueryService,
	)

	dormitoryService := dormitory.NewDormitoryService(dormitoryRepo, groupRepo, teamRepo, userRepo)
	teamService := team.NewTeamService(teamRepo, groupRepo, dormitoryRepo, userRepo, teamQueryService)
	taskCatalogService := cataloguc.NewCatalogService(taskRepo, areaRepo, groupRepo)
	dutySettingsService := dutysettingsuc.NewDutySettingsService(groupRepo, teamRepo, areaRepo, taskRepo, dutyRepo, dutyTaskRepo, userRepo)
	pushSubscriptionService := notificationuc.NewNotificationService(pushSubscriptionRepo)

	var pushSender ports.PushSender = notificationinfra.NoopPushSender{}
	if cfg.WebPush.Enabled {
		pushSender = notificationinfra.NewWebPushSender(cfg.WebPush)
	}

	userNotificationService := notificationuc.NewDeliveryService(
		notificationRepo,
		pushSubscriptionRepo,
		pushSender,
		cfg.WebPush.Enabled,
	)
	dutyReminderService := notificationuc.NewDutyReminderService(
		notificationQueryService,
		notificationQueryService,
		notificationQueryService,
		userNotificationService,
	)
	bus.Subscribe(events.TopicWeekStarted, notificationuc.NewWeekStartedHandler(dutyReminderService).Handle)
	bus.Subscribe(events.TopicTasksReadyForReview, notificationuc.NewTasksReadyForReviewHandler(userNotificationService).Handle)

	cronScheduler, err := scheduler.NewScheduler(cleaningService, dutyReminderService, cfg)
	if err != nil {
		return nil, fmt.Errorf("scheduler init failed: %w", err)
	}

	app := fiber.New()

	app.Get("/app", func(c *fiber.Ctx) error {
		return c.Redirect("/app/tasks", fiber.StatusTemporaryRedirect)
	})

	app.Static("/app", "./web/app")

	app.Get("/app/*", func(c *fiber.Ctx) error {
		return c.SendFile("./web/app/index.html")
	})

	residentAuthHandler := http.NewResidentAuthHandler(userService, cfg.AuthSecret)
	residentAuthHandler.RegisterRoutes(app)

	notificationAPIHandler := http.NewNotificationAPIHandler(cfg.WebPush, pushSubscriptionService)
	notificationAPIHandler.RegisterRoutes(app, http.ResidentAuthMiddleware(cfg.AuthSecret))

	residentDutyService := residentusecase.NewResidentDutyService(
		userRepo,
		teamRepo,
		groupRepo,
		dormitoryRepo,
		dutyRepo,
		taskRepo,
		areaRepo,
		cleaningService,
	)

	residentAPIHandler := http.NewResidentAPIHandler(residentDutyService)
	residentAPIHandler.RegisterRoutes(app, http.ResidentAuthMiddleware(cfg.AuthSecret))

	dormitoryAPIHandler := http.NewDormitoryAPIHandler(dormitoryService)
	dormitoryAPIHandler.RegisterRoutes(app, http.ResidentAuthMiddleware(cfg.AuthSecret))

	groupAPIHandler := http.NewGroupAPIHandler(dormitoryService, cleaningService, dutySettingsService)
	groupAPIHandler.RegisterRoutes(app, http.ResidentAuthMiddleware(cfg.AuthSecret))

	teamAPIHandler := http.NewTeamAPIHandler(dormitoryService, teamService)
	teamAPIHandler.RegisterRoutes(app, http.ResidentAuthMiddleware(cfg.AuthSecret))

	areaAPIHandler := http.NewAreaAPIHandler(dormitoryService, taskCatalogService)
	areaAPIHandler.RegisterRoutes(app, http.ResidentAuthMiddleware(cfg.AuthSecret))

	taskCatalogAPIHandler := http.NewTaskCatalogAPIHandler(dormitoryService, taskCatalogService)
	taskCatalogAPIHandler.RegisterRoutes(app, http.ResidentAuthMiddleware(cfg.AuthSecret))

	userAPIHandler := http.NewUserAPIHandler(userService, dormitoryService)
	userAPIHandler.RegisterRoutes(app, http.ResidentAuthMiddleware(cfg.AuthSecret))

	return &Container{
		Config:          cfg,
		DB:              db,
		EventBus:        bus,
		CleaningService: cleaningService,
		Scheduler:       cronScheduler,
		HTTPServer:      app,
	}, nil
}

func (c *Container) Close() {
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("Database connection closed gracefully")
		}
	}
}
