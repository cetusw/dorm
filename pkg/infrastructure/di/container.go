package di

import (
	"database/sql"
	"fmt"
	"log"

	"dorm/pkg/core/ports"
	"dorm/pkg/infrastructure/config"
	"dorm/pkg/infrastructure/eventbus"
	"dorm/pkg/infrastructure/mysql"
)

type Container struct {
	Config *config.AppConfig
	DB     *sql.DB

	EventBus ports.EventBus

	// 3. Domain Repositories (появятся на Этапе 2)
	// UserRepository user.Repository
	// TaskRepository task.Repository

	// 4. UseCases / Services (появятся на Этапе 2)
	// CleaningService ports.CleaningUseCase

	// 5. Adapters (появятся на Этапе 3)
	// TelegramBot *telegram.Adapter
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

	return &Container{
		Config:   cfg,
		DB:       db,
		EventBus: bus,
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
