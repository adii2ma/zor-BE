package main

import (
	"context"
	"log"
	"os"
	"time"

	"be-zor/internal/config"
	"be-zor/internal/database"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up":
		group, err := database.Migrate(ctx, db)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(database.FormatMigrationGroup(group))
	case "down":
		group, err := database.Rollback(ctx, db)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(database.FormatMigrationGroup(group))
	case "status":
		statuses, err := database.MigrationStatus(ctx, db)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("\n" + database.FormatMigrationStatus(statuses))
	default:
		log.Fatalf("unsupported command %q, use up, down, or status", command)
	}
}
