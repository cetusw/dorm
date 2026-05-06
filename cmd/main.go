package main

import (
	"dorm/pkg/infrastructure/di"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	//go func() {
	//	log.Println("Starting Telegram Bot...")
	//	ctn.Bot.Start()
	//}()

	go func() {
		log.Println("Starting Admin Panel on :8080...")
		if err := ctn.HTTPServer.Listen(":8080"); err != nil {
			log.Printf("Fiber error: %v", err)
		}
	}()

	//log.Println("Boot: triggering StartNewWeek...")
	//ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	//defer cancel()
	//if err := ctn.CleaningService.StartNewWeek(ctx); err != nil {
	//	log.Printf("Failed to start weekly duty: %v", err)
	//} else {
	//	log.Println("Weekly duty started successfully!")
	//}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
}
