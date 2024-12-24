package spaces

import (
	"errors"
	"fmt"
	"languageboostergo/db"
	"languageboostergo/types"
	"languageboostergo/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	types.ServiceConfig
}

func (service *Service) CreateSpace(c *gin.Context) {
	var request CreateSpaceDto

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var user db.User
	if _, err := utils.HandleGormError(c, service.DB.First(&user, c.MustGet("userId").(uint)), "User not found"); err != nil {
		return
	}

	newSpace := db.Space{
		Name: request.Name,
	}

	service.DB.Create(&newSpace)
	newSpaceMember := db.SpaceMember{
		UserID:    user.ID,
		SpaceID:   newSpace.ID,
		Role:      db.Owner,
		CreatedAt: time.Now(),
	}
	service.DB.Create(&newSpaceMember)

	c.JSON(200, newSpace.ToSimpleSpace())
}

func (service *Service) GetSpaceMembers(c *gin.Context) {
	spaceId, err := utils.GetRouteParam(c, "spaceId", "Space id is invalid")
	if err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	var foundSpaceMembers int64
	if err := service.DB.Model(&db.SpaceMember{}).Where("user_id = ?", userId).Where("space_id = ?", spaceId).Count(&foundSpaceMembers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error when authorizing space"})
		return
	}

	if foundSpaceMembers == 0 {
		c.JSON(http.StatusForbidden, gin.H{"message": "You don't have access to this space"})
		return
	}

	var members []db.SpaceMember
	if err := service.DB.Where("space_id = ?", spaceId).Joins("User").Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error when fetching space members"})
		return
	}

	var simpleMembers []db.SpaceMemberResponse
	for _, member := range members {
		simpleMembers = append(simpleMembers, member.ToResponse())
	}

	c.JSON(http.StatusOK, simpleMembers)
}

func (service *Service) GetById(c *gin.Context) {
	spaceId, err := utils.GetRouteParam(c, "spaceId", "Space id is invalid")
	if err != nil {
		return
	}
	userId := c.MustGet("userId").(uint)

	var foundSpaceMembers int64
	if err := service.DB.Model(&db.SpaceMember{}).Where("user_id = ?", userId).Where("space_id = ?", spaceId).Count(&foundSpaceMembers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error when authorizing space"})
		return
	}

	if foundSpaceMembers == 0 {
		c.JSON(http.StatusForbidden, gin.H{"message": "You don't have access to this space"})
		return
	}

	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.First(&foundSpace, spaceId), "Space not found"); err != nil {
		return
	}

	c.JSON(200, foundSpace.ToSimpleSpace())
}

func (service *Service) UpdateSpace(c *gin.Context) {
	spaceId, err := utils.GetRouteParam(c, "spaceId", "Space id is invalid")
	if err != nil {
		return
	}

	var request CreateSpaceDto
	if err := c.BindJSON(&request); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	var foundSpaceMember db.SpaceMember
	if err := service.DB.Where("user_id = ?", userId).Where("space_id = ?", spaceId).First(&foundSpaceMember).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, gin.H{"message": "You don't have access to this space"})
		} else {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error when authorizing space"})
		}
		return
	}

	if foundSpaceMember.Role != db.Admin && foundSpaceMember.Role != db.Owner {
		c.JSON(http.StatusForbidden, gin.H{"message": "You don't have enough permissions to update this space"})
		return
	}

	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.First(&foundSpace, spaceId), "Space not found"); err != nil {
		return
	}

	foundSpace.Name = request.Name

	if err := service.DB.Save(&foundSpace).Error; err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error when updating space"})
		return
	}

	c.JSON(http.StatusOK, foundSpace.ToSimpleSpace())

}

func (service *Service) LeaveSpace(c *gin.Context) {
	spaceId, err := utils.GetRouteParam(c, "spaceId", "Space id is invalid")
	if err != nil {
		return
	}

	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.First(&foundSpace, spaceId), "Space not found"); err != nil {
		return
	}

	var foundUser db.User
	if _, err := utils.HandleGormError(c, service.DB.First(&foundUser, c.MustGet("userId").(uint)), "User not found"); err != nil {
		return
	}

	spaceErr := service.DB.Model(&foundUser).Association("Spaces").Delete([]db.Space{foundSpace})
	if spaceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": spaceErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully removed you from this space"})
}

func (service *Service) ListUserSpaces(c *gin.Context) {
	userId := c.MustGet("userId").(uint)

	var spaceMembers []db.SpaceMember
	if err := service.DB.Joins("Space").Where("user_id = ?", userId).Find(&spaceMembers).Error; err != nil {
		fmt.Println(err.Error())
		return
	}

	simpleSpaces := make([]db.SimpleSpace, len(spaceMembers))
	for i, member := range spaceMembers {
		simpleSpaces[i] = member.Space.ToSimpleSpace()
	}

	c.JSON(http.StatusOK, simpleSpaces)
}

func CheckRole(role string) error {
	if role != db.Viewer && role != db.Editor && role != db.Admin && role != db.Owner {
		return errors.New("role is invalid")
	}

	return nil
}

func (service *Service) DeleteMember(c *gin.Context) {
	var request DeleteMemberDto
	if err := c.BindJSON(&request); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)

	var foundSpaceMember db.SpaceMember
	if err := service.DB.Where("user_id = ?", userId).Where("space_id = ?", request.SpaceID).First(&foundSpaceMember).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, gin.H{"message": "You don't have access to this space"})
		} else {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error when authorizing space"})
		}
		return
	}

	if foundSpaceMember.Role != db.Admin && foundSpaceMember.Role != db.Owner {
		c.JSON(http.StatusForbidden, gin.H{"message": "You don't have enough permissions to delete other members"})
		return
	}

	var foundTargetMember db.SpaceMember
	if _, err := utils.HandleGormError(c, service.DB.Where("user_id = ?", request.UserID).Where("space_id = ?", request.SpaceID).First(&foundTargetMember), "User not found"); err != nil {
		return
	}

	if foundTargetMember.Role == db.Owner {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot delete the owner of this space"})
		return
	}

	if err := service.DB.Delete(&foundTargetMember).Error; err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed deleting member"})
		return
	}
}

func (service *Service) UpdateMember(c *gin.Context) {
	var request UpdateMemberDto
	if err := c.BindJSON(&request); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	// First check if the current user is admin and present in the space
	var foundCurrentMember db.SpaceMember
	if err := service.DB.Where("space_id = ?", request.SpaceID).Where("user_id = ?", userId).First(&foundCurrentMember).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "You don't have access to this space"})
		return
	}

	if foundCurrentMember.Role != db.Admin && foundCurrentMember.Role != db.Owner {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot update other members role"})
		return
	}

	// Create new space member
	if err := CheckRole(request.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if request.Role == db.Owner {
		c.JSON(http.StatusConflict, gin.H{"message": "This space already has an owner, you can transfer the ownership to a new person"})
		return
	}

	var foundSpaceMember db.SpaceMember
	if _, err := utils.HandleGormError(c, service.DB.Where("user_id = ?", request.UserID).Where("space_id = ?", request.SpaceID).First(&foundSpaceMember), "Member not found"); err != nil {
		return
	}

	foundSpaceMember.Role = request.Role
	service.DB.Save(&foundSpaceMember)
}

func (service *Service) AddUserToSpace(c *gin.Context) {
	var request AddUserToSpaceDto
	if err := c.BindJSON(&request); err != nil {
		return
	}

	spaceId := request.SpaceID
	userId := c.MustGet("userId").(uint)

	// First check if the current user is in this space and has appropriate roles
	var foundSpaceMember db.SpaceMember
	err := service.DB.Where("user_id = ?", userId).Where("space_id = ?", spaceId).First(&foundSpaceMember).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not part of this space"})
		return
	} else if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	} else if foundSpaceMember.Role != db.Admin && foundSpaceMember.Role != db.Owner {
		c.JSON(http.StatusForbidden, gin.H{"message": "You don't have enough permission to invite another users"})
		return
	}

	// Target user email
	targetUserEmail := request.Email
	var foundUser db.User
	if _, err := utils.HandleGormError(c, service.DB.Preload("SpaceMembers", "space_id = ?", spaceId).Where("email = ?", targetUserEmail).First(&foundUser), "User not found"); err != nil {
		return
	}

	// Check if target user in this space already
	if len(foundUser.SpaceMembers) > 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "User is already in this space"})
		return

	}
	// Create new space member
	if err = CheckRole(request.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if request.Role == db.Owner {
		c.JSON(http.StatusConflict, gin.H{"message": "This space already has an owner, you can transfer the ownership to a new person"})
		return
	}

	// Check role
	newSpaceMember := db.SpaceMember{
		UserID:    foundUser.ID,
		SpaceID:   foundSpaceMember.SpaceID,
		CreatedAt: time.Now(),
		Role:      request.Role,
	}

	service.DB.Create(&newSpaceMember)
}
