package mutations

import (
	"github.com/gin-gonic/gin"
	"languageboostergo/db"
	"net/http"
	"strconv"
)

var conn = db.GetDb()

type CreateMutationDto struct {
	ProjectID uint                     `json:"projectId" binding:"required"`
	Key       string                   `json:"key" binding:"required"`
	Status    string                   `json:"status" binding:"required"`
	Values    []CreateMutationDtoValue `json:"values" binding:"required"`
}

type CreateMutationDtoValue struct {
	LanguageID uint   `json:"languageId" binding:"required"`
	Value      string `json:"value" binding:"required"`
	Status     string `json:"status" binding:"required"`
}

type UpdateMutationDto struct {
	Key    string `json:"key"`
	Status string `json:"status"`
}

type UpdateMutationValueDto struct {
	Value  string `json:"value"`
	Status string `json:"status"`
}

func UpdateMutation(c *gin.Context) {
	mutationId, err := strconv.ParseUint(c.Param("mutationId"), 10, 16)
	if err != nil {
		panic("Mutation ID is not number serializable")
	}

	var request UpdateMutationDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var updatedMutation db.Mutation
	updatedMutation.ID = uint(mutationId)
	conn.First(&updatedMutation)

	// Check if key already exists or not
	var mutations []db.Mutation
	conn.Where("project_id = ? AND key = ?", updatedMutation.ProjectID, request.Key).Find(&mutations).Limit(1)
	if len(mutations) > 0 && mutations[0].ID != updatedMutation.ID {
		c.JSON(405, gin.H{"message": "Mutation with this key already exists"})
		c.Abort()
		return
	}

	if request.Key != "" {
		updatedMutation.Key = request.Key
	}

	if request.Key != "" {
		updatedMutation.Status = request.Status
	}

	conn.Save(&updatedMutation)
	c.JSON(200, updatedMutation.ToSimpleMutation())
}

func ListByProject(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Param("projectId"), 10, 16)
	if err != nil {
		panic("Project ID is not number serializable")
	}
	var mutations []db.Mutation
	conn.Preload("MutationValues").Find(&mutations, "mutations.project_id = ?", uint(projectId))
	simpleMutations := make([]db.SimpleMutation, len(mutations))
	for i, v := range mutations {
		simpleMutations[i] = v.ToSimpleMutation()
	}
	c.JSON(200, simpleMutations)
}

func UpdateMutationValue(c *gin.Context) {
	mutationValueId, err := strconv.ParseUint(c.Param("mutationValueId"), 10, 16)
	if err != nil {
		panic("Mutation Value ID is not number serializable")
	}

	var request UpdateMutationValueDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var updatedMutationValue db.MutationValue
	updatedMutationValue.ID = uint(mutationValueId)
	conn.First(&updatedMutationValue)

	if request.Value != "" {
		updatedMutationValue.Value = request.Value
	}

	if request.Status != "" {
		updatedMutationValue.Status = request.Status
	}

	conn.Save(&updatedMutationValue)

	c.JSON(200, updatedMutationValue.ToSimpleMutationValue())
}

func CreateMutation(c *gin.Context) {
	var data CreateMutationDto
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var mutations []db.Mutation
	conn.Where("project_id = ? AND key = ?", data.ProjectID, data.Key).Find(&mutations).Limit(1)
	if len(mutations) > 0 {
		c.JSON(405, gin.H{"message": "Mutation with this key already exists"})
		c.Abort()
		return
	}

	// Mutation does not exist yet
	// Get languages by project
	var languages []db.Language
	conn.Where("project_id = ?", data.ProjectID).Find(&languages)

	// Map languages to sent values
	mutationValues := make([]db.MutationValue, len(languages))
	for i, language := range languages {
		// Find language in values
		var foundValue *CreateMutationDtoValue
		for _, value := range data.Values {
			if value.LanguageID == language.ID {
				foundValue = &value
			}
		}

		if foundValue != nil {
			mutationValues[i] = db.MutationValue{
				LanguageId: language.ID,
				Value:      foundValue.Value,
				Status:     foundValue.Status,
			}
		} else {
			mutationValues[i] = db.MutationValue{
				LanguageId: language.ID,
				Value:      "",
				Status:     "",
			}
		}
	}

	mutation := db.Mutation{
		ProjectID:      data.ProjectID,
		Status:         data.Status,
		Key:            data.Key,
		MutationValues: mutationValues,
	}

	conn.Create(&mutation)

	c.JSON(200, mutation.ToSimpleMutation())
}
