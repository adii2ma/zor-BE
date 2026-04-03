package main

import (
	"context"
	"log"
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

	if err := database.ApplySchema(ctx, db); err != nil {
		log.Fatal(err)
	}

	log.Println("schema applied")
}
