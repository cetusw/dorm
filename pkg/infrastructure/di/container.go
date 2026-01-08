package di

import (
	"database/sql"
	"fmt"
	"log"

	"dorm/pkg/adapters/telegram"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/usecase/cleaning"
	"dorm/pkg/core/usecase/user"
	"dorm/pkg/infrastructure/config"
	"dorm/pkg/infrastructure/eventbus"
	"dorm/pkg/infrastructure/gsheets"
	"dorm/pkg/infrastructure/mysql"
	"dorm/pkg/infrastructure/mysql/repository"
)

type Container struct {
	Config          *config.AppConfig
	DB              *sql.DB
	EventBus        ports.EventBus
	CleaningService ports.CleaningUseCase
	Bot             *telegram.BotAdapter
	GSheetsClient   gsheets.SpreadsheetClient
}

func NewContainer(configPath string) (*Container, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	gsheetsClient, err := gsheets.NewGoogleSheetsClient(cfg.GoogleCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create gsheets client: %w", err)
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

	return &Container{
		Config:          cfg,
		DB:              db,
		EventBus:        bus,
		CleaningService: cleaningService,
		Bot:             botAdapter,
		GSheetsClient:   gsheetsClient,
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

// TODO: поменять названия gsheets на нормальное googlesheets
