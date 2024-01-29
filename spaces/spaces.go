package spaces

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"languageboostergo/db"
	"net/http"
	"strconv"
)

var conn = db.GetDb()

type CreateSpaceDto struct {
	Name string `json:"name" binding:"required"`
}

func CreateSpace(c *gin.Context) {
	var request CreateSpaceDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user db.User
	conn.First(&user, c.MustGet("userId").(uint))

	newSpace := db.Space{
		Name: request.Name,
	}

	newSpace.Users = append(newSpace.Users, user)
	conn.Save(&newSpace)

	c.JSON(200, newSpace.ToSimpleSpace())
}

type UpdateSpaceDto struct {
	Name string `json:"name"`
}

func GetById(c *gin.Context) {
	spaceIdParam, err := strconv.ParseUint(c.Param("spaceId"), 10, 32)
	if err != nil {
		panic("Space ID is not number serializable")
	}
	spaceId := uint(spaceIdParam)
	userId := c.MustGet("userId").(uint)

	var foundSpace db.Space
	conn.Preload("Projects").Preload("Users").First(&foundSpace, spaceId)

	userIsInSpace := false
	for _, user := range foundSpace.Users {
		if user.ID == userId {
			userIsInSpace = true
			break
		}
	}

	if !userIsInSpace {
		c.JSON(403, "You are not in this space")
		return
	}

	c.JSON(200, foundSpace.ToSimpleSpace())
}

func UpdateSpace(c *gin.Context) {
	spaceIdParam, err := strconv.ParseUint(c.Param("spaceId"), 10, 32)
	if err != nil {
		panic("Space ID is not number serializable")
	}
	spaceId := uint(spaceIdParam)

	var foundSpace db.Space
	conn.Preload("Users").First(&foundSpace, spaceId)

	userId := c.MustGet("userId").(uint)
	userInSpace := false
	for _, user := range foundSpace.Users {
		if user.ID == userId {
			userInSpace = true
			break
		}
	}

	if !userInSpace {
		c.JSON(403, "You are not in this space, stupid")
		return
	}

	var request CreateSpaceDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn.Model(&foundSpace).Updates(db.Space{Name: request.Name})
	c.JSON(200, foundSpace.ToSimpleSpace())

}

func LeaveSpace(c *gin.Context) {
	spaceIdParam, err := strconv.ParseUint(c.Param("spaceId"), 10, 32)
	if err != nil {
		panic("Space ID is not number serializable")
	}
	spaceId := uint(spaceIdParam)

	var foundSpace db.Space
	conn.First(&foundSpace, spaceId)

	var foundUser db.User
	conn.First(&foundUser, c.MustGet("userId").(uint))

	fmt.Println(foundUser)
	fmt.Println(foundSpace)

	spaceErr := conn.Model(&foundUser).Association("Spaces").Delete([]db.Space{foundSpace})
	if spaceErr != nil {
		return
	}

	c.JSON(200, "Successfully removed you from this space")
}

func ListUserSpaces(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var foundUser db.User
	conn.Preload("Spaces").First(&foundUser, userId)

	simpleSpaces := make([]db.SimpleSpace, len(foundUser.Spaces))
	for i, v := range foundUser.Spaces {
		simpleSpaces[i] = v.ToSimpleSpace()
	}

	c.JSON(200, simpleSpaces)
}

func AddUserToSpace(c *gin.Context) {
	spaceId, err := strconv.ParseUint(c.Param("spaceId"), 10, 32)
	if err != nil {
		panic("Space ID is not number serializable")
	}

	userUsername := c.Param("username")

	var foundSpace db.Space
	conn.Preload("Users").First(&foundSpace, uint(spaceId))

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
		c.JSON(403, "You are not in this space stupid")
		c.Abort()
		return
	}

	if userIsInSpace {
		c.JSON(400, "User is already in this space")
		c.Abort()
		return
	}

	var newUser db.User
	userErr := conn.Where("username = ?", userUsername).First(&newUser).Error
	if userErr != nil {
		c.JSON(400, "Request user doesn't exist")
		c.Abort()
		return
	}

	if newUser.ID == userId {
		c.JSON(400, "You are the user, stupid")
		c.Abort()
		return
	}

	foundSpace.Users = append(foundSpace.Users, newUser)

	conn.Save(&foundSpace)

	c.JSON(200, foundSpace.ToSimpleSpace())
}
