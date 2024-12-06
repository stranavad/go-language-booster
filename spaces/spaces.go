package spaces

import (
	"languageboostergo/db"
	"languageboostergo/types"
	"languageboostergo/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)


type Service struct {
	types.ServiceConfig
}


type CreateSpaceDto struct {
	Name string `json:"name" binding:"required"`
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

	newSpace.Users = append(newSpace.Users, user)
	service.DB.Save(&newSpace)

	c.JSON(200, newSpace.ToSimpleSpace())
}

type UpdateSpaceDto struct {
	Name string `json:"name"`
}

func (service *Service) GetById(c *gin.Context) {
	spaceId, err := utils.GetRouteParam(c, "spaceId", "Space id is invalid")
	if err != nil {
		return
	}
	userId := c.MustGet("userId").(uint)

	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.Preload("Users").Preload("Projects").First(&foundSpace, spaceId), "Space not found"); err != nil {
		return
	}

	userIsInSpace := false
	for _, user := range foundSpace.Users {
		if user.ID == userId {
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
	if _, err := utils.HandleGormError(c, service.DB.Preload("Users").First(&foundSpace, spaceId), "Space not found"); err != nil {
		return
	}

	userId := c.MustGet("userId").(uint)
	userInSpace := false
	for _, user := range foundSpace.Users {
		if user.ID == userId {
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

	var foundUser db.User
	if _, err := utils.HandleGormError(c, service.DB.Preload("Spaces").First(&foundUser, userId), "User not found"); err != nil {
		return
	}

	simpleSpaces := make([]db.SimpleSpace, len(foundUser.Spaces))
	for i, v := range foundUser.Spaces {
		simpleSpaces[i] = v.ToSimpleSpace()
	}

	c.JSON(200, simpleSpaces)
}

func (service *Service) AddUserToSpace(c *gin.Context) {
	spaceId, err := utils.GetRouteParam(c, "spaceId", "Space id is invalid")
	if err != nil {
		return
	}

	userUsername := c.Param("username")

	var foundSpace db.Space
	if _, err := utils.HandleGormError(c, service.DB.Preload("Users").First(&foundSpace, spaceId), "Space not found"); err != nil {
		return
	}

	// Now we have to check whether user is in this space
	userId := c.MustGet("userId").(uint)
	foundUser := false
	userIsInSpace := false
	for _, user := range foundSpace.Users {
		if user.ID == userId {
			foundUser = true
		}
		if user.Username == userUsername {
			userIsInSpace = true
		}
	}

	if !foundUser {
		c.JSON(http.StatusForbidden, gin.H{"message": "Cannot access this space"})
		return
	}

	if userIsInSpace {
		c.JSON(http.StatusNotFound,gin.H{"message":  "User not found"})
		return
	}

	var newUser db.User
	userErr := service.DB.Where("username = ?", userUsername).First(&newUser).Error
	if userErr != nil {
		c.JSON(http.StatusNotFound, "Target user not found")
		return
	}

	if newUser.ID == userId {
		c.JSON(http.StatusConflict, "You are the target user")
		return
	}

	foundSpace.Users = append(foundSpace.Users, newUser)

	service.DB.Save(&foundSpace)

	c.JSON(200, foundSpace.ToSimpleSpace())
}
