package mutations

import (
	"languageboostergo/auth"
	"languageboostergo/db"
	"languageboostergo/types"
	"languageboostergo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	types.ServiceConfig
}

func (service *Service) CreateMutationValue(c *gin.Context) {
	var request CreateMutationValueDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var foundMutation db.Mutation
	if _, err := utils.HandleGormError(c, service.DB.Preload("Version").First(&foundMutation, request.MutationId), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, foundMutation.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Cannot access this project"})
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

	service.DB.Create(&newMutationValue)
	c.JSON(200, newMutationValue.ToSimpleMutationValue())
}

func (service *Service) UpdateMutation(c *gin.Context) {
	mutationId, err := utils.GetRouteParam(c, "mutationId", "Mutation id is invalid")
	if err != nil {
		return
	}

	var request UpdateMutationDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updatedMutation db.Mutation
	if _, err = utils.HandleGormError(c, service.DB.First(&updatedMutation, mutationId), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, updatedMutation.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Cannot access this project"})
		return
	}

	// Check if key already exists or not
	var mutationCount int64
	service.DB.
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

	service.DB.Save(&updatedMutation)
	c.JSON(http.StatusOK, updatedMutation.ToSimpleMutation())
}

func (service *Service) GetById(c *gin.Context) {
	mutationId, err := utils.GetRouteParam(c, "mutationId", "Mutation id is invalid")
	if err != nil {
		return
	}

	var mutation db.Mutation
	if _, err = utils.HandleGormError(c, service.DB.Preload("MutationValues").First(&mutation, mutationId), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, mutation.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Cannot access this project"})
		return
	}

	c.JSON(http.StatusOK, mutation.ToSimpleMutation())
}

func (service *Service) DeleteById(c *gin.Context) {
	mutationId, err := utils.GetRouteParam(c, "mutationId", "Mutation id is invalid")
	if err != nil {
		return
	}

	var mutation db.Mutation
	if _, err = utils.HandleGormError(c, service.DB.Preload("Version").First(&mutation, mutationId), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	if !auth.IsUserInProject(userId, mutation.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You don't have access to this project"})
		return
	}

	if mutation.Version.Locked {
		c.JSON(http.StatusBadRequest, gin.H{"message": "This version is locked, you cannot delete any mutations"})
		return
	}

	service.DB.Delete(&mutation)
	c.JSON(http.StatusOK, mutation)
}

func (service *Service) SearchByProject(c *gin.Context) {
	projectId, err := utils.GetRouteParam(c, "projectId", "Project id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Cannot access this project"})
		return
	}

	foundVersion, err := GetVersionFromQuery(c, projectId, service.DB)
	if err != nil {
		return
	}

	var request SearchMutationsDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build or queries for languages
	languageConditions := service.DB.Where("1 = 0") // TODO this may not work, claude recommended it for some way
	for _, lang := range request.Languages {
		languageConditions.Or("language_id = ? AND value LIKE ?", lang.LanguageId, "%"+lang.Search+"%")
	}

	// Prefilter ID for mutations
	subQuery := service.DB.
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
	service.DB.
		Model(&db.MutationValue{}).
		Where(languageConditions).
		Where("mutation_id IN (?)", subQuery).
		Distinct("mutation_id").
		Pluck("mutation_id", &mutationIds)

	var mutations []db.Mutation
	service.DB.Where("id IN (?)", mutationIds).Preload("MutationValues").Order("key asc").Find(&mutations)

	simpleMutations := make([]db.SimpleMutation, len(mutations))
	for i, v := range mutations {
		simpleMutations[i] = v.ToSimpleMutation()
	}

	c.JSON(http.StatusOK, simpleMutations)
}

func GetVersionFromQuery(c *gin.Context, projectId uint, conn *gorm.DB) (*db.Version, error) {
	versionQuery := c.DefaultQuery("version", "latest")

	// Using the current version, apply no filters
	if versionQuery == "latest" {
		return nil, nil
	}

	// If version is specified, check if it exists
	var foundVersion db.Version
	if _, err := utils.HandleGormError(c, conn.
		Where("project_id = ?", projectId).
		Where("name = ?", versionQuery).
		First(&foundVersion),
		"Version not found"); err != nil {
		return nil, err
	}

	return &foundVersion, nil
}

func (service *Service) ListByProject(c *gin.Context) {
	projectIdParam, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Project ID is invalid"})
	}

	projectId := uint(projectIdParam)

	// Permission check
	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(http.StatusForbidden, "You are not in this project")
		return
	}

	// Version check
	foundVersion, err := GetVersionFromQuery(c, projectId, service.DB)
	if err != nil {
		return
	}

	// Create "query builder"
	qb := service.DB.Preload("MutationValues").Order("key asc").Limit(100).Where("mutations.project_id = ?", projectId)
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

	c.JSON(http.StatusOK, simpleMutations)
}

func (service *Service) UpdateMutationValue(c *gin.Context) {
	mutationValueId, err := utils.GetRouteParam(c, "mutationValueId", "Mutation value id is invalid")
	if err != nil {
		return
	}

	var request UpdateMutationValueDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updatedMutationValue db.MutationValue
	service.DB.Joins("Mutation").Joins("Mutation.Version").First(&updatedMutationValue, mutationValueId)

	foundMutation := updatedMutationValue.Mutation
	foundVersion := foundMutation.Version

	// Permission check
	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, foundMutation.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not in this project"})
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

	service.DB.Save(&updatedMutationValue)

	c.JSON(http.StatusOK, updatedMutationValue.ToSimpleMutationValue())
}

func (service *Service) CreateMutation(c *gin.Context) {
	var data CreateMutationDto
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("userId").(uint)
	if !auth.IsUserInProject(userId, data.ProjectId) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not in this project"})
		return
	}

	// Check for versions
	var foundVersion *db.Version
	if data.VersionId != nil {
		foundVersion = &db.Version{} // Initialize not-nil pointer

		if _, err := utils.HandleGormError(c, service.DB.First(foundVersion, data.VersionId), "Version not found"); err != nil {
			return
		}

		if foundVersion.Locked {
			c.JSON(http.StatusBadRequest, gin.H{"message": "This version is already locked, so you cannot create new mutations in it"})
			return
		}

	}

	var matchedMutationsCount int64
	qb := service.DB.Where("project_id = ?", data.ProjectId).Where("key = ?", data.Key)

	if foundVersion != nil {
		qb.Where("version_id = ?", foundVersion.ID)
	}

	qb.Count(&matchedMutationsCount)
	if matchedMutationsCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "Mutation with this key already exists in this version"})
		return
	}

	// Mutation does not exist yet
	// Get languages by project
	var languages []db.Language
	service.DB.Where("project_id = ?", data.ProjectId).Find(&languages)

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

	service.DB.Create(&mutation)

	c.JSON(http.StatusOK, mutation.ToSimpleMutation())
}
