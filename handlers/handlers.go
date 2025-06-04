package handlers

import (
	"fmt"
	"net/http"
	"time"

	"cyivor/cosint/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
func HomeHandler(capir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.tmpl", gin.H{
			"title": "Dashboard",
			"capir": capir,
		})
	}
}

// localhost:x/<key>/cosint/identity
func VerifyIdentity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "placeholder"})
}

// localhost:x/<key>/cosint/ext-apis/snusbase GET
func SnusHandler(extapir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "snusbase.tmpl", gin.H{
			"title":   "Snusbase search",
			"extapir": extapir,
		})
	}
}

// localhost:x/<key>/cosint/ext-apis/snusbase POST
func SnusResults(capir string, snusKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if snusKey == "" {
			c.HTML(http.StatusUnauthorized, "enverror.tmpl", gin.H{
				"title": "Environment Variable Error",
				"key":   "SNUSBASE_KEY",
			})
			time.Sleep(5 * time.Second)
			c.Redirect(http.StatusFound, capir)
			return
		}

		searchTerm := c.PostForm("search")
		field := c.PostForm("field")

		logger := c.MustGet("logger").(*zap.Logger)

		if searchTerm == "" || field == "" {
			logger.Error("Invalid form data", zap.String("search", searchTerm), zap.String("field", field))
			c.JSON(http.StatusBadRequest, gin.H{"error": "search and field parameters are required"})
			return
		}

		searchBody := types.RequestBody{
			Terms: []string{searchTerm},
			Types: []string{field},
		}

		response, err := sendRequest("data/search", snusKey, searchBody)
		if err != nil {
			logger.Error("Failed to send Snusbase API request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Snusbase API request failed: %v", err)})
			return
		}

		/*
			logger.Info("Snusbase API request successful",
				zap.String("search", searchTerm),
				zap.String("field", field),
				zap.Any("response", response))
		*/

		c.JSON(http.StatusOK, response)
	}
}
