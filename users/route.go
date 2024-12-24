package users

import (
	"languageboostergo/types"
	"languageboostergo/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, config types.ServiceConfig) {
	userService := Service{ServiceConfig: config}

	usersGroup := r.Group("/users")
	usersGroup.GET("current", utils.AuthMiddleware(), userService.GetCurrent)
	usersGroup.GET("spaces", utils.AuthMiddleware(), userService.GetUserSpaces)
	usersGroup.POST("register", userService.CreateUser)
	usersGroup.POST("login", userService.LoginUser)
}
