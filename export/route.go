package export

import (
	"languageboostergo/types"
	"languageboostergo/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, config types.ServiceConfig){
	exportService := Service{ServiceConfig: config}
	exportsGroup := r.Group("/export")
	exportsGroup.Use(utils.AuthMiddleware())
	exportsGroup.POST("", exportService.ByProjectIdAndLanguageId)
}
