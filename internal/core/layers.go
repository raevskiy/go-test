package core

import (
	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/repository"
	"cruder/internal/service"
	"database/sql"
	"github.com/gin-gonic/gin"
)

func SetupAppLayers(db *sql.DB, xApiKey string) (*repository.Repository, *gin.Engine) {
	repositories := repository.NewRepository(db)
	services := service.NewService(repositories)
	controllers := controller.NewController(services)
	httpRouterEngine := gin.Default()
	handler.New(httpRouterEngine, controllers.Users, xApiKey)

	return repositories, httpRouterEngine
}
