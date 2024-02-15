package internal

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const AdminSessionsName = "admin"

func BuildRouter(app *App) *gin.Engine {
	router := gin.Default()

	store := cookie.NewStore([]byte(app.Config.CookieSecret))

	router.LoadHTMLGlob(app.Config.TemplateGLOB)
	router.Use(gin.Logger())
	router.Use(sessions.Sessions("admin", store))
	router.Use(gin.Recovery())

	router.GET("/", app.Home)
	router.POST("/search", app.Search)
	router.GET("/signin", app.SignIn)
	router.POST("/submit_signin", app.SubmitSignIn)

	adminRouter := router.Group(AdminSessionsName)
	adminRouter.Use(AdminAuth)
	adminRouter.GET("/", app.Admin)

	return router
}

func AdminAuth(c *gin.Context) {
	if !CheckAdmin(c) {
		c.Redirect(http.StatusSeeOther, "/signin")
		c.Abort()
	}
	c.Next()
}
