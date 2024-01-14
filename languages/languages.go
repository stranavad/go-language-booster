package languages

import (
	"github.com/gin-gonic/gin"
	"languageboostergo/auth"
	"languageboostergo/db"
	"net/http"
	"strconv"
)

var conn = db.GetDb()

type CreateLanguageDto struct {
	ProjectId uint   `json:"projectId" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type UpdateLanguageDto struct {
	Name string `json:"name" binding:"required"`
}

func CreateLanguage(c *gin.Context) {
	var data CreateLanguageDto
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, data.ProjectId) {
		c.JSON(403, "You are not in this project")
		return
	}

	var newLanguage db.Language
	newLanguage.ProjectID = data.ProjectId
	newLanguage.Name = data.Name
	conn.Create(&newLanguage)
	c.JSON(200, newLanguage.ToSimpleLanguage())
}

func GetLanguagesByProjectId(c *gin.Context) {
	projectIdParam, err := strconv.ParseUint(c.Param("projectId"), 10, 16)
	if err != nil {
		panic("Project ID is not number serializable")
	}

	projectId := uint(projectIdParam)
	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(403, "You are not in this project")
		return
	}

	var languages []db.SimpleLanguage

	err = conn.Model(&db.Language{}).Where("project_id = ?", projectId).Find(&languages).Error
	if err != nil {
		c.JSON(500, "Internal server error")
	}

	c.JSON(200, languages)

}

func UpdateLanguage(c *gin.Context) {
	languageId, err := strconv.ParseUint(c.Param("languageId"), 10, 16)
	if err != nil {
		panic("Language ID is not number serializable")
	}

	var request UpdateLanguageDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updatedLanguage db.Language
	conn.First(&updatedLanguage, uint(languageId))

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, updatedLanguage.ProjectID) {
		c.JSON(403, "You are not in this project")
		return
	}

	if request.Name != "" {
		updatedLanguage.Name = request.Name
	}

	conn.Save(&updatedLanguage)
	c.JSON(200, updatedLanguage.ToSimpleLanguage())
}
