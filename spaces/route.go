package spaces

import (
	"languageboostergo/types"
	"languageboostergo/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, config types.ServiceConfig) {
	spacesService := Service{ServiceConfig: config}

	spacesGroup := r.Group("/spaces")
	spacesGroup.Use(utils.AuthMiddleware())
	spacesGroup.PUT(":spaceId", spacesService.UpdateSpace)
	spacesGroup.GET("", spacesService.ListUserSpaces)
	spacesGroup.GET(":spaceId", spacesService.GetById)
	spacesGroup.GET(":spaceId/members", spacesService.GetSpaceMembers)
	spacesGroup.POST("", spacesService.CreateSpace)
	spacesGroup.POST("invite", spacesService.AddUserToSpace)
	spacesGroup.PUT("member", spacesService.UpdateMember)
	spacesGroup.DELETE("member", spacesService.DeleteMember)
	spacesGroup.POST("leave/:spaceId", spacesService.LeaveSpace)
}
