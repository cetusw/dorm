package di

import (
	"database/sql"
	"dorm/pkg/adapters/http"
	"dorm/pkg/core/usecase/dormitory"
	"dorm/pkg/core/usecase/team"
	"dorm/pkg/infrastructure/mysql/query"
	"fmt"
	"log"

	"dorm/pkg/adapters/sheets/infrastructure"
	"dorm/pkg/core/ports"
	cataloguc "dorm/pkg/core/usecase/catalog"
	"dorm/pkg/core/usecase/cleaning"
	"dorm/pkg/core/usecase/user"
	"dorm/pkg/infrastructure/config"
	"dorm/pkg/infrastructure/eventbus"
	"dorm/pkg/infrastructure/mysql"
	"dorm/pkg/infrastructure/mysql/repository"
	"dorm/pkg/infrastructure/scheduler"
	"dorm/pkg/infrastructure/spreadsheet"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type Container struct {
	Config          *config.AppConfig
	DB              *sql.DB
	EventBus        ports.EventBus
	CleaningService ports.CleaningUseCase
	//Bot                *telegram.BotAdapter
	SpreadsheetClient  spreadsheet.Client
	SpreadsheetAdapter *infrastructure.SpreadsheetAdapter
	Scheduler          *scheduler.Scheduler
	HTTPServer         *fiber.App
}

type SpreadsheetClientFactory func(credentialsJSON string) (spreadsheet.Client, error)

type ContainerOptions struct {
	NewSpreadsheetClient SpreadsheetClientFactory
}

func NewContainer(configPath string) (*Container, error) {
	return NewContainerWithOptions(configPath, ContainerOptions{})
}

func NewContainerWithOptions(configPath string, opts ContainerOptions) (*Container, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := mysql.NewConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	newSpreadsheetClient := opts.NewSpreadsheetClient
	if newSpreadsheetClient == nil {
		newSpreadsheetClient = func(credentialsJSON string) (spreadsheet.Client, error) {
			return spreadsheet.NewSpreadsheetClient(credentialsJSON)
		}
	}

	spreadsheetClient, err := newSpreadsheetClient(cfg.GoogleCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create spreadsheet client: %w", err)
	}

	bus := eventbus.NewInMemoryEventBus()

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	dutyRepo := repository.NewDutyRepository(db)
	areaRepo := repository.NewAreaRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	dormitoryRepo := repository.NewDormitoryRepository(db)

	cleaningService := cleaning.NewCleaningService(
		userRepo,
		teamRepo,
		groupRepo,
		dutyRepo,
		taskRepo,
		areaRepo,
		bus,
	)

	userQueryService := query.NewUserQueryService(db)
	teamQueryService := query.NewTeamQueryService(db)

	userService := user.NewUserService(
		userRepo,
		teamRepo,
		groupRepo,
		dormitoryRepo,
		userQueryService,
	)

	dormitoryService := dormitory.NewDormitoryService(dormitoryRepo)
	teamService := team.NewTeamService(teamRepo, groupRepo, dormitoryRepo, userRepo, teamQueryService)
	taskCatalogService := cataloguc.NewCatalogService(taskRepo, areaRepo)

	//botAdapter, err := telegram.NewBotAdapter(cfg.BotToken, cleaningService, userService)
	//if err != nil {
	//	return nil, fmt.Errorf("bot init failed: %w", err)
	//}

	sheetsAdapter, err := infrastructure.NewSpreadsheetAdapter(cleaningService, userService, spreadsheetClient, bus, cfg)
	if err != nil {
		return nil, fmt.Errorf("sheets adapter init failed: %w", err)
	}

	cronScheduler := scheduler.NewScheduler(cleaningService, cfg)

	engine := html.New("./web", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	adminHandler := http.NewAdminHandler(userService, cleaningService, dormitoryService, teamService, taskCatalogService, sheetsAdapter)
	adminHandler.RegisterRoutes(app)

	return &Container{
		Config:          cfg,
		DB:              db,
		EventBus:        bus,
		CleaningService: cleaningService,
		//Bot:                botAdapter,
		SpreadsheetClient:  spreadsheetClient,
		SpreadsheetAdapter: sheetsAdapter,
		Scheduler:          cronScheduler,
		HTTPServer:         app,
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

// TODO: поменять названия spreadsheet на нормальное googlesheets
