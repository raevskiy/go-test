package controller

import (
	"cruder/internal/controller/dto"
	"cruder/internal/model"
	"errors"
	"net/http"
	"strconv"

	"cruder/internal/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service service.UserService
}

const errorKey = "error"
const genericServerErrorValue = "It's not you. It's us. We are already working on it."
const invalidIdClientErrorValue = "invalid id"

func NewUserController(service service.UserService) *UserController {
	return &UserController{service: service}
}

func (c *UserController) GetAllUsers(ctx *gin.Context) {
	users, err := c.service.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{errorKey: genericServerErrorValue})
		return
	}

	ctx.JSON(http.StatusOK, toUserResponses(users))
}

func (c *UserController) GetUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")

	user, err := c.service.GetByUsername(username)
	createSingleUserResponse(user, err, ctx)
}

func (c *UserController) GetUserByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{errorKey: invalidIdClientErrorValue})
		return
	}

	user, err := c.service.GetByID(id)
	createSingleUserResponse(user, err, ctx)
}

func createSingleUserResponse(user *model.User, err error, ctx *gin.Context) {
	if errors.Is(err, service.BusinessErrNoUsers) {
		ctx.JSON(http.StatusNotFound, gin.H{errorKey: err.Error()})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{errorKey: genericServerErrorValue})
		return
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
}

func toUserResponses(users []model.User) []dto.UserResponse {
	allUsersResponses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		allUsersResponses = append(allUsersResponses, toUserResponse(&user))
	}
	return allUsersResponses
}

func toUserResponse(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		UUID:     user.UUID,
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
	}
}
