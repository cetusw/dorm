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
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer ctn.Close()

	log.Println("Starting Scheduler...")
	ctn.Scheduler.Start()

	go func() {
		log.Println("Starting Telegram Bot...")
		ctn.Bot.Start()
	}()

	ctx := context.Background()
	if err := ctn.CleaningService.StartNewWeek(ctx); err != nil {
		log.Printf("Failed to start weekly duty: %v", err)
	} else {
		log.Println("Weekly duty started successfully!")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
}
