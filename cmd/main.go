package main

import (
	"cruder/internal/core"
	"cruder/internal/repository"
	"log"
	"os"
)

func main() {
	dataSourceName := os.Getenv("POSTGRES_DSN")
	if dataSourceName == "" {
		dataSourceName = "host=localhost port=5432 user=postgres password=postgres dbname=cruderdb sslmode=disable"
	}

	dbConnection, err := repository.NewPostgresConnection(dataSourceName)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	_, httpRouterEngine := core.SetupAppLayers(dbConnection.DB())
	if err := httpRouterEngine.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

