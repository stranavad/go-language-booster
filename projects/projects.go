package projects

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

func (service *Service) CreateProject(c *gin.Context) {
	var request CreateProjectDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.Preload("Users").First(&foundSpace, request.SpaceId), "Space not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	var foundSpaceMember db.SpaceMember
	if err := service.DB.Where("user_id = ?", userId).Where("space_id = ?", request.SpaceId).First(&foundSpaceMember).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot access this space"})
		return
	}

	if foundSpaceMember.Role != db.Owner && foundSpaceMember.Role != db.Admin {
		c.JSON(http.StatusForbidden, gin.H{"message": "You don't the required permissions for creating project inside this space"})
		return
	}

	newProject := db.Project{
		Name:    request.Name,
		SpaceID: foundSpace.ID,
	}

	service.DB.Create(&newProject)

	c.JSON(http.StatusCreated, newProject.ToSimpleProject())
}

func (service *Service) GetById(c *gin.Context) {
	projectId, err := utils.GetRouteParam(c, "projectId", "Project id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(http.StatusForbidden, "Cannot access this project")
		return
	}

	var foundProject db.Project
	if _, err := utils.HandleGormError(c, service.DB.First(&foundProject, projectId), "Project not found"); err != nil {
		return
	}

	c.JSON(http.StatusOK, foundProject.ToSimpleProject())
}

func (service *Service) UpdateProject(c *gin.Context) {
	projectId, err := utils.GetRouteParam(c, "projectId", "Project id is invalid")
	if err != nil {
		return
	}

	var request UpdateProjectDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(http.StatusForbidden, "Cannot access this project")
		return
	}

	var updateData db.Project
	service.DB.First(&updateData, projectId)

	if request.Name != "" {
		updateData.Name = request.Name
	}

	service.DB.Save(&updateData)
	c.JSON(http.StatusCreated, updateData.ToSimpleProject())
}

func (service *Service) ListProjects(c *gin.Context) {
	spaceId, err := utils.GetRouteParam(c, "spaceId", "Space id is invalid")
	if err != nil {
		return
	}


	userId := c.MustGet("userId").(uint)

	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.Preload("Members", "user_id = ?", userId).Preload("Projects").First(&foundSpace, spaceId), "Space not found"); err != nil {
		return
	}

	if len(foundSpace.Members) == 0 {
		c.JSON(http.StatusForbidden, "Cannot access this space")
		return
	}

	c.JSON(http.StatusOK, foundSpace.ToSimpleSpace().Projects)
}
