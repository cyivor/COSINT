package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// localhost:x/
func RootHandler(apiRoute string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":    "Redirecting",
			"apiRoute": apiRoute,
		})
	}
}

// localhost:x/home.tmpl
func HomeHandler(apiRoute string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.tmpl", gin.H{
			"title":    "Dashboard",
			"apiRoute": apiRoute,
		})
	}
}

// localhost:x/<key>/cosint/identity
func VerifyIdentity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "placeholder"})
}
