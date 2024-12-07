package languages

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

func (service *Service) CreateLanguage(c *gin.Context) {
	var data CreateLanguageDto
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, data.ProjectId) {
		c.JSON(http.StatusForbidden, "Cannot access this project")
		return
	}

	var newLanguage db.Language
	newLanguage.ProjectID = data.ProjectId
	newLanguage.Name = data.Name
	service.DB.Create(&newLanguage)
	c.JSON(200, newLanguage.ToSimpleLanguage())
}

func (service *Service) GetLanguagesByProjectId(c *gin.Context) {
	projectId, err := utils.GetRouteParam(c, "projectId", "Project id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(http.StatusForbidden, "Cannot access this project")
		return
	}

	var languages []db.SimpleLanguage

	err = service.DB.Model(&db.Language{}).Where("project_id = ?", projectId).Find(&languages).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed getting languages"})
		return
	}

	c.JSON(http.StatusOK, languages)

}

func (service *Service) UpdateLanguage(c *gin.Context) {
	languageId, err := utils.GetRouteParam(c, "language", "Language id is invalid")
	if err != nil {
		return
	}

	var request UpdateLanguageDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updatedLanguage db.Language
	service.DB.First(&updatedLanguage, languageId)

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, updatedLanguage.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not in this project"})
		return
	}

	if request.Name != "" {
		updatedLanguage.Name = request.Name
	}

	service.DB.Save(&updatedLanguage)
	c.JSON(http.StatusOK, updatedLanguage.ToSimpleLanguage())
}
