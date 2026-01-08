package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dorm/pkg/infrastructure/di"
)

func main() {
	configPath := "config.json"

	ctn, err := di.NewContainer(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to initialize container: %v", err)
	}
	defer ctn.Close()

	log.Println("✅ Application initialized successfully")

	go func() {
		log.Println("🚀 Starting Telegram Bot...")
		ctn.Bot.Start()
	}()

	ctx := context.Background()
	if err := ctn.CleaningService.StartNewWeek(ctx); err != nil {
		log.Printf("⚠️ Failed to start weekly duty: %v", err)
	} else {
		log.Println("✅ Weekly duty started successfully!")
	}

	// 4. (Опционально) Запускаем слушателя событий (для обновления таблиц)
	// Пока у нас нет SheetsAdapter, можно просто логировать события в консоль
	// ctn.EventBus.Subscribe("task.completed", func(ctx context.Context, e interface{}) error {
	//     log.Printf("EVENT: Task completed! %+v", e)
	//     return nil
	// })

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down...")
}
