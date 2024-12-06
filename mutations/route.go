package mutations

import (
	"languageboostergo/types"
	"languageboostergo/utils"

	"github.com/gin-gonic/gin"
)


func RegisterRoutes(r *gin.Engine, config types.ServiceConfig){
	mutationService := Service{ServiceConfig: config}
	mutationsGroup := r.Group("/mutations")
	mutationsGroup.Use(utils.AuthMiddleware())
	mutationsGroup.POST("/project/:projectId/search",mutationService.SearchByProject)
	mutationsGroup.GET(":mutationId",mutationService.GetById)
	mutationsGroup.GET("/project/:projectId",mutationService.ListByProject)
	mutationsGroup.POST("",mutationService.CreateMutation)
	mutationsGroup.PUT(":mutationId",mutationService.UpdateMutation)
	mutationsGroup.DELETE(":mutationId",mutationService.DeleteById)
	mutationsGroup.POST("/value",mutationService.CreateMutationValue)
	mutationsGroup.PUT("/value/:mutationValueId",mutationService.UpdateMutationValue)
}
