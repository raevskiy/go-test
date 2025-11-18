package handler

import (
	"cruder/internal/controller"
	"cruder/internal/repository"
	"cruder/internal/service"
	"database/sql"
	"github.com/gin-gonic/gin"
)

func SetupAppLayersAndRouter(db *sql.DB) *gin.Engine {
	repositories := repository.NewRepository(db)
	services := service.NewService(repositories)
	controllers := controller.NewController(services)
	httpRouterEngine := gin.Default()
	New(httpRouterEngine, controllers.Users)

	return httpRouterEngine
}
