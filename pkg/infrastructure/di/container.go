package di

import (
	"database/sql"
	"fmt"
	"log"

	"dorm/pkg/adapters/sheets/infrastructure"
	"dorm/pkg/adapters/telegram"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/usecase/cleaning"
	"dorm/pkg/core/usecase/user"
	"dorm/pkg/infrastructure/config"
	"dorm/pkg/infrastructure/eventbus"
	"dorm/pkg/infrastructure/mysql"
	"dorm/pkg/infrastructure/mysql/repository"
	"dorm/pkg/infrastructure/scheduler"
	"dorm/pkg/infrastructure/spreadsheet"
)

type Container struct {
	Config             *config.AppConfig
	DB                 *sql.DB
	EventBus           ports.EventBus
	CleaningService    ports.CleaningUseCase
	Bot                *telegram.BotAdapter
	SpreadsheetClient  spreadsheet.Client
	SpreadsheetAdapter *infrastructure.SpreadsheetAdapter
	Scheduler          *scheduler.Scheduler
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

	spreadsheetClient, err := spreadsheet.NewSpreadsheetClient(cfg.GoogleCredentials)
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

	userService := user.NewUserService(userRepo, teamRepo, groupRepo, dormitoryRepo)

	botAdapter, err := telegram.NewBotAdapter(cfg.BotToken, cleaningService, userService)
	if err != nil {
		return nil, fmt.Errorf("bot init failed: %w", err)
	}

	sheetsAdapter, err := infrastructure.NewSpreadsheetAdapter(cleaningService, userService, spreadsheetClient, bus, cfg)
	if err != nil {
		return nil, fmt.Errorf("sheets adapter init failed: %w", err)
	}

	cronScheduler := scheduler.NewScheduler(cleaningService, cfg)

	return &Container{
		Config:             cfg,
		DB:                 db,
		EventBus:           bus,
		CleaningService:    cleaningService,
		Bot:                botAdapter,
		SpreadsheetClient:  spreadsheetClient,
		SpreadsheetAdapter: sheetsAdapter,
		Scheduler:          cronScheduler,
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
