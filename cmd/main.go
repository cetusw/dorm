package main

import (
	"database/sql"
	"dorm/internal/dorm/application/bot"
	"dorm/internal/dorm/application/scheduler"
	"dorm/internal/dorm/application/service"
	"fmt"
	"log"
	"os"
	"time"

	"dorm/internal/dorm/infrastructure"
	"dorm/internal/dorm/infrastructure/mysql/repository"
)

func main() {
	log.Printf(".env file loaded successfully.")
	db := connectDatabase()
	userRepository := repository.NewUserRepository(db)
	dutyRepository := repository.NewDutyRepository(db)
	dutyTaskRepository := repository.NewDutyTaskRepository(db)
	teamRepository := repository.NewTeamRepository(db)
	taskRepository := repository.NewTaskRepository(db)
	areaRepository := repository.NewAreaRepository(db)
	groupRepository := repository.NewGroupRepository(db)

	telegram, err := infrastructure.NewTelegram(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Failed to create Telegram client: %v", err)
	}

	sheets, err := infrastructure.NewSheets(os.Getenv("SHEETS_CREDENTIALS"))
	if err != nil {
		log.Fatalf("Failed to create Sheets client: %v", err)
	}

	userService := service.NewUserService(userRepository)
	dutyService := service.NewDutyService(dutyRepository)
	dutyTaskService := service.NewDutyTaskService(dutyTaskRepository)
	teamService := service.NewTeamService(teamRepository)
	taskService := service.NewTaskService(taskRepository)
	sheetsService := service.NewSheetsService(sheets)
	groupService := service.NewGroupService(groupRepository)
	areaService := service.NewAreaService(areaRepository)
	cleaningService := service.NewCleaningService(
		sheetsService,
		userService,
		dutyService,
		dutyTaskService,
		teamService,
		taskService,
		groupService,
		areaService,
	)
	botContext := bot.NewBot(
		telegram,
		userService,
		areaService,
		taskService,
		dutyService,
		dutyTaskService,
		groupService,
		teamService,
		cleaningService,
	)

	// TODO: remove
	err = cleaningService.StartNewWeek()
	if err != nil {
		log.Println(err)
	}

	s := scheduler.NewScheduler(cleaningService)
	s.RegisterJobs()
	s.Start()

	updates := telegram.GetUpdates(telegram.Bot)

	for update := range updates {
		log.Println(botContext.State.GetName())
		if update.Message != nil {
			err := botContext.State.Handle(botContext, &update)
			if err != nil {
				log.Printf("Failed to handle update: %v", err)
			}
		} else if update.CallbackQuery != nil {
			err := botContext.State.HandleCallback(botContext, &update)
			if err != nil {
				log.Printf("Failed to handle update: %v", err)
			}
		}
	}
}

func connectDatabase() *sql.DB {
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	for i := 0; i < 5; i++ {
		db, err := sql.Open("mysql", connStr)
		if err != nil {
			log.Printf("Attempt %d: failed to open db connection: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := db.Ping(); err != nil {
			log.Printf("Attempt %d: failed to ping db: %v", i+1, err)
			db.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		if _, err := db.Exec("SET time_zone = '+03:00'"); err != nil {
			log.Fatalf("failed to set session time zone: %v", err)
		}
		log.Println("Database connection successful!")
		return db
	}

	log.Fatalf("Failed to connect to database after multiple retries")
	return nil
}
