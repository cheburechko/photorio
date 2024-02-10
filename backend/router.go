package main

import "github.com/gin-gonic/gin"

func BuildRouter(app *App) *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob(app.Config.TemplateGLOB)
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", app.Home)
	router.POST("/search", app.Search)
	return router
}
