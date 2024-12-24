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
	if err := c.BindJSON(&request); err != nil {
		return
	}

	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.Preload("Members").First(&foundSpace, request.SpaceId), "Space not found"); err != nil {
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
		Languages: []db.Language{
			{
				Name:    request.PrimaryLanguage,
				Primary: true,
			},
		},
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
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(http.StatusForbidden, "Cannot access this project")
		return
	}

	var updateData db.Project
	if _, err := utils.HandleGormError(c, service.DB.First(&updateData, projectId), "Project not found"); err != nil {
		return
	}

	if request.Name != "" {
		updateData.Name = request.Name
	}

	service.DB.Save(&updateData)
	service.DB.First(&updateData)
	c.JSON(http.StatusCreated, updateData.ToSimpleProject())
}

func (service *Service) DeleteProject(c *gin.Context) {
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

	service.DB.Delete(&foundProject)
}

func (service *Service) ListProjects(c *gin.Context) {
	spaceId, err := utils.GetRouteParam(c, "spaceId", "Space id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	var foundSpaceMembers int64
	if err := service.DB.Model(&db.SpaceMember{}).Where("user_id = ?", userId).Where("space_id = ?", spaceId).Count(&foundSpaceMembers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong with authorizing"})
		return
	}

	if foundSpaceMembers == 0 {
		c.JSON(http.StatusForbidden, gin.H{"message": "You don't have access to this space"})
		return
	}

	var foundProjects []db.Project
	if err := service.DB.Where("space_id = ?", spaceId).Find(&foundProjects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong with fetching projects"})
		return
	}

	var parsedProjects []db.SimpleProject
	for _, project := range foundProjects {
		parsedProjects = append(parsedProjects, project.ToSimpleProject())
	}

	c.JSON(http.StatusOK, parsedProjects)
}
