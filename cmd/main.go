package main

import (
	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/repository"
	"cruder/internal/service"
	"log"
	"os"

	"github.com/gin-gonic/gin"
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

	repositories := repository.NewRepository(dbConnection.DB())
	services := service.NewService(repositories)
	controllers := controller.NewController(services)
	httpRouterEngine := gin.Default()
	handler.New(httpRouterEngine, controllers.Users)
	if err := httpRouterEngine.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
