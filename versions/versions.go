package versions

import (
	"errors"
	"languageboostergo/auth"
	"languageboostergo/db"
	"languageboostergo/types"
	"languageboostergo/utils"
	"net/http"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type Service struct {
	types.ServiceConfig
}

func (service *Service) DeleteVersion(c *gin.Context) {
	versionId, err := utils.GetRouteParam(c, "versionId", "Version id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	var version db.Version
	err = service.DB.Find(&version, versionId).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Version not found"})
		return
	}

	// Check if user is in project and if the project exists
	if !auth.IsUserInProject(userId, version.ProjectID) {
		c.JSON(403, "You don't have access to this version")
		return
	}

	service.DB.Delete(&version)
}

func (service *Service) PublishVersion(c *gin.Context) {
	// Get user from auth header
	userId := c.MustGet("userId").(uint)

	// Get request body
	var request types.PublishVersionDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Version name latest is reserved for current state of the mutations
	if request.Name == "latest" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create version with the name latest"})
		return
	}

	// Check if user is in project and if the project exists
	if !auth.IsUserInProject(userId, request.ProjectID) {
		c.JSON(403, "You cannot update this project")
		return
	}

	// Check if version with this name already exists
	var foundVersionsCount int64
	service.DB.Model(&db.Version{}).Where("project_id = ? AND name = ?", request.ProjectID, request.Name).Count(&foundVersionsCount)

	if foundVersionsCount != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Version with this name already exists"})
		return
	}

	version := request.ToModel()

	// Create a version and retrieve all fields
	service.DB.Create(&version)
	service.DB.Find(&version)

	// Fetch languages in the current project
	var languages []db.Language
	service.DB.Where("project_id = ?", request.ProjectID).Find(&languages)

	// No we'll select all mutations without version
	var mutations []db.Mutation
	service.DB.Where("mutations.project_id = ?", request.ProjectID).Where("mutations.version_id IS NULL").Preload("MutationValues").Order("key asc").Find(&mutations)

	// No we'll copy those mutations
	var newMutations []db.Mutation

	for _, mutation := range mutations {
		var newMutationsValues []db.MutationValue

		for _, mutationValue := range mutation.MutationValues {
			newMutationsValues = append(newMutationsValues, db.MutationValue{
				Value:      mutationValue.Value,
				LanguageId: mutationValue.LanguageId,
				Status:     mutationValue.Status,
			})
		}

		newMutations = append(newMutations, db.Mutation{
			Key:            mutation.Key,
			ProjectID:      mutation.ProjectID,
			VersionID:      &version.ID,
			Status:         mutation.Status,
			MutationValues: newMutationsValues, // create mutation values with association mode
		})
	}

	// Enable default transaction for this huge insert to keep data consistent
	service.DB.Session(&gorm.Session{SkipDefaultTransaction: false}).CreateInBatches(&newMutations, 100)

	c.JSON(http.StatusCreated, version)
}
