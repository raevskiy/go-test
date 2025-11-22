package handler

import (
	"cruder/internal/controller"

	"github.com/gin-gonic/gin"
)

func New(router *gin.Engine, userController *controller.UserController) *gin.Engine {
	apiV1Group := router.Group("/api/v1")
	{
		userGroup := apiV1Group.Group("/users")
		{
			userGroup.GET("/", userController.GetAllUsers)
			userGroup.GET("/username/:username", userController.GetUserByUsername)
			userGroup.GET("/id/:id", userController.GetUserByID)	//This should never exist, to be honest. We are not even going to test it.
			userGroup.DELETE("/:uuid", userController.DeleteUserByUuid)
			userGroup.PATCH("/:uuid", userController.PatchUserByUuid)
		}
	}

	return router
}
