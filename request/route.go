package request

import (
	"github.com/gin-gonic/gin"
	"languageboostergo/types"
	"languageboostergo/utils"
)

func RegisterRoutes(r *gin.Engine, config types.ServiceConfig) {
	requestsService := Service{ServiceConfig: config}
	requestsGroup := r.Group("/requests")
	requestsGroup.Use(utils.AuthMiddleware())
	requestsGroup.POST("", requestsService.CreateRequest)
	requestsGroup.GET(":projectId", requestsService.GetRequests)
	requestsGroup.DELETE(":requestId", requestsService.DeleteRequest)
}
