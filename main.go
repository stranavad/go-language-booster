package main

import (
	"languageboostergo/db"
	"languageboostergo/export"
	"languageboostergo/languages"
	"languageboostergo/mutations"
	"languageboostergo/projects"
	"languageboostergo/spaces"
	"languageboostergo/types"
	"languageboostergo/users"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)


func main() {
	r := gin.Default()

	//r.Use(cors.Default())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
		AllowHeaders:    []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:   []string{"Content-Length", "Content-Type", "Authorization"},
	}))

	conn := db.GetDb()

	config := types.ServiceConfig {
		DB: conn,
	}

	/* Register modules */
	users.RegisterRoutes(r, config)
	spaces.RegisterRoutes(r, config)
	projects.RegisterRoutes(r, config)
	languages.RegisterRoutes(r, config)
	mutations.RegisterRoutes(r, config)
	export.RegisterRoutes(r, config)

	err := r.Run()

	if err != nil {
		panic("Cannot run application, panic from main.go")
	}
}
