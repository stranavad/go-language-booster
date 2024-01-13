package main

import (
	"github.com/gin-gonic/gin"
	"languageboostergo/export"
	"languageboostergo/languages"
	"languageboostergo/mutations"
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

	mutationsGroup := r.Group("/mutations")
	mutationsGroup.GET("/project/:projectId", mutations.ListByProject)
	mutationsGroup.POST("", mutations.CreateMutation)
	mutationsGroup.PUT(":mutationId", mutations.UpdateMutation)
	mutationsGroup.PUT("/value/:mutationValueId", mutations.UpdateMutationValue)

	exportsGroup := r.Group("/export")
	exportsGroup.POST("", export.ByProjectIdAndLanguageId)
	r.Run()
}
