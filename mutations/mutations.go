package mutations

import (
	"errors"
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
	if _, err := utils.HandleGormError(c, service.DB.Preload("Branch").Where("id = ?", request.MutationId).First(&foundMutation), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, foundMutation.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Cannot access this project"})
		return
	}

	if foundMutation.Branch != nil && foundMutation.Branch.Locked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This branch is locked, you cannot change anything"})
		return
	}

	var newMutationValue db.MutationValue
	newMutationValue.Value = request.Value
	newMutationValue.LanguageId = request.LanguageId
	newMutationValue.MutationID = request.MutationId

	service.DB.Create(&newMutationValue)
	c.JSON(http.StatusOK, newMutationValue.ToSimpleMutationValue())
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

	var foundMutation db.Mutation
	if _, err = utils.HandleGormError(c, service.DB.First(&foundMutation, mutationId), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	if !auth.IsUserInProject(userId, foundMutation.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Cannot access this project"})
		return
	}


	// Check if the branch is locked and whether the user can update it
	foundBranch, err := service.CanUpdateBranch(c, foundMutation.ProjectID, foundMutation.BranchID, userId)
	if err != nil {
		return
	}

	// Check if key already exists or not
	var mutationCount int64
	conn := service.DB.
		Where("project_id = ?", foundMutation.ProjectID).
		Where("key = ?", request.Key).
		Not("id = ?", foundMutation.ID)

	if foundBranch != nil {
		conn.Where("branch_id = ?", foundBranch.ID)
	} else {
		conn.Where("branch_id IS null")
	}

	conn.Count(&mutationCount)

	if mutationCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Mutation with this key already exists"})
		return
	}

	if request.Key != "" {
		foundMutation.Key = request.Key
	}


	service.DB.Save(&foundMutation)
	c.JSON(http.StatusOK, foundMutation.ToSimpleMutation())
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
	if _, err = utils.HandleGormError(c, service.DB.First(&mutation, mutationId), "Mutation not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	if !auth.IsUserInProject(userId, mutation.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You don't have access to this project"})
		return
	}

	if _, err := service.CanUpdateBranch(c, mutation.ProjectID, mutation.BranchID, userId); err != nil {
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

	foundBranch, err := GetBranchFromQuery(c, projectId, service.DB)
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

	if foundBranch != nil {
		subQuery.Where("branch_id = ?", foundBranch.ID)
	} else {
		subQuery.Where("branch_id IS NULL")
	}

	if request.Key != "" {
		subQuery.Where("key like ?", "%"+request.Key+"%")
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

func GetBranchFromQuery(c *gin.Context, projectId uint, conn *gorm.DB) (*db.Branch, error) {
	branchQuery := c.DefaultQuery("branch", "latest")

	// Using the current branch, apply no filters
	if branchQuery == "latest" {
		return nil, nil
	}

	// If branch is specified, check if it exists
	var foundBranch db.Branch
	if _, err := utils.HandleGormError(c, conn.
		Where("project_id = ?", projectId).
		Where("name = ?", branchQuery).
		First(&foundBranch),
		"Branch not found"); err != nil {
		return nil, err
	}

	return &foundBranch, nil
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

	// Branch check
	foundBranch, err := GetBranchFromQuery(c, projectId, service.DB)
	if err != nil {
		return
	}

	// Create "query builder"
	qb := service.DB.Preload("MutationValues").Order("key asc").Limit(100).Where("mutations.project_id = ?", projectId)
	if foundBranch == nil {
		qb = qb.Where("mutations.branch_id IS NULL")
	} else {
		qb = qb.Where("mutations.branch_id = ?", foundBranch.ID)
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
	service.DB.Joins("Mutation").First(&updatedMutationValue, mutationValueId)

	foundMutation := updatedMutationValue.Mutation

	// Permission check
	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, foundMutation.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not in this project"})
		return
	}

	_, err = service.CanUpdateBranch(c, foundMutation.ProjectID, foundMutation.BranchID, userId)
	if err != nil {
		return
	}

	updatedMutationValue.Value = request.Value

	service.DB.Save(&updatedMutationValue)

	c.JSON(http.StatusOK, updatedMutationValue.ToSimpleMutationValue())
}

func (service *Service) CanUpdateBranch(c *gin.Context, projectId uint, branchId *uint, userId uint) (*db.Branch,error) {
	var foundSpaceMember db.SpaceMember
	if _, err := utils.HandleGormError(c, service.DB.Where("user_id = ?", userId).Where("space_id = ?", service.DB.Model(&db.Project{}).Select("space_id").Where("id = ?", projectId)).First(&foundSpaceMember), "Space member not found"); err != nil {
		return nil, err
	}

	if foundSpaceMember.Role == db.Viewer {
		err := errors.New("You don't have enough permissions to update this branch")
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return nil, err
	}

	// Check for branchs
	var foundBranch *db.Branch
	if branchId != nil {
		foundBranch = &db.Branch{} // Initialize not-nil pointer

		if _, err := utils.HandleGormError(c, service.DB.Where("id = ?", branchId).First(foundBranch), "Branch not found"); err != nil {
			return nil, err
		}

		if foundBranch.Locked {
			err := errors.New("Branch is already locked")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return nil, err
		}
	}

	// We need to check if main branch editing is enabled for the current user role
	if foundBranch == nil {
		var foundProjectSettings db.ProjectSettings
		if err := service.DB.Where("project_id = ?", projectId).First(&foundProjectSettings).Error; err != nil {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error when getting project settings"})
			return nil, err
		}

		// If the project has disabled current branch edit
		// Or if the user doesn't have required role
		if foundProjectSettings.EditCurrentBranchRole == nil || !utils.HasRequiredRole(foundSpaceMember.Role, *foundProjectSettings.EditCurrentBranchRole) {
			err := errors.New("You don't have enough permissions to update the main branch")
			c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
			return nil, err
		}
	}

	return foundBranch, nil
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

	foundBranch, err := service.CanUpdateBranch(c, data.ProjectId, data.BranchId, userId)
	if err != nil {
		return
	}

	var matchedMutationsCount int64
	qb := service.DB.Where("project_id = ?", data.ProjectId).Where("key = ?", data.Key)

	if foundBranch != nil {
		qb.Where("branch_id = ?", foundBranch.ID)
	} else {
		qb.Where("branch_id IS null")
	}

	qb.Count(&matchedMutationsCount)
	if matchedMutationsCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "Mutation with this key already exists in this branch"})
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
			}
		} else {
			mutationValues[i] = db.MutationValue{
				LanguageId: language.ID,
				Value:      "",
			}
		}
	}

	var foundBranchId *uint
	if foundBranch != nil {
		foundBranchId = &foundBranch.ID
	}

	mutation := db.Mutation{
		ProjectID:      data.ProjectId,
		Key:            data.Key,
		MutationValues: mutationValues, // create with association mode
		BranchID:      foundBranchId,
	}

	service.DB.Create(&mutation)

	c.JSON(http.StatusOK, mutation.ToSimpleMutation())
}
