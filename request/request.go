package request

import (
	"languageboostergo/db"
	"languageboostergo/types"
	"languageboostergo/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)


type Service struct {
	types.ServiceConfig
}

func (service *Service) GetRequestById(c *gin.Context){
	requestId, err := utils.GetRouteParam(c, "requestId", "Request id is invalid")
	if err != nil {
		return
	}


	var foundRequest db.Request
	if _, err := utils.HandleGormError(c, service.DB.Where("id = ?", requestId).First(&foundRequest), "Request not found"); err != nil {
		return
	}

	var baseMutations []db.Mutation
	var targetMutations []db.Mutation

	// We use unscoped here to retrieve deleted mutations
	service.DB.Unscoped().Where("project_id = ?", foundRequest.ProjectID).Where("branch_id = ?", foundRequest.BaseBranchID).Preload("MutationVaues").Find(&baseMutations)
	targetQb := service.DB.Unscoped().Where("project_id = ?", foundRequest.ProjectID).Preload("MutationValues")

	if foundRequest.TargetBranchID != nil {
		targetQb.Where("branch_id = ?", foundRequest.TargetBranchID)
	} else {
		targetQb.Where("branch_id IS null")
	}

	targetQb.Find(&targetMutations)
}

func (service *Service) GetRequests(c *gin.Context){
	projectId, err := utils.GetRouteParam(c, "projectId", "Project id is invalid")
	if err != nil {
		return
	}

	var requests []db.Request
	if err := service.DB.Where("project_id = ?", projectId).Find(&requests).Error; err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error when loading requests"})
		return
	}

	c.JSON(http.StatusOK, requests)
}


func(service *Service) CreateRequest(c *gin.Context){
	var request CreateRequestDto
	if c.BindJSON(&request) != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	var foundProject db.Project
	if _, err := utils.HandleGormError(c, service.DB.Where("id = ?", request.ProjectID).First(&foundProject), "Project not found"); err != nil {
		return
	}

	var foundBaseBranch db.Branch
	if _, err := utils.HandleGormError(c, service.DB.Where("id = ?", request.BaseBranchID).Where("project_id = ?", foundProject.ID).First(&foundBaseBranch), "Base branch not found"); err != nil {
		return
	}

	var foundTargetBranch *db.Branch
	if request.TargetBranchID != nil {
		if _, err := utils.HandleGormError(c, service.DB.Where("id = ?", request.TargetBranchID).Where("project_id = ?", foundProject.ID).First(&foundTargetBranch), "Target branch not found"); err != nil {
			return
		}
	}

	var targetBranchId *uint
	if foundTargetBranch != nil {
		targetBranchId = &foundTargetBranch.ID
	}


	createdRequest := db.Request{
		Name: request.Name,
		ProjectID: foundProject.ID,
		BaseBranchID: foundBaseBranch.ID,
		TargetBranchID: targetBranchId,
		UserID: userId,
	}

	service.DB.Create(&createdRequest)
	service.DB.Find(&createdRequest)

	c.JSON(http.StatusOK, createdRequest)
}
