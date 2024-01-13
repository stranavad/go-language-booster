package projects

import (
	"github.com/gin-gonic/gin"
	"languageboostergo/db"
	"net/http"
	"strconv"
)

var conn = db.GetDb()

type CreateProjectDto struct {
	Name string `json:"name" binding:"required"`
}

func CreateProject(c *gin.Context) {
	var request CreateProjectDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newProject := db.Project{
		Name: request.Name,
	}

	conn.Create(&newProject)

	c.JSON(200, newProject.ToSimpleProject())
}

func UpdateProject(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Param("projectId"), 10, 16)
	if err != nil {
		panic("Project ID is not number serializable")
	}

	var request CreateProjectDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updateData db.Project
	updateData.ID = uint(projectId)
	updateData.Name = request.Name
	conn.Save(&updateData)
	c.JSON(200, updateData.ToSimpleProject())
}

func ListProjects(c *gin.Context) {
	var projects []db.Project
	conn.Find(&projects)

	projectsResponse := make([]db.SimpleProject, len(projects))
	for i, v := range projects {
		projectsResponse[i] = v.ToSimpleProject()
	}

	c.JSON(200, projectsResponse)
}
