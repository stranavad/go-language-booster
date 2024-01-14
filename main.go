package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"languageboostergo/export"
	"languageboostergo/languages"
	"languageboostergo/mutations"
	"languageboostergo/projects"
	"languageboostergo/spaces"
	"languageboostergo/users"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		userId, err := users.ParseToken(tokenString)
		if err != nil {
			c.JSON(403, "Invalid token")
			c.Abort()
			return
		}

		c.Set("userId", userId)
		c.Next()
	}
}

func main() {
	r := gin.Default()

	//r.Use(cors.Default())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
		AllowHeaders:    []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:   []string{"Content-Length", "Content-Type", "Authorization"},
	}))

	usersGroup := r.Group("/users")
	usersGroup.GET("current", AuthMiddleware(), users.GetCurrent)
	usersGroup.POST("create", users.CreateUser)
	usersGroup.POST("login", users.LoginUser)

	spacesGroup := r.Group("/spaces")
	spacesGroup.Use(AuthMiddleware())
	spacesGroup.PUT(":spaceId", spaces.UpdateSpace)
	spacesGroup.GET("", spaces.ListUserSpaces)
	spacesGroup.GET(":spaceId", spaces.GetById)
	spacesGroup.POST("", spaces.CreateSpace)
	spacesGroup.POST("add-user/:spaceId/:username", spaces.AddUserToSpace)
	spacesGroup.POST("leave/:spaceId", spaces.LeaveSpace)

	projectsGroup := r.Group("/projects")
	projectsGroup.Use(AuthMiddleware())
	projectsGroup.GET("by-id/:projectId", projects.GetById)
	projectsGroup.GET(":spaceId", projects.ListProjects)
	projectsGroup.POST("", projects.CreateProject)
	projectsGroup.PUT(":projectId", projects.UpdateProject)

	languagesGroup := r.Group("/languages")
	languagesGroup.Use(AuthMiddleware())
	languagesGroup.GET(":projectId", languages.GetLanguagesByProjectId)
	languagesGroup.POST("", languages.CreateLanguage)
	languagesGroup.PUT(":languageId", languages.UpdateLanguage)

	mutationsGroup := r.Group("/mutations")
	mutationsGroup.Use(AuthMiddleware())
	mutationsGroup.GET(":mutationId", mutations.GetById)
	mutationsGroup.GET("/project/:projectId", mutations.ListByProject)
	mutationsGroup.POST("", mutations.CreateMutation)
	mutationsGroup.PUT(":mutationId", mutations.UpdateMutation)
	mutationsGroup.DELETE(":mutationId", mutations.DeleteById)
	mutationsGroup.POST("/value", mutations.CreateMutationValue)
	mutationsGroup.PUT("/value/:mutationValueId", mutations.UpdateMutationValue)

	exportsGroup := r.Group("/export")
	exportsGroup.Use(AuthMiddleware())
	exportsGroup.POST("", export.ByProjectIdAndLanguageId)
	r.Run()
}
