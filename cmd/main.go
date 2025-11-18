package main

import (
	"cruder/internal/handler"
	"cruder/internal/repository"
	"log"
	"os"
)

func main() {
	dataSourceName := os.Getenv("POSTGRES_DSN")
	if dataSourceName == "" {
		dataSourceName = "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	dbConnection, err := repository.NewPostgresConnection(dataSourceName)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	httpRouterEngine := handler.SetupAppLayersAndRouter(dbConnection.DB())
	if err := httpRouterEngine.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

