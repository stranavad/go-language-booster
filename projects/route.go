package projects

import (
	"languageboostergo/types"
	"languageboostergo/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, config types.ServiceConfig){
	projectService := Service{ServiceConfig: config}
	projectsGroup := r.Group("/projects")
	projectsGroup.Use(utils.AuthMiddleware())
	projectsGroup.GET("by-id/:projectId", projectService.GetById)
	projectsGroup.GET(":spaceId", projectService.ListProjects)
	projectsGroup.POST("", projectService.CreateProject)
	projectsGroup.PUT(":projectId", projectService.UpdateProject)
}
