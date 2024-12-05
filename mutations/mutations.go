package mutations

import (
	"errors"
	"gorm.io/gorm"
	"languageboostergo/auth"
	"languageboostergo/db"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var conn = db.GetDb()

type CreateMutationDto struct {
	ProjectId uint                     `json:"projectId" binding:"required"`
	Key       string                   `json:"key" binding:"required"`
	Status    string                   `json:"status" binding:"required"`
	Values    []CreateMutationDtoValue `json:"values" binding:"required"`
	VersionId *uint                    `json:"versionId"`
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
	Value      string `json:"value" binding:"required"`
	MutationId uint   `json:"mutationId" binding:"required"`
	LanguageId uint   `json:"languageId" binding:"required"`
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

type DBResponse struct {
	Error error
	Found bool
}

func HandleGormError(c *gin.Context, result *gorm.DB, notFoundMessage string) (bool, error) {
	if result.Error == nil {
		return true, nil
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundMessage})

		return false, result.Error
	}

	log.Printf("Database error: %v", result.Error)

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
	return false, result.Error
}

func CreateMutationValue(c *gin.Context) {
	var request CreateMutationValueDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var foundMutation db.Mutation
	if _, err := HandleGormError(c, conn.Preload("Version").First(&foundMutation, request.MutationId), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, foundMutation.ProjectID) {
		c.JSON(403, "Cannot access this project")
		return
	}

	if foundMutation.Version != nil && foundMutation.Version.Locked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This version is locked, you cannot change anything"})
		return
	}

	var newMutationValue db.MutationValue
	newMutationValue.Value = request.Value
	newMutationValue.LanguageId = request.LanguageId
	newMutationValue.MutationID = request.MutationId

	conn.Create(&newMutationValue)
	c.JSON(200, newMutationValue.ToSimpleMutationValue())
}

func UpdateMutation(c *gin.Context) {
	mutationIdParam, err := strconv.ParseUint(c.Param("mutationId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Mutation id is invalid"})
		return
	}

	var request UpdateMutationDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mutationId := uint(mutationIdParam)
	var updatedMutation db.Mutation
	if _, err = HandleGormError(c, conn.First(&updatedMutation, mutationId), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, updatedMutation.ProjectID) {
		c.JSON(403, "Cannot access this project")
		return
	}

	// Check if key already exists or not
	var mutationCount int64
	conn.
		Where("project_id = ?", updatedMutation.ProjectID).
		Where("key = ?", request.Key).
		Not("id = ?", updatedMutation.ID).
		Count(&mutationCount)

	if mutationCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Mutation with this key already exists"})
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "Mutation id is invalid"})
		return
	}

	mutationId := uint(mutationIdParam)
	var mutation db.Mutation

	if _, err = HandleGormError(c, conn.Preload("MutationValues").First(&mutation, mutationId), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, mutation.ProjectID) {
		c.JSON(403, "You are not in this project")
		return
	}

	c.JSON(200, mutation.ToSimpleMutation())
}

func DeleteById(c *gin.Context) {
	mutationIdParam, err := strconv.ParseUint(c.Param("mutationId"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Mutation id is invalid"})
		return
	}

	mutationId := uint(mutationIdParam)
	var mutation db.Mutation
	if _, err = HandleGormError(c, conn.Preload("Version").First(&mutation, mutationId), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	if !auth.IsUserInProject(userId, mutation.ProjectID) {
		c.JSON(http.StatusForbidden, "You don't have access to this project")
		return
	}

	if mutation.Version.Locked {
		c.JSON(http.StatusBadRequest, gin.H{"message": "This version is locked, you cannot delete any mutations"})
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

	foundVersion, err := GetVersionFromQuery(c, projectId)
	if err != nil {
		return
	}

	var request SearchMutationsDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build or queries for languages
	languageConditions := conn.Where("1 = 0") // TODO this may not work, claude recommended it for some way
	for _, lang := range request.Languages {
		languageConditions.Or("language_id = ? AND value LIKE ?", lang.LanguageId, "%"+lang.Search+"%")
	}

	// Prefilter ID for mutations
	subQuery := conn.
		Model(&db.Mutation{}).
		Select("id").
		Where("project_id = ?", projectId)

	if foundVersion != nil {
		subQuery.Where("version_id = ?", foundVersion.ID)
	} else {
		subQuery.Where("version_id IS NULL")
	}

	if request.Key != "" {
		subQuery.Where("key like ?", "%"+request.Key+"%")
	}

	if request.Status != "" {
		subQuery.Where("status = ?", request.Status)
	}

	var mutationIds []uint
	conn.
		Model(&db.MutationValue{}).
		Where(languageConditions).
		Where("mutation_id IN (?)", subQuery).
		Distinct("mutation_id").
		Pluck("mutation_id", &mutationIds)

	var mutations []db.Mutation
	conn.Where("id IN (?)", mutationIds).Preload("MutationValues").Order("key asc").Find(&mutations)

	simpleMutations := make([]db.SimpleMutation, len(mutations))
	for i, v := range mutations {
		simpleMutations[i] = v.ToSimpleMutation()
	}

	c.JSON(200, simpleMutations)
}

func GetVersionFromQuery(c *gin.Context, projectId uint) (*db.Version, error) {
	versionQuery := c.DefaultQuery("version", "latest")

	// Using the current version, apply no filters
	if versionQuery == "latest" {
		return nil, nil
	}

	// If version is specified, check if it exists
	var foundVersion db.Version
	if _, err := HandleGormError(c, conn.
		Where("project_id = ?", projectId).
		Where("name = ?", versionQuery).
		First(&foundVersion),
		"Version not found"); err != nil {
		return nil, err
	}

	return &foundVersion, nil
}

func ListByProject(c *gin.Context) {
	projectIdParam, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Project ID is invalid"})
	}

	projectId := uint(projectIdParam)

	// Permission check
	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(403, "You are not in this project")
		return
	}

	// Version check
	foundVersion, err := GetVersionFromQuery(c, projectId)
	if err != nil {
		return
	}

	// Create "query builder"
	qb := conn.Preload("MutationValues").Order("key asc").Limit(100).Where("mutations.project_id = ?", projectId)
	if foundVersion == nil {
		qb = qb.Where("mutations.version_id IS NULL")
	} else {
		qb = qb.Where("mutations.version_id = ?", foundVersion.ID)
	}

	var mutations []db.Mutation
	qb.Find(&mutations)

	simpleMutations := make([]db.SimpleMutation, len(mutations))
	for i, v := range mutations {
		simpleMutations[i] = v.ToSimpleMutation()
	}
	c.JSON(200, simpleMutations)
}

func UpdateMutationValue(c *gin.Context) {
	mutationValueIdParam, err := strconv.ParseUint(c.Param("mutationValueId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Mutation id is invalid"})
		return
	}

	var request UpdateMutationValueDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mutationValueId := uint(mutationValueIdParam)

	var updatedMutationValue db.MutationValue
	conn.Joins("Mutation").Joins("Mutation.Version").First(&updatedMutationValue, mutationValueId)

	foundMutation := updatedMutationValue.Mutation
	foundVersion := foundMutation.Version

	// Permission check
	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, foundMutation.ProjectID) {
		c.JSON(403, "You are not in this project")
		return
	}

	// Check if the version is not already locked
	if foundVersion != nil && foundVersion.Locked {
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot update this value since the version is already locked"})
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

	// Check for versions
	var foundVersion *db.Version
	if data.VersionId != nil {
		foundVersion = &db.Version{} // Initialize not-nil pointer

		if _, err := HandleGormError(c, conn.First(foundVersion, data.VersionId), "Version not found"); err != nil {
			return
		}

		if foundVersion.Locked {
			c.JSON(http.StatusBadRequest, gin.H{"message": "This version is already locked, so you cannot create new mutations in it"})
			return
		}

	}

	var matchedMutationsCount int64
	qb := conn.Where("project_id = ?", data.ProjectId).Where("key = ?", data.Key)

	if foundVersion != nil {
		qb.Where("version_id = ?", foundVersion.ID)
	}

	qb.Count(&matchedMutationsCount)
	if matchedMutationsCount > 0 {
		c.JSON(405, gin.H{"message": "Mutation with this key already exists in this version"})
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

	var foundVersionId *uint
	if foundVersion != nil {
		foundVersionId = &foundVersion.ID
	}

	mutation := db.Mutation{
		ProjectID:      data.ProjectId,
		Status:         data.Status,
		Key:            data.Key,
		MutationValues: mutationValues, // create with association mode
		VersionID:      foundVersionId,
	}

	conn.Create(&mutation)

	c.JSON(200, mutation.ToSimpleMutation())
}
