package main

import (
	"database/sql"
	"dorm/internal/dorm/application/bot"
	"dorm/internal/dorm/application/scheduler"
	"dorm/internal/dorm/application/service"
	"fmt"
	"log"
	"os"

	"dorm/internal/dorm/infrastructure"
	"dorm/internal/dorm/infrastructure/mysql/repository"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Printf(".env file loaded successfully.")
	db := connectDatabase()
	userRepository := repository.NewUserRepository(db)
	dutyRepository := repository.NewDutyRepository(db)
	dutyTaskRepository := repository.NewDutyTaskRepository(db)
	teamRepository := repository.NewTeamRepository(db)
	taskRepository := repository.NewTaskRepository(db)
	areaRepository := repository.NewAreaRepository(db)

	telegram, err := infrastructure.NewTelegram(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Failed to create Telegram client: %v", err)
	}

	sheets, err := infrastructure.NewSheets(os.Getenv("SHEETS_CREDENTIALS"), os.Getenv("SPREADSHEET_ID"))
	if err != nil {
		log.Fatalf("Failed to create Sheets client: %v", err)
	}

	userService := service.NewUserService(userRepository)
	dutyService := service.NewDutyService(dutyRepository)
	dutyTaskService := service.NewDutyTaskService(
		dutyTaskRepository,
		taskRepository,
		dutyRepository,
		teamRepository,
	)
	teamService := service.NewTeamService(teamRepository)
	taskService := service.NewTaskService(taskRepository)
	sheetsService := service.NewSheetsService(sheets)
	cleaningService := service.NewCleaningService(
		sheetsService,
		userService,
		dutyService,
		dutyTaskService,
		teamService,
	)
	areaService := service.NewAreaService(areaRepository)
	botContext := bot.NewBot(
		telegram,
		userService,
		areaService,
		taskService,
		dutyTaskService,
	)

	s := scheduler.NewScheduler(cleaningService)
	s.RegisterJobs()
	s.Start()

	updates := telegram.GetUpdates(telegram.Bot)

	for update := range updates {
		log.Println(botContext.State.GetName())
		if update.Message != nil {
			botContext.State.Handle(botContext, &update)
		} else if update.CallbackQuery != nil {
			botContext.State.HandleCallback(botContext, &update)
		}
	}
}

func connectDatabase() *sql.DB {
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("failed to open db connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	log.Println("Database connection successful!")

	return db
}
