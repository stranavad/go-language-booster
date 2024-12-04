package projects

import (
	"languageboostergo/auth"
	"languageboostergo/db"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var conn = db.GetDb()

type CreateProjectDto struct {
	Name    string `json:"name" binding:"required"`
	SpaceId uint   `json:"spaceId"`
}

func CreateProject(c *gin.Context) {
	var request CreateProjectDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var foundSpace db.Space
	conn.Preload("Users").First(&foundSpace, request.SpaceId)
	userId := c.MustGet("userId").(uint)
	userInSpace := false
	for _, user := range foundSpace.Users {
		if user.ID == userId {
			userInSpace = true
			break
		}
	}

	if !userInSpace {
		c.JSON(403, "You are not in this space stupid")
		return
	}

	newProject := db.Project{
		Name:    request.Name,
		SpaceID: foundSpace.ID,
	}

	conn.Create(&newProject)

	c.JSON(200, newProject.ToSimpleProject())
}

func GetById(c *gin.Context) {
	projectIdParam, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Project id is invalid"})
		return
	}

	userId := c.MustGet("userId").(uint)
	projectId := uint(projectIdParam)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(403, "You cannot read this project")
		return
	}

	var foundProject db.Project
	conn.First(&foundProject, projectId)
	c.JSON(200, foundProject.ToSimpleProject())
}

func UpdateProject(c *gin.Context) {
	projectIdParam, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Project id is invalid"})
		return
	}

	var request CreateProjectDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("userId").(uint)
	projectId := uint(projectIdParam)

	if !auth.IsUserInProject(userId, projectId) {
		c.JSON(403, "You cannot update this project")
		return
	}

	var updateData db.Project
	conn.First(&updateData, projectId)

	if request.Name != "" {
		updateData.Name = request.Name
	}

	conn.Save(&updateData)
	c.JSON(200, updateData.ToSimpleProject())
}

func ListProjects(c *gin.Context) {
	spaceId, err := strconv.ParseUint(c.Param("spaceId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Space id is invalid"})
		return
	}
	
	var foundSpace db.Space
	conn.Preload("Users").Preload("Projects").First(&foundSpace, uint(spaceId))

	// Check user relevance
	userId := c.MustGet("userId").(uint)
	userInSpace := false
	for _, user := range foundSpace.Users {
		if user.ID == userId {
			userInSpace = true
			break
		}
	}

	if !userInSpace {
		c.JSON(403, "Cannot access this space")
		return
	}

	c.JSON(200, foundSpace.ToSimpleSpace().Projects)
}
