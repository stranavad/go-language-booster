package branch

import (
	"github.com/gin-gonic/gin"
	"languageboostergo/types"
	"languageboostergo/utils"
)

func RegisterRoutes(r *gin.Engine, config types.ServiceConfig) {
	branchService := Service{ServiceConfig: config}
	branchGroup := r.Group("/branches")
	branchGroup.Use(utils.AuthMiddleware())
	branchGroup.POST("", branchService.CreateBranch)
	branchGroup.GET(":projectId", branchService.GetBranches)
	branchGroup.DELETE(":branchId", branchService.DeleteBranch)
}
