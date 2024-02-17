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
	router.Use(sessions.Sessions(AdminSessionsName, store))
	router.Use(gin.Recovery())

	router.GET("/", app.Home)
	router.POST("search", app.Search)
	router.GET("signin", app.SignIn)
	router.POST("submit_signin", app.SubmitSignIn)

	adminRouter := router.Group("/admin")
	adminRouter.Use(AdminAuth)
	adminRouter.GET("/", app.Admin)
	adminRouter.POST("submit_caption_task", app.SubmitCaptionTask)
	adminRouter.GET("caption_tasks", app.CaptionTasks)

	apiRouter := router.Group("/api")
	apiRouter.Use(AdminAuth)
	apiRouter.POST("register", app.CreateUser)

	return router
}

func AdminAuth(c *gin.Context) {
	if !CheckAdmin(c) {
		c.Redirect(http.StatusSeeOther, "/signin")
		c.Abort()
	}
	c.Next()
}
