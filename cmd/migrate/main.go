package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	migrator "dorm/data/mysql"
	"dorm/pkg/infrastructure/config"
	infraMysql "dorm/pkg/infrastructure/mysql"
)

func main() {
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	cfg, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("failed to load database config: %v", err)
	}

	db, err := infraMysql.NewConnection(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	switch command {
	case "up":
		if err := migrator.ApplyMigrations(db); err != nil {
			log.Fatalf("migration up failed: %v", err)
		}
		log.Println("Database migrations applied successfully")
	case "down":
		if err := migrator.RollbackMigration(db); err != nil {
			log.Fatalf("migration down failed: %v", err)
		}
		log.Println("Database migration rolled back successfully")
	case "version":
		version, dirty, err := migrator.MigrationVersion(db)
		if err != nil {
			log.Fatalf("migration version failed: %v", err)
		}
		fmt.Printf("version=%d dirty=%t\n", version, dirty)
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("usage: migrate force <version>")
		}
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("invalid migration version: %v", err)
		}
		if err := migrator.ForceMigrationVersion(db, version); err != nil {
			log.Fatalf("migration force failed: %v", err)
		}
		log.Printf("Forced migration version to %d", version)
	default:
		log.Fatalf("unknown command %q; expected one of: up, down, version, force", command)
	}
}
