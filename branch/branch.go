package branch

import (
	"errors"
	"languageboostergo/auth"
	"languageboostergo/db"
	"languageboostergo/types"
	"languageboostergo/utils"
	"net/http"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type Service struct {
	types.ServiceConfig
}

func (service *Service) DeleteBranch(c *gin.Context) {
	branchId, err := utils.GetRouteParam(c, "branchId", "Branch id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	var branch db.Branch
	err = service.DB.Find(&branch, branchId).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Branch not found"})
		return
	}

	// Check if user is in project and if the project exists
	if !auth.IsUserInProject(userId, branch.ProjectID) {
		c.JSON(403, "You don't have access to this branch")
		return
	}

	service.DB.Delete(&branch)
}

func (service *Service) CreateBranch(c *gin.Context) {
	// Get user from auth header
	userId := c.MustGet("userId").(uint)

	// Get request body
	var request PublishBranchDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Branch name latest is reserved for current state of the mutations
	if request.Name == "latest" {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot create branch with the name latest"})
		return
	}

	// Check if user is in project and if the project exists
	if !auth.IsUserInProject(userId, request.ProjectID) {
		c.JSON(403, "You cannot update this project")
		return
	}

	// Check if branch with this name already exists
	var foundBranchesCount int64
	service.DB.Model(&db.Branch{}).Where("project_id = ? AND name = ?", request.ProjectID, request.Name).Count(&foundBranchesCount)

	if foundBranchesCount != 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "Branch with this name already exists"})
		return
	}

	branch := request.ToModel()
	branch.UserID = userId

	// Create a branch and retrieve all fields
	service.DB.Create(&branch)
	service.DB.Find(&branch)

	c.JSON(http.StatusCreated, branch)
}
