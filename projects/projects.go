package projects

import (
	"languageboostergo/auth"
	"languageboostergo/db"
	"languageboostergo/types"
	"languageboostergo/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)


type CreateProjectDto struct {
	Name    string `json:"name" binding:"required"`
	SpaceId uint   `json:"spaceId"`
}

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
	userInSpace := false
	for _, user := range foundSpace.Users {
		if user.ID == userId {
			userInSpace = true
			break
		}
	}

	if !userInSpace {
		c.JSON(http.StatusForbidden, "Cannot access this space")
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

	var request CreateProjectDto
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


	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.Preload("Users").Preload("Projects").First(&foundSpace, spaceId), "Space not found"); err != nil {
		return
	}

	// Check user relevance
	userId := c.MustGet("userId").(uint)
	userInSpace := false
	for _, user := range foundSpace.Users {
		if user.ID == userId {
			userInSpace = true
			break
		}
	}

	if !userInSpace {
		c.JSON(http.StatusForbidden, "Cannot access this space")
		return
	}

	c.JSON(http.StatusOK, foundSpace.ToSimpleSpace().Projects)
}
