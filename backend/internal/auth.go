package internal

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const AdminField = "admin"

func SetAdmin(c *gin.Context) error {
	session := sessions.Default(c)
	session.Set("admin", true)
	return session.Save()
}

func CheckAdmin(c *gin.Context) bool {
	session := sessions.Default(c)
	return session.Get("admin") == true
}
