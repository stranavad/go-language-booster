package mutations

import (
	"github.com/gin-gonic/gin"
	"languageboostergo/auth"
	"languageboostergo/db"
	"net/http"
	"strconv"
)

var conn = db.GetDb()

type CreateMutationDto struct {
	ProjectId uint                     `json:"projectId" binding:"required"`
	Key       string                   `json:"key" binding:"required"`
	Status    string                   `json:"status" binding:"required"`
	Values    []CreateMutationDtoValue `json:"values" binding:"required"`
}

type CreateMutationDtoValue struct {
	LanguageId uint   `json:"languageId" binding:"required"`
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

type CreateMutationValueDto struct {
	Value      string `json:"value"`
	MutationId uint   `json:"mutationId"`
	LanguageId uint   `json:"languageId"`
}

type SearchMutationLanguageDto struct {
	LanguageId uint   `json:"languageId" binding:"required"`
	Search     string `json:"search" binding:"required"`
}

type SearchMutationsDto struct {
	Key       string                      `json:"key"`
	Status    string                      `json:"status"`
	Languages []SearchMutationLanguageDto `json:"languages"`
}

func CreateMutationValue(c *gin.Context) {
	var request CreateMutationValueDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var foundMutation db.Mutation
	conn.First(&foundMutation, request.MutationId)

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, foundMutation.ProjectID) {
		c.JSON(403, "You are not in this project")
		return
	}

	var newMutationValue db.MutationValue
	newMutationValue.Value = request.Value
	newMutationValue.LanguageId = request.LanguageId
	newMutationValue.MutationId = request.MutationId

	conn.Create(&newMutationValue)
	c.JSON(200, newMutationValue.ToSimpleMutationValue())
}

func UpdateMutation(c *gin.Context) {
	mutationIdParam, err := strconv.ParseUint(c.Param("mutationId"), 10, 32)
	if err != nil {
		panic("Mutation ID is not number serializable")
	}

	var request UpdateMutationDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mutationId := uint(mutationIdParam)
	var updatedMutation db.Mutation
	conn.First(&updatedMutation, mutationId)

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, updatedMutation.ProjectID) {
		c.JSON(403, "You are not in this project")
		return
	}

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

	if request.Status != "" {
		updatedMutation.Status = request.Status
	}

	conn.Save(&updatedMutation)
	c.JSON(200, updatedMutation.ToSimpleMutation())
}

func GetById(c *gin.Context) {
	mutationIdParam, err := strconv.ParseUint(c.Param("mutationId"), 10, 32)
	if err != nil {
		panic("Mutation ID is not number serializable")
	}

	mutationId := uint(mutationIdParam)
	var mutation db.Mutation
	dbErr := conn.Preload("MutationValues").First(&mutation, mutationId).Error

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, mutation.ProjectID) {
		c.JSON(403, "You are not in this project")
		return
	}

	if dbErr != nil {
		c.JSON(200, nil)
		c.Abort()
		return
	}

	c.JSON(200, mutation.ToSimpleMutation())
}

func DeleteById(c *gin.Context) {
	mutationIdParam, err := strconv.ParseUint(c.Param("mutationId"), 10, 32)
	if err != nil {
		panic("Mutation ID is not number serializable")
	}

	mutationId := uint(mutationIdParam)
	var mutation db.Mutation
	conn.First(&mutation, mutationId)

	userId := c.MustGet("userId").(uint)
	if !auth.IsUserInProject(userId, mutation.ProjectID) {
		c.JSON(403, "You are not in this project")
		return
	}

	conn.Delete(&mutation)
	c.JSON(200, mutation)
}

func SearchByProject(c *gin.Context) {
	projectIdParam, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(405, "Project ID is invalid")
		return
	}

	projectId := uint(projectIdParam)

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(403, "You are not in this project")
		return
	}

	var request SearchMutationsDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mvDb := conn
	for _, lang := range request.Languages {
		mvDb = mvDb.Or("language_id = ? AND value LIKE ?", lang.LanguageId, "%"+lang.Search+"%")
	}

	var mutationValues []db.MutationValue
	mvDb.Select("mutation_id").Find(&mutationValues)

	var mutationIds []uint
	uniqueIds := make(map[uint]bool)

	for _, value := range mutationValues {
		if _, exists := uniqueIds[value.MutationId]; !exists {
			mutationIds = append(mutationIds, value.MutationId)
			uniqueIds[value.MutationId] = true
		}
	}

	if len(mutationIds) == 0 {
		var res []db.SimpleMutation
		c.JSON(200, res)
		return
	}

	var mutations []db.Mutation
	resDb := conn.Preload("MutationValues").Order("key asc").Where(mutationIds).Where("project_id = ?", projectId)

	if request.Key != "" {
		resDb.Where("key like ?", "%"+request.Key+"%")
	}

	if request.Status != "" {
		resDb.Where("status = ?", request.Status)
	}

	resDb.Limit(100).Find(&mutations)

	simpleMutations := make([]db.SimpleMutation, len(mutations))
	for i, v := range mutations {
		simpleMutations[i] = v.ToSimpleMutation()
	}

	c.JSON(200, simpleMutations)
}

func ListByProject(c *gin.Context) {
	projectIdParam, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		panic("Project ID is not number serializable")
	}

	projectId := uint(projectIdParam)

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(403, "You are not in this project")
		return
	}

	var mutations []db.Mutation
	conn.Preload("MutationValues").Order("key asc").Limit(100).Find(&mutations, "mutations.project_id = ?", projectId)
	simpleMutations := make([]db.SimpleMutation, len(mutations))
	for i, v := range mutations {
		simpleMutations[i] = v.ToSimpleMutation()
	}
	c.JSON(200, simpleMutations)
}

func UpdateMutationValue(c *gin.Context) {
	mutationValueIdParam, err := strconv.ParseUint(c.Param("mutationValueId"), 10, 32)
	if err != nil {
		panic("Mutation Value ID is not number serializable")
	}

	var request UpdateMutationValueDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mutationValueId := uint(mutationValueIdParam)

	var updatedMutationValue db.MutationValue
	conn.First(&updatedMutationValue, mutationValueId)

	var foundMutation db.Mutation
	conn.First(&foundMutation, updatedMutationValue.MutationId)

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, foundMutation.ProjectID) {
		c.JSON(403, "You are not in this project")
		return
	}

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

	userId := c.MustGet("userId").(uint)
	if !auth.IsUserInProject(userId, data.ProjectId) {
		c.JSON(403, "You are not in this project")
		return
	}

	var mutations []db.Mutation
	conn.Where("project_id = ? AND key = ?", data.ProjectId, data.Key).Find(&mutations).Limit(1)
	if len(mutations) > 0 {
		c.JSON(405, gin.H{"message": "Mutation with this key already exists"})
		c.Abort()
		return
	}

	// Mutation does not exist yet
	// Get languages by project
	var languages []db.Language
	conn.Where("project_id = ?", data.ProjectId).Find(&languages)

	// Map languages to sent values
	mutationValues := make([]db.MutationValue, len(languages))
	for i, language := range languages {
		// Find language in values
		var foundValue *CreateMutationDtoValue
		for _, value := range data.Values {
			if value.LanguageId == language.ID {
				foundValue = &value
				break
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
		ProjectID:      data.ProjectId,
		Status:         data.Status,
		Key:            data.Key,
		MutationValues: mutationValues,
	}

	conn.Create(&mutation)

	c.JSON(200, mutation.ToSimpleMutation())
}
