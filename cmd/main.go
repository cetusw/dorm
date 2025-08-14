package main

import (
	"dorm/internal/dorm/application/bot"
	"dorm/internal/dorm/application/service"
	"log"
	"os"

	"dorm/internal/dorm/infrastructure"
	"dorm/internal/dorm/infrastructure/mysql/repository"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Printf(".env file loaded successfully.")
	userRepository, err := repository.NewUserRepository(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Database connection established.")
	defer userRepository.Close()

	telegram, err := infrastructure.NewTelegram(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Failed to create Telegram client: %v", err)
	}
	//sheets, err := infrastructure.NewSheets(os.Getenv("SHEETS_CREDENTIALS"))
	//if err != nil {
	//	log.Fatalf("Failed to create Sheets client: %v", err)
	//}
	userService := service.NewUserService(userRepository, telegram)
	botContext := bot.NewBot(telegram, userService)

	updates := telegram.GetUpdates(telegram.Bot)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		botContext.State.Handle(botContext, &update)
	}

	c := cron.New()
	c.AddFunc("30 7 * * TUE,THU,SAT", func() {
	})

	c.Start()

	log.Println("Dorm cleaning bot started...")
	select {}
}
