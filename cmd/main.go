package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"dorm/data/mysql"
	"dorm/pkg/common/config"
	"dorm/pkg/dorm/application/bot"
	"dorm/pkg/dorm/application/scheduler"
	"dorm/pkg/dorm/application/service"
	"dorm/pkg/dorm/application/service/notification"
	"dorm/pkg/dorm/application/service/report"
	"dorm/pkg/dorm/application/service/sheets"
	"dorm/pkg/dorm/infrastructure"
	"dorm/pkg/dorm/infrastructure/mysql/repository"
)

func main() {
	configPath := filepath.Join(filepath.Dir(os.Args[0]), "config.json")
	configData, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	db := connectDatabase()
	mysql.ApplyMigrations(db)
	userRepository := repository.NewUserRepository(db)
	dutyRepository := repository.NewDutyRepository(db)
	dutyTaskRepository := repository.NewDutyTaskRepository(db)
	teamRepository := repository.NewTeamRepository(db)
	taskRepository := repository.NewTaskRepository(db)
	areaRepository := repository.NewAreaRepository(db)
	groupRepository := repository.NewGroupRepository(db)
	dormRepository := repository.NewDormitoryRepository(db)

	telegram, err := infrastructure.NewTelegram(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Failed to create Telegram client: %v", err)
	}

	newSheets, err := infrastructure.NewSheets(os.Getenv("SHEETS_CREDENTIALS"))
	if err != nil {
		log.Fatalf("Failed to create Sheets client: %v", err)
	}

	userService := service.NewUserService(userRepository)
	dutyService := service.NewDutyService(dutyRepository)
	dutyTaskService := service.NewDutyTaskService(dutyTaskRepository)
	teamService := service.NewTeamService(teamRepository)
	taskService := service.NewTaskService(taskRepository)
	builder := sheets.NewDutySheetBuilder()
	styler := sheets.NewDutySheetStyler()
	sheetsService := sheets.NewSheetsService(newSheets, builder, styler)
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
	syncService := service.NewSyncService(
		newSheets,
		groupService,
		dutyService,
		dutyTaskService,
		userService,
	)
	notificationService := notification.NewNotificationService(
		userService,
		telegram,
	)
	reportService := report.NewCleaningReportService(
		cleaningService,
		groupService,
		teamService,
		dormRepository,
		dutyService,
		dutyTaskService,
		notificationService,
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

	s := scheduler.NewScheduler(
		*configData,
		cleaningService,
		syncService,
		reportService,
	)
	s.RegisterJobs()
	s.Start()

	err = cleaningService.StartNewWeek()
	if err != nil {
		log.Printf("Failed to start new week: %v", err)
	}

	//err = syncService.SyncAllActiveDuties()
	//if err != nil {
	//	log.Printf("Failed to sync duties: %v", err)
	//}

	updates := telegram.GetUpdates(telegram.Bot)

	for update := range updates {
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
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&allowNativePasswords=true",
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
