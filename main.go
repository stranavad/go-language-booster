package main

import (
	"github.com/gin-gonic/gin"
	"languageboostergo/languages"
	"languageboostergo/projects"
)

func main() {
	r := gin.Default()

	projectsGroup := r.Group("/projects")
	projectsGroup.GET("", projects.ListProjects)
	projectsGroup.POST("", projects.CreateProject)
	projectsGroup.PUT(":projectId", projects.UpdateProject)

	languagesGroup := r.Group("/languages")
	languagesGroup.GET(":projectId", languages.GetLanguagesByProjectId)
	languagesGroup.POST("", languages.CreateLanguage)
	languagesGroup.PUT(":languageId", languages.UpdateLanguage)
	r.Run()
}
