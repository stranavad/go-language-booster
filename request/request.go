package request

import (
	"languageboostergo/auth"
	"languageboostergo/db"
	"languageboostergo/types"
	"languageboostergo/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Service struct {
	types.ServiceConfig
}

func (service *Service) GetRequests(c *gin.Context) {
	projectId, err := utils.GetRouteParam(c, "projectId", "Project id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot access this project"})
		return
	}

	var requests []db.Request
	if err := service.DB.Joins("User").Joins("Branch").Joins("Branch.User").Find(&requests).Error; err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error when loading requests"})
		return
	}

	response := make([]db.RequestResponse, len(requests))
	for index, request := range requests {
		response[index] = request.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (service *Service) DeleteRequest(c *gin.Context) {
	requestId, err := utils.GetRouteParam(c, "requestId", "Request id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	var foundRequest db.Request
	if _, err := utils.HandleGormError(c, service.DB.Where("id = ?", requestId).First(&foundRequest), "Request not found"); err != nil {
		return
	}

	if !auth.IsUserInProject(userId, foundRequest.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot access this project"})
		return
	}

	if err := service.DB.Delete(&foundRequest).Error; err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}

func (service *Service) CreateRequest(c *gin.Context) {
	var request CreateRequestDto
	if c.BindJSON(&request) != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	var foundBaseBranch db.Branch
	if _, err := utils.HandleGormError(c, service.DB.Where("id = ?", request.BranchID).First(&foundBaseBranch), "Base branch not found"); err != nil {
		return
	}

	if !auth.IsUserInProject(userId, foundBaseBranch.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot access this project"})
		return
	}

	// Check if the request already exists
	var foundRequestCount int64
	if err := service.DB.Model(&db.Request{}).Where("branch_id = ?", foundBaseBranch.ID).Count(&foundRequestCount).Error; err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	if foundRequestCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "Request already exists"})
		return
	}

	createdRequest := db.Request{
		Name:      request.Name,
		ProjectID: foundBaseBranch.ProjectID,
		BranchID:  foundBaseBranch.ID,
		UserID:    userId,
	}

	service.DB.Create(&createdRequest)
	service.DB.Joins("User").Joins("Branch").Find(&createdRequest)

	c.JSON(http.StatusOK, createdRequest.ToResponse())
}
