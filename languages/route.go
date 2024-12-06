package languages

import (
	"languageboostergo/types"
	"languageboostergo/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, config types.ServiceConfig){
	languageService := Service{ServiceConfig: config}
	languagesGroup := r.Group("/languages")
	languagesGroup.Use(utils.AuthMiddleware())
	languagesGroup.GET(":projectId", languageService.GetLanguagesByProjectId)
	languagesGroup.POST("", languageService.CreateLanguage)
	languagesGroup.PUT(":languageId", languageService.UpdateLanguage)
}
