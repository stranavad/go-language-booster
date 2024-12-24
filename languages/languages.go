package languages

import (
	"fmt"
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
	if err := c.BindJSON(&data); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, data.ProjectId) {
		c.JSON(http.StatusForbidden, "Cannot access this project")
		return
	}

	// Check if this language already exists
	var foundLanguageCount int64
	if err := service.DB.Model(&db.Language{}).Where("project_id = ?", data.ProjectId).Where("name = ?", data.Name).Count(&foundLanguageCount).Error; err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, "Failed creating new language")
		return
	}

	if foundLanguageCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "Language with this name already exists"})
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

	var languages []db.Language

	err = service.DB.Where("project_id = ?", projectId).Find(&languages).Error
	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed getting languages"})
		return
	}

	var simpleLanguages []db.SimpleLanguage
	for _, language := range languages {
		simpleLanguages = append(simpleLanguages, language.ToSimpleLanguage())
	}

	c.JSON(http.StatusOK, simpleLanguages)

}

func (service *Service) DeleteLanguage(c *gin.Context) {
	languageId, err := utils.GetRouteParam(c, "languageId", "Language id is invalid")
	if err != nil {
		return
	}

	var foundLanguage db.Language
	if _, err := utils.HandleGormError(c, service.DB.Where("id = ?", languageId).First(&foundLanguage), "Language not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	if !auth.IsUserInProject(userId, foundLanguage.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not in this project"})
		return
	}

	if err := service.DB.Delete(&db.MutationValue{}, "language_id = ?", foundLanguage.ID).Error; err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed deleting language"})
		return
	}

	if err := service.DB.Delete(&foundLanguage).Error; err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed deleting language"})
		return
	}
}

func (service *Service) UpdateLanguage(c *gin.Context) {
	languageId, err := utils.GetRouteParam(c, "language", "Language id is invalid")
	if err != nil {
		return
	}

	var request UpdateLanguageDto
	if err := c.BindJSON(&request); err != nil {
		return
	}

	var updatedLanguage db.Language
	if _, err := utils.HandleGormError(c, service.DB.Where("id = ?", languageId).First(&updatedLanguage), "Language not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	if !auth.IsUserInProject(userId, updatedLanguage.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not in this project"})
		return
	}

	var foundLanguages int64
	if err := service.DB.Model(&db.Language{}).Where("project_id = ?", updatedLanguage.ProjectID).Where("name = ?", request.Name).Not("id = ?", languageId).Count(&foundLanguages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed updating language"})
		return
	}

	if foundLanguages > 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "Language with this name already exists"})
		return
	}

	if request.Name != "" {
		updatedLanguage.Name = request.Name
	}

	service.DB.Save(&updatedLanguage)
	service.DB.First(&updatedLanguage)
	c.JSON(http.StatusOK, updatedLanguage.ToSimpleLanguage())
}
