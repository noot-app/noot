package main

import (
	"fmt"
	"log"
	"os"

	"github.com/grantbirki/noot/internal/storage"
)

func main() {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./noot.db"
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	if err := store.Seed(); err != nil {
		log.Fatalf("Failed to seed: %v", err)
	}

	fmt.Println("Database seeded successfully")
}
