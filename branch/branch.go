package branch

import (
	"gorm.io/gorm"
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

func (service *Service) DeleteBranch(c *gin.Context) {
	branchId, err := utils.GetRouteParam(c, "branchId", "Branch id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	var branch db.Branch
	if _, err := utils.HandleGormError(c, service.DB.Find(&branch, branchId), "Branch not found"); err != nil {
		return
	}

	// Check if user is in project and if the project exists
	if !auth.IsUserInProject(userId, branch.ProjectID) {
		c.JSON(403, "You don't have access to this branch")
		return
	}

	if err := service.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&db.Mutation{}, "branch_id = ?", branch.ID).Error; err != nil {
			return err
		}

		if err := tx.Delete(&db.Request{}, "branch_id = ?", branch.ID).Error; err != nil {
			return err
		}

		if err := tx.Delete(&branch).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed deleting branch"})
		return
	}
}

func (service *Service) CreateBranch(c *gin.Context) {
	// Get user from auth header
	userId := c.MustGet("userId").(uint)

	// Get request body
	var request CreateBranchDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Branch name latest is reserved for current state of the mutations
	if request.Name == "latest" {
		c.JSON(http.StatusConflict, gin.H{"message": "Cannot create branch with the name latest"})
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

	createdBranch := db.Branch{
		Name:      request.Name,
		ProjectID: request.ProjectID,
		UserID:    userId,
	}

	// Create a branch and retrieve all fields
	service.DB.Create(&createdBranch)
	service.DB.Joins("User").Find(&createdBranch)

	c.JSON(http.StatusCreated, createdBranch.ToResponse())
}

func (service *Service) GetBranches(c *gin.Context) {
	projectId, err := utils.GetRouteParam(c, "projectId", "Project id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	// Check if user is in project and if the project exists
	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(403, "You cannot update this project")
		return
	}

	var branches []db.Branch
	if err := service.DB.Joins("User").Where("project_id = ?", projectId).Find(&branches).Error; err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed loading project branches"})
		return
	}

	response := make([]db.BranchResponse, len(branches))
	for index, branch := range branches {
		response[index] = branch.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}
