package main

import (
	"log"

	"dorm/pkg/infrastructure/di"
)

const configPath = "../config.json"

func main() {
	dependencyContainer, err := di.NewContainer(configPath)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer dependencyContainer.Close()

	log.Println("Application initialized successfully")
	log.Println("Connected to DB...")

	// 2. В будущем запуск бота будет выглядеть так:
	// log.Println("Starting Bot...")
	// dependencyContainer.TelegramBot.Start()

	// А пока просто чтобы процесс не падал сразу, можно поставить ожидание (или убрать для теста)
	select {}
}
