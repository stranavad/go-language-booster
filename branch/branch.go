package branch

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

func (service *Service) DeleteBranch(c *gin.Context) {
	branchId, err := utils.GetRouteParam(c, "branchId", "Branch id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	var branch db.Branch
	err = service.DB.Find(&branch, branchId).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Branch not found"})
		return
	}

	// Check if user is in project and if the project exists
	if !auth.IsUserInProject(userId, branch.ProjectID) {
		c.JSON(403, "You don't have access to this branch")
		return
	}

	service.DB.Delete(&branch)
}

func (service *Service) CreateBranch(c *gin.Context) {
	// Get user from auth header
	userId := c.MustGet("userId").(uint)

	// Get request body
	var request PublishBranchDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Branch name latest is reserved for current state of the mutations
	if request.Name == "latest" {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot create branch with the name latest"})
		return
	}

	// Check if user is in project and if the project exists
	if !auth.IsUserInProject(userId, request.ProjectID) {
		c.JSON(403, "You cannot update this project")
		return
	}

	// Check if branch with this name already exists
	var foundBranchsCount int64
	service.DB.Model(&db.Branch{}).Where("project_id = ? AND name = ?", request.ProjectID, request.Name).Count(&foundBranchsCount)

	if foundBranchsCount != 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "Branch with this name already exists"})
		return
	}

	branch := request.ToModel()
	branch.UserID = userId

	var baseBranch *db.Branch
	if request.BaseBranchID != nil {
		if err := service.DB.Where("id = ?", request.BaseBranchID).Where("project_id = ?", request.ProjectID).First(&baseBranch).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound){
				c.JSON(http.StatusNotFound, gin.H{"message": "Base branch not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error when loading base branch"})
			}

			return
		}
	}

	// Keep base branch ID for structure
	if baseBranch != nil {
		branch.BaseBranchID = &baseBranch.ID
	}

	// Create a branch and retrieve all fields
	service.DB.Create(&branch)
	service.DB.Find(&branch)

	// Fetch languages in the current project
	var languages []db.Language
	service.DB.Where("project_id = ?", request.ProjectID).Find(&languages)

	// No we'll select all mutations without branch
	var mutations []db.Mutation
	qb := service.DB.Where("mutations.project_id = ?", request.ProjectID).Preload("MutationValues").Order("key asc")

	if baseBranch != nil {
		qb.Where("mutations.branch_id = ?", baseBranch.ID)
	} else {
		qb.Where("mutations.branch_id IS nul")
	}

	qb.Find(&mutations)

	// No we'll copy those mutations
	var newMutations []db.Mutation

	for _, mutation := range mutations {
		var newMutationsValues []db.MutationValue

		for _, mutationValue := range mutation.MutationValues {
			newMutationsValues = append(newMutationsValues, db.MutationValue{
				Value:      mutationValue.Value,
				LanguageId: mutationValue.LanguageId,
				Model: gorm.Model{
					UpdatedAt: mutationValue.UpdatedAt, // for keeping track with git
					CreatedAt: mutationValue.UpdatedAt, // for keeping track of original updated value for git
				},
			})
		}

		newMutations = append(newMutations, db.Mutation{
			Key:            mutation.Key,
			ProjectID:      mutation.ProjectID,
			BranchID:      &branch.ID,
			MutationValues: newMutationsValues, // create mutation values with association mode
		})
	}

	// Enable default transaction for this huge insert to keep data consistent
	service.DB.Session(&gorm.Session{SkipDefaultTransaction: false}).CreateInBatches(&newMutations, 100)

	c.JSON(http.StatusCreated, branch)
}
