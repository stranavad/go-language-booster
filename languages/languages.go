package languages

import (
	"github.com/gin-gonic/gin"
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

	var newLanguage db.Language
	newLanguage.ProjectID = data.ProjectId
	newLanguage.Name = data.Name
	conn.Create(&newLanguage)
	c.JSON(200, newLanguage.ToSimpleLanguage())
}

func GetLanguagesByProjectId(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Param("projectId"), 10, 16)
	if err != nil {
		panic("Project ID is not number serializable")
	}
	var languages []db.Language

	err = conn.Where("project_id = ?", projectId).Find(&languages).Error
	if err != nil {
		c.JSON(500, "Internal server error")
	}

	simpleLanguages := make([]db.SimpleLanguage, len(languages))
	for i, v := range languages {
		simpleLanguages[i] = v.ToSimpleLanguage()
	}
	
	c.JSON(200, simpleLanguages)

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
	updatedLanguage.ID = uint(languageId)
	updatedLanguage.Name = request.Name
	conn.Save(&updatedLanguage)
	c.JSON(200, updatedLanguage.ToSimpleLanguage())
}
