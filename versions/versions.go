package versions

import (
	"languageboostergo/auth"
	"languageboostergo/db"
	"languageboostergo/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Service struct {
	types.ServiceConfig
}

func (service *Service) PublishVersion(c *gin.Context) {
	// Get user from auth header
	userId := c.MustGet("userId").(uint)

	// Get request body
	var request types.PublishVersionDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user is in project and if the project exists
	if !auth.IsUserInProject(userId, request.ProjectID) {
		c.JSON(403, "You cannot update this project")
		return
	}

	// Check if version with this name already exists
	var foundVersionsCount int64
	service.DB.Model(&db.Version{}).Where("project_id = ? AND name = ?", request.ProjectID, request.Name).Count(&foundVersionsCount)

	if foundVersionsCount != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Version with this name already exists"})
		return
	}

	version := request.ToModel()
	
	// Create a version and retrieve all fields
	service.DB.Create(&version)
	service.DB.Find(&version)
	
	// No we'll select all mutations without version
}



func (service *Service) GetVersions() {
	service.CreateVersion()
}
