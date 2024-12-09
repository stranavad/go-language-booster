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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		UserID: user.ID,
		SpaceID: newSpace.ID,
		Role: db.Owner,
		CreatedAt: time.Now(),
	}
	service.DB.Create(&newSpaceMember)

	c.JSON(200, newSpace.ToSimpleSpace())
}


func (service *Service) GetById(c *gin.Context) {
	spaceId, err := utils.GetRouteParam(c, "spaceId", "Space id is invalid")
	if err != nil {
		return
	}
	userId := c.MustGet("userId").(uint)

	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.
		Joins("Members").
		Joins("Projects").
		First(&foundSpace, spaceId), "Space not found"); err != nil {
		return
	}

	userIsInSpace := false
	for _, member := range foundSpace.Members {
		if member.UserID == userId {
			userIsInSpace = true
			break
		}
	}

	if !userIsInSpace {
		c.JSON(http.StatusForbidden, gin.H{"message": "Cannot access this space"})
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
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.Joins("Members").First(&foundSpace, spaceId), "Space not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	userInSpace := false
	for _, member := range foundSpace.Members {
		if member.UserID == userId {
			userInSpace = true
			break
		}
	}

	if !userInSpace {
		c.JSON(http.StatusForbidden, gin.H{"message": "Cannot access this space"})
		return
	}


	if _, err := utils.HandleGormError(c, service.DB.Model(&foundSpace).Updates(db.Space{Name: request.Name}), ""); err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": spaceErr.Error()})
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
		return errors.New("Role is invalid")
	}

	return nil
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
	if errors.Is(err, gorm.ErrRecordNotFound){
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not part of this space"})
		return
	} else if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	} else if foundSpaceMember.Role != db.Admin && foundSpaceMember.Role != db.Owner {
		c.JSON(http.StatusForbidden, gin.H{"message": "Only admins and owner can invite someone else to the space"})
		return
	}

	// Target user username
	targetUserUsername := request.Username
	var foundUser db.User
	if _, err := utils.HandleGormError(c, service.DB.Joins("SpaceMembers").Where("username = ?", targetUserUsername).First(&foundUser), "User not found"); err != nil {
		return
	}

	// Check if target user in this space already
	userIsInSpace := false
	for _, member := range foundUser.SpaceMembers {
		if member.SpaceID == spaceId {
			userIsInSpace = true
			break
		}
	}

	if userIsInSpace {
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
	newSpaceMember := db.SpaceMember {
		UserID: foundUser.ID,
		SpaceID: foundSpaceMember.SpaceID,
		CreatedAt: time.Now(),
		Role: request.Role,
	}

	service.DB.Create(&newSpaceMember)
}
